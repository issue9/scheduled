// SPDX-License-Identifier: MIT

package scheduled

import (
	"context"
	"sync"
	"time"
)

// Server 管理所有的定时任务
type Server struct {
	jobs           []*Job
	nextScheduled  chan struct{} // 需要指行下一次调度任务
	exitSchedule   chan struct{} // 需要退出 Server.schedule 方法
	scheduleLocker sync.Mutex

	loc        *time.Location
	running    bool
	erro, info Logger
}

// NewServer 声明 Server 对象实例
//
// loc 指定当前所采用的时区，若为 nil，则会采用 [time.Local] 的值；
// erro 计划任务发生的错误，向此输出，可以为空，表示不输出；
// info 计划任务的执行信息，向此输出，可以为空，表示不输出；
func NewServer(loc *time.Location, erro, info Logger) *Server {
	if loc == nil {
		loc = time.Local
	}

	return &Server{
		jobs:          make([]*Job, 0, 100),
		nextScheduled: make(chan struct{}, 1),
		exitSchedule:  make(chan struct{}, 1),

		loc:  loc,
		erro: erro,
		info: info,
	}
}

// Location 返回当前任务相关联的时区信息
func (s *Server) Location() *time.Location { return s.loc }

// Serve 运行服务
func (s *Server) Serve(ctx context.Context) error {
	s.running = true
	defer func() {
		s.running = false
	}()

	// 初始化任务
	now := time.Now()
	for _, job := range s.jobs {
		job.init(now)
	}

	s.sendNextScheduled()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.nextScheduled:
			s.schedule(ctx)
		}
	}
}

func (s *Server) sendNextScheduled() {
	if len(s.nextScheduled) == 0 {
		s.nextScheduled <- struct{}{}
	}
}

// 调度计划任务
//
// 每完成一个计划任务时，都会调用此函数重新计算调度时间，
// 并重新生成一个最近时间的定时器。如果上一个定时器还未结束，
// 则会自动结束上一个定时器，schedule 会保证同一时间，
// 只有一个函数实例在运行。
func (s *Server) schedule(ctx context.Context) {
	s.scheduleLocker.Lock()
	defer s.scheduleLocker.Unlock()

	sortJobs(s.jobs) // 按执行时间进行排序

	var dur time.Duration

	if len(s.jobs) == 0 || s.jobs[0].next.IsZero() {
		dur = time.Minute
	} else {
		dur = time.Until(s.jobs[0].next)
	}

	// dur > 0 表示没有需要立即执行的，根据最早的一条任务做一个计时器。
	if dur > 0 {
		timer := time.NewTimer(dur)
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				s.sendNextScheduled()
				return
			case <-s.exitSchedule:
				return
			}
		}
	}

	now := time.Now()
	for _, j := range s.jobs {
		if j.next.After(now) || j.next.IsZero() {
			break
		}

		if j.State() == Running && j.Delay() {
			continue
		}

		j.calcState() // 计算关键信息
		go j.run(now, s.erro, s.info)
	}

	s.sendNextScheduled()
}
