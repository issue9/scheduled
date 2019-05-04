// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package cron 定时任务
package cron

import "context"

// Cron 管理所有的定时任务
type Cron struct {
	jobs     []*Job
	channels chan *Job
}

// New 声明 Cron 对象实例
func New() *Cron {
	return &Cron{
		jobs: make([]*Job, 0, 100),
	}
}

// Serve 运行服务
func (c *Cron) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case j := <-c.channels:
			go func(j *Job) {
				if err := j.f(); err != nil {
					// TODO
				}

				now := j.next.Next(j.last)
				j.last = now
				// TODO
			}(j)
		}
	}
}
