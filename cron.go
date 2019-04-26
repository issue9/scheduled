// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package cron 定时任务
package cron

import (
	"context"
	"time"
)

// 表示任务状态
const (
	Stoped State = iota
	Running
	Failed
)

// State 状态值类型
type State int8

// Cron 管理所有的定时任务
type Cron struct {
	jobs     []*Job
	channels chan *Job
}

// JobFunc 每一个定时任务实际上执行的函数签名
type JobFunc func() error

// Job 一个定时任务的基本接口
type Job struct {
	name  string
	f     JobFunc
	last  time.Time
	state State
	next  Nexter
}

// New 声明 Cron 对象实例
func New() *Cron {
	return &Cron{
		jobs: make([]*Job, 0, 100),
	}
}

// New 添加一个新的定时任务
func (c *Cron) New(name string, f JobFunc, n Nexter) {
	c.jobs = append(c.jobs, &Job{
		name:  name,
		f:     f,
		next:  n,
		state: Stoped,
	})
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
