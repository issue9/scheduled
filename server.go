// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package scheduled

import (
	"log"
	"sync"
	"time"
)

// Server 管理所有的定时任务
type Server struct {
	// 任务列表
	// 调度算法会保证按执行时间顺序进行排序。
	jobs []*Job

	// 需要立马运行的任务
	nextJob chan *Job

	timer          *time.Timer
	scheduleLocker sync.Mutex
	stop           chan struct{}
	loc            *time.Location
	running        bool
}

// NewServer 声明 Server 对象实例
//
// loc 指定当前所采用的时区，若为 nil，则会采用 time.Local 的值。
func NewServer(loc *time.Location) *Server {
	if loc == nil {
		loc = time.Local
	}

	return &Server{
		jobs:    make([]*Job, 0, 100),
		stop:    make(chan struct{}, 1),
		nextJob: make(chan *Job, 10),
		loc:     loc,
	}
}

// Location 返回当前任务相关联的时区信息
func (s *Server) Location() *time.Location {
	return s.loc
}

// Serve 运行服务
//
// errlog 定时任务的错误信息在此通道输出，若为空，则不输出；
// infolog 如果不为空，则会输出一些额外的提示信息，方便调试。
func (s *Server) Serve(errlog, infolog *log.Logger) error {
	if s.running {
		return ErrRunning
	}
	s.running = true

	if len(s.jobs) == 0 {
		s.running = false
		return ErrNoJobs
	}

	now := s.now()
	for _, job := range s.jobs {
		job.init(now)
	}

	s.schedule()

	for {
		select {
		case <-s.stop:
			return nil
		case j := <-s.nextJob:
			go func() {
				j.run(errlog, infolog)
				s.schedule()
			}()
		}
	}
}

// 调度计划任务
//
// 每完成一个计划任务时，都会调用此函数重新计算调度时间，
// 并重新生成一个最近时间的定时器。如果上一个定时器还未结束，
// 则会自动结束上一个定时器，schedule 会保证同一时间，
// 只有一个函数实例在运行。
func (s *Server) schedule() {
	if s.timer != nil {
		s.timer.Stop() // 多次调用或是已过期，都不会 panic
	}

	s.scheduleLocker.Lock()
	defer s.scheduleLocker.Unlock()

	sortJobs(s.jobs)

	// 下一次计划任务执行的时间
	next := s.jobs[0].next

	if next.IsZero() { // 没有需要运行的任务
		s.running = false
		s.Stop()
		return
	}

	dur := next.Sub(s.now())
	if dur < 0 {
		dur = 0
	}

	s.timer = time.NewTimer(dur)

	// s.timer 可能造成 schedule 函数的长时间等待
	n, ok := <-s.timer.C
	if !ok {
		return
	}

	for _, j := range s.jobs {
		if j.State() == Running { // 上一次任务还没结束，则跳过该任务
			continue
		}

		if j.next.IsZero() || j.next.After(n) {
			return
		}

		j.at = n
		s.nextJob <- j
	}
}

// Stop 停止当前服务
func (s *Server) Stop() {
	if !s.running {
		return
	}

	s.running = false

	if s.timer != nil {
		s.timer.Stop()
	}

	// NOTE: 不能通过关闭 nextJob 来结束 Server。
	// 因为 schedule() 是异步执行的，
	// 会源源不断地推内容到 nextJob，如果关闭，可能会造成 schedule() panic
	s.stop <- struct{}{}
}

func (s *Server) now() time.Time {
	return time.Now().In(s.Location())
}
