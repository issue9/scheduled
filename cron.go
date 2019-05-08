// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package scheduled 定时任务
package scheduled

import (
	"errors"
	"log"
	"sort"
	"time"
)

// 一些错误的定义
var (
	ErrNoJobs  = errors.New("任务列表为空")
	ErrRunning = errors.New("任务已经在运行")
)

// Server 管理所有的定时任务
type Server struct {
	jobs    []*Job
	stop    chan struct{}
	running bool
}

// NewServer 声明 Server 对象实例
func NewServer() *Server {
	return &Server{
		jobs: make([]*Job, 0, 100),
		stop: make(chan struct{}, 1),
	}
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
		return ErrNoJobs
	}

	now := time.Now()
	for _, job := range s.jobs {
		job.init(now)
	}

	for {
		sortJobs(s.jobs)

		if s.jobs[0].next.IsZero() { // 没有需要运行的任务
			return ErrNoJobs
		}

		dur := s.jobs[0].next.Sub(time.Now())
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
				if j.next.IsZero() || j.next.After(n) {
					break
				}
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

func sortJobs(jobs []*Job) {
	sort.SliceStable(jobs, func(i, j int) bool {
		if jobs[i].next.IsZero() {
			return false
		}
		if jobs[j].next.IsZero() {
			return true
		}
		return jobs[i].next.Before(jobs[j].next)
	})
}
