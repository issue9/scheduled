// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package cron 定时任务
package cron

import (
	"sort"
	"time"
)

// Cron 管理所有的定时任务
type Cron struct {
	jobs []*Job
	stop chan struct{}
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
	now := time.Now()
	for _, job := range c.jobs {
		job.init(now)
	}

	for {
		sortJobs(c.jobs)

		if len(c.jobs) == 0 {
			time.Sleep(24 * time.Hour) // 没有内容
		}

		timer := time.NewTicker(c.jobs[0].next.Sub(now))
		for {
			select {
			case <-c.stop:
				timer.Stop()
				return nil
			case n := <-timer.C:
				go c.jobs[0].run(n)
			}
		}
	}
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
