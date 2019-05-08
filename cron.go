// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package cron 定时任务
package cron

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
//
// errlog 定时任务的错误信息在此通道输出，若为空，则不输出。
func (c *Cron) Serve(errlog *log.Logger) error {
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

		if c.jobs[0].next.IsZero() { // 没有需要运行的任务
			time.Sleep(24 * time.Hour)
		}

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
					go j.run(n, errlog)
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
			return false
		}
		if jobs[j].next.IsZero() {
			return true
		}
		return jobs[i].next.Before(jobs[j].next)
	})
}
