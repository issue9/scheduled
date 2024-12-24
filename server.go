// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package scheduled

import (
	"context"
	"time"
)

// Server 管理所有的定时任务
type Server struct {
	jobs     []*Job
	works    chan *Job     // 等待运行的任务
	schedule chan struct{} // 重新执行调度任务

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
		works:    make(chan *Job, 100),
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
	s.running = true
	defer func() {
		s.running = false
	}()

	// 初始化任务
	now := time.Now()
	for _, job := range s.jobs {
		job.init(now)
	}

LOOP:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case j := <-s.works:
			go j.run(now, s.erro, s.info)
		default: // 没有在执行的任务了，则计算一次时间
			if len(s.jobs) == 0 {
				<-time.After(time.Second) // 没有任务，1 秒后再次执行调试
				continue LOOP
			}

			now = time.Now()

			sortJobs(s.jobs)
			dur := s.jobs[0].Next().Sub(now)
			if dur > 0 { //
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
				if !j.Next().IsZero() && ((j.Next().Before(now) || j.Next().Equal(now)) &&
					(!j.Delay() || j.State() != Running)) {
					j.calcState(now) // 先计算状态，再异步运行。
					s.works <- j
				}
			}
		}
	}
}
