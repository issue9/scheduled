// SPDX-License-Identifier: MIT

package scheduled

import (
	"log"
	"sync"
	"time"
)

// Server 管理所有的定时任务
type Server struct {
	jobs           []*Job
	nextScheduled  chan struct{} // 需要指行下一次调度任务
	exitTimer      chan struct{}
	scheduleLocker sync.Mutex
	stop           chan struct{}

	loc             *time.Location
	running         bool
	errlog, infolog *log.Logger
}

// NewServer 声明 Server 对象实例
//
// loc 指定当前所采用的时区，若为 nil，则会采用 time.Local 的值；
// errlog 定时任务的错误信息在此通道输出，若为空，则不输出；
// infolog 如果不为空，则会输出一些额外的提示信息，方便调试。
func NewServer(loc *time.Location, errlog, infolog *log.Logger) *Server {
	if loc == nil {
		loc = time.Local
	}

	return &Server{
		jobs:          make([]*Job, 0, 100),
		nextScheduled: make(chan struct{}, 1),
		exitTimer:     make(chan struct{}, 1),
		stop:          make(chan struct{}, 1),

		loc:     loc,
		errlog:  errlog,
		infolog: infolog,
	}
}

// Location 返回当前任务相关联的时区信息
func (s *Server) Location() *time.Location { return s.loc }

// Serve 运行服务
func (s *Server) Serve() error {
	if s.running {
		return ErrRunning
	}

	s.running = true

	// 初始化任务
	now := time.Now()
	for _, job := range s.jobs {
		job.init(now)
	}

	s.sendNextScheduled()

	for {
		select {
		case <-s.stop:
			return nil
		case <-s.nextScheduled:
			s.schedule()
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
func (s *Server) schedule() {
	s.scheduleLocker.Lock()
	defer s.scheduleLocker.Unlock()

	sortJobs(s.jobs) // 按执行时间进行排序

	var dur time.Duration

	if len(s.jobs) == 0 || s.jobs[0].next.IsZero() {
		dur = time.Minute
	} else {
		dur = time.Until(s.jobs[0].next)
	}

	// dur >0 表示没有需要立即执行的，根据最早的一条任务做一个计时器。
	if dur > 0 {
		timer := time.NewTimer(dur)
		for {
			select {
			case <-timer.C:
				s.sendNextScheduled()
				return
			case <-s.exitTimer:
				s.sendNextScheduled()
				return
			}
		}
	}

	now := time.Now()
	for _, j := range s.jobs {
		if j.next.After(now) {
			break
		}

		if j.State() == Running && j.Delay() { // 上一次任务还没结束，且是 delay 模式，则跳过此次任务
			continue
		}

		j.calcState() // 计算关键信息
		go j.run(now, s.errlog, s.infolog)
	}

	s.sendNextScheduled()
}

// Stop 停止当前服务
func (s *Server) Stop() {
	if !s.running {
		return
	}

	s.running = false

	// NOTE: 不能通过关闭 nextScheduled 来结束 Server。
	// 因为 schedule() 是异步执行的，
	// 会源源不断地推内容到 nextJob，如果关闭，可能会造成 schedule() panic
	s.stop <- struct{}{}
}
