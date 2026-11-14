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
	schedule   chan struct{} // 重新执行调度任务
	loc        *time.Location
	erro, info Logger

	wait sync.WaitGroup

	// 以下内容需要上锁

	locker sync.RWMutex
	jobs   []*job
	state  State
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
		jobs:     make([]*job, 0, 100),
		schedule: make(chan struct{}, 1),

		loc:  loc,
		erro: erro,
		info: info,
	}
}

// Location 返回当前任务相关联的时区信息
func (s *Server) Location() *time.Location { return s.loc }

// State 当前服务的状态
//
// 当前对象不会返回 [Failed] 状态，该状态只对 [Job] 有效。
func (s *Server) State() State {
	s.locker.RLock()
	state := s.state
	s.locker.RUnlock()
	return state
}

// Serve 运行服务
//
// ctx 用于控制服务的退出，不能为空。
// 并不是 ctx.Done 就立即返回，还需要等待所有的 Job 完成当次任务。
func (s *Server) Serve(ctx context.Context) error {
	if ctx == nil {
		panic("参数 ctx 不能为空")
	}

	now := time.Now().In(s.Location())
	s.locker.Lock()
	if state := s.state; state != Stopped {
		s.locker.Unlock()
		return errServerState(state)
	}

	s.state = Running
	for _, job := range s.jobs {
		job.init(now)
	}
	s.locker.Unlock()

	defer func() {
		// 设置 jobs 的状态，并清空任务列表
		s.locker.Lock()
		for _, j := range s.jobs {
			j.locker.Lock()
			j.state = Cancelled
			j.next = time.Time{}
			j.locker.Unlock()
		}
		s.jobs = nil        // 直接释放，反正 Server 不具备重启条件
		s.state = Cancelled // 只能通过 ctx.Done 退出，所以肯定是 Cancelled
		s.locker.Unlock()

		s.wait.Wait() // 等待所有任务结束
	}()

LOOP:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default: // 没有在执行的任务了，则计算一次时间
			var job1 *job

			s.locker.Lock()
			if len(s.jobs) == 0 {
				s.locker.Unlock()

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-s.schedule:
				case <-time.After(time.Second): // 没有任务，1 秒后再次执行调度
				}
				continue LOOP
			} else {
				sortJobs(s.jobs)
				job1 = s.jobs[0]
				s.locker.Unlock()
			}

			job1.locker.RLock()
			j1next := job1.next
			j1state := job1.state
			job1.locker.RUnlock()

			// 队列中有任务，但是第一个任务的调度时间 0 或是状态不对，那么只能是以下两种情况：
			//  - 表示所有的任务都不需要执行，创建一个 1 秒的定时器，等待下一轮检测；
			//  - job1 的任务刚结束了一轮执行，刷新了 next 或是正在执行中改变了状态，等待 s.schedule 的到来；
			if j1next.IsZero() || j1state == Running || j1state == Cancelled {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-s.schedule:
				case <-time.After(time.Second): // 没有任务，1 秒后再次执行调度
				}
				continue LOOP
			}

			// 队列中有任务，但是还没有到时间，创建定时器。

			now = time.Now().In(s.Location())
			dur := j1next.Sub(now)
			if dur > 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(dur):
				case <-s.schedule:
				}
				continue LOOP
			}

			// 队列中有任务，且已经到执行时间了，依次调度任务。

			s.locker.RLock()
			for _, j := range s.jobs {
				j.locker.Lock()
				state := j.state
				next := j.next

				// 可能会存在通过 Server.New 返回的函数主动取消单个 Job 的方式，所以还要考虑 j.State 的值。
				if state == Cancelled || state == Running || next.IsZero() {
					j.locker.Unlock()
					continue
				}

				if dur := next.Sub(now); dur <= 0 {
					j.state = Running
					j.locker.Unlock()

					n := now // 保证下面的异步能得到正确的 now 值
					s.wait.Go(func() { j.run(n) })
				} else {
					j.locker.Unlock()

					s.locker.RUnlock()

					// 这是个有序的队列，只要第一个的时间不符合要求，后续的肯定不符合。
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(dur):
					case <-s.schedule:
					}
					continue LOOP // 当前分支已经调用 s.locker.RUnlock，所以直接跳出到 LOOP 循环。
				}
			}
			s.locker.RUnlock()
		}
	}
}
