// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package cron 定时任务
package cron

import (
	"errors"
	"sort"
	"time"
)

// 一些错误的定义
var (
	ErrNoJobs  = errors.New("任务列表为空")
	ErrRunning = errors.New("任务已经在运行")
)

// Cron 管理所有的定时任务
type Cron struct {
	jobs    []*Job
	stop    chan struct{}
	running bool
}

// New 声明 Cron 对象实例
func New() *Cron {
	return &Cron{
		jobs: make([]*Job, 0, 100),
		stop: make(chan struct{}, 1),
	}
}

// Serve 运行服务
func (c *Cron) Serve() error {
	if c.running {
		return ErrRunning
	}

	c.running = true

	if len(c.jobs) == 0 {
		return ErrNoJobs
	}

	now := time.Now()
	for _, job := range c.jobs {
		job.init(now)
	}

	for {
		sortJobs(c.jobs)

		timer := time.NewTicker(c.jobs[0].next.Sub(now))
		for {
			select {
			case <-c.stop:
				timer.Stop()
				return nil
			case n := <-timer.C:
				for _, j := range c.jobs {
					if j.next.IsZero() || j.next.After(n) {
						break
					}
					go c.jobs[0].run(n)
				}
			} // end select
		}
	}
}

// Stop 停止当前服务
func (c *Cron) Stop() {
	if !c.running {
		return
	}

	c.stop <- struct{}{}
}

func sortJobs(jobs []*Job) {
	sort.SliceStable(jobs, func(i, j int) bool {
		if jobs[i].next.IsZero() {
			return true
		}
		if jobs[j].next.IsZero() {
			return false
		}
		return jobs[i].next.Before(jobs[j].next)
	})
}
