// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package scheduled

import (
	"context"
	"time"
)

// Server 管理所有的定时任务
type Server struct {
	jobs      []*Job
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
		for _, j := range s.jobs {
			j.locker.Lock()
			j.state = Cancelled
			j.next = time.Time{}
			j.locker.Unlock()
		}
		s.jobs = nil // 直接释放，反正 Server 不具备重启条件
	}()

	// 初始化任务
	now := time.Now().In(s.Location())
	for _, job := range s.jobs {
		job.init(now)
	}

LOOP:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default: // 没有在执行的任务了，则计算一次时间
			if len(s.jobs) == 0 {
				<-time.After(time.Second) // 没有任务，1 秒后再次执行调试
				continue LOOP
			}

			now = time.Now().In(s.Location())

			sortJobs(s.jobs)
			dur := s.jobs[0].Next().Sub(now)
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
		}
	}
}
