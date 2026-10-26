// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package scheduled

import (
	"context"
	"sync"
	"time"
)

// Server 管理所有的定时任务
type Server struct {
	jobsMu sync.RWMutex
	jobs   []*Job
	wait   sync.WaitGroup

	schedule  chan struct{} // 重新执行调度任务
	cancelled bool          // 已经接受到 context.Context.Done() 的信号，表示服务已经被取消

	loc        *time.Location
	running    bool
	erro, info Logger
}

// NewServer 声明 [Server]
//
// loc 指定当前所采用的时区，若为 nil，则会采用 [time.Local] 的值；
// erro 计划任务发生的错误，向此输出，可以为空，表示不输出；
// info 计划任务的执行信息，向此输出，可以为空，表示不输出；
func NewServer(loc *time.Location, erro, info Logger) *Server {
	if loc == nil {
		loc = time.Local
	}

	if erro == nil {
		erro = &defaultLogger{}
	}

	if info == nil {
		info = &defaultLogger{}
	}

	return &Server{
		jobs:     make([]*Job, 0, 100),
		schedule: make(chan struct{}, 1),

		loc:  loc,
		erro: erro,
		info: info,
	}
}

// Location 返回当前任务相关联的时区信息
func (s *Server) Location() *time.Location { return s.loc }

// Serve 运行服务
//
// 并不是 ctx.Done 就立即返回，还需要等待所有的 Job 完成当次任务。
func (s *Server) Serve(ctx context.Context) error {
	if s.cancelled {
		return ErrServerCancelled()
	}
	if s.running {
		return ErrServerRunning()
	}

	s.running = true
	defer func() {
		s.running = false
		s.cancelled = true

		// 设置 jobs 的状态，并清空任务列表
		s.jobsMu.Lock()
		for _, j := range s.jobs {
			j.locker.Lock()
			j.state = Cancelled
			j.next = time.Time{}
			j.locker.Unlock()
		}
		s.jobs = nil // 直接释放，反正 Server 不具备重启条件
		s.jobsMu.Unlock()

		s.wait.Wait() // 等待所有任务结束
	}()

	// 初始化任务
	now := time.Now().In(s.Location())
	s.jobsMu.RLock()
	for _, job := range s.jobs {
		job.init(now)
	}
	s.jobsMu.RUnlock()

LOOP:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default: // 没有在执行的任务了，则计算一次时间
			var job1 *Job
			s.jobsMu.Lock()
			if len(s.jobs) == 0 {
				s.jobsMu.Unlock()
				<-time.After(time.Second) // 没有任务，1 秒后再次执行调试
				continue LOOP
			} else {
				sortJobs(s.jobs)
				job1 = s.jobs[0]
				s.jobsMu.Unlock()
			}

			// 队列中有任务，但是第一个任务的调度时间 0，表示所有的任务都不需要执行。
			// 那么就创建一个 1 秒的定时器，看看一秒之后有没有新的任务加入且需要执行的。
			if job1.next.IsZero() {
				<-time.After(time.Second)
				continue LOOP
			}

			// 队列中有任务，但是还没有到时间，创建定时器。

			now = time.Now().In(s.Location())
			dur := job1.Next().Sub(now)
			if dur > 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(dur):
					continue LOOP
				case <-s.schedule:
					continue LOOP
				}
			}

			// 队列中有任务，且已经到执行时间了，依次调度任务。

			s.jobsMu.RLock()
			for _, j := range s.jobs {
				if j.State() == Cancelled {
					continue
				}

				if !j.Next().IsZero() && ((j.Next().Before(now) || j.Next().Equal(now)) &&
					(!j.Delay() || (j.State() != Running && j.State() != Cancelled))) {
					j.calcState(now) // 先计算状态，再异步运行。
					go j.run(now)
				}
			}
			s.jobsMu.RUnlock()
		}
	}
}
