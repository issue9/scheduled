// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package scheduled

import (
	"log"
	"time"
)

// Server 管理所有的定时任务
type Server struct {
	jobs    []*Job
	stop    chan struct{}
	loc     *time.Location
	running bool
}

// NewServer 声明 Server 对象实例
//
// loc 指定当前所采用的时区，若为 nil，则会采用 time.Local 的值。
func NewServer(loc *time.Location) *Server {
	if loc == nil {
		loc = time.Local
	}

	return &Server{
		jobs: make([]*Job, 0, 100),
		stop: make(chan struct{}, 1),
		loc:  loc,
	}
}

// Location 返回当前任务相关联的时区信息
func (s *Server) Location() *time.Location {
	return s.loc
}

// Serve 运行服务
//
// errlog 定时任务的错误信息在此通道输出，若为空，则不输出。
func (s *Server) Serve(errlog *log.Logger) error {
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

	for {
		sortJobs(s.jobs)

		if s.jobs[0].next.IsZero() { // 没有需要运行的任务
			s.running = false
			return ErrNoJobs
		}

		now = s.now()
		dur := s.jobs[0].next.Sub(now)
		if dur < 0 {
			dur = 0
		}
		timer := time.NewTimer(dur)

		select {
		case <-s.stop:
			timer.Stop()
			return nil
		case n := <-timer.C:
			for _, j := range s.jobs {
				if j.next.IsZero() || j.next.After(n) || j.State() == Running {
					break
				}

				// 确保在状态变为 running 时，才执行 go 协程，以防止在 run
				// 中还未改变状态，已经开始新一轮的 for 循环。
				j.state = Running
				go j.run(n, errlog)
			}
		} // end select
	}
}

// Stop 停止当前服务
func (s *Server) Stop() {
	if !s.running {
		return
	}

	s.running = false
	s.stop <- struct{}{}
}

func (s *Server) now() time.Time {
	return time.Now().In(s.Location())
}
