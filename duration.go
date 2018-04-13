// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import "time"

// Duration 以固定时间执行的定时任务
type Duration struct {
	name     string
	state    State
	duration time.Duration
	err      error // 错误信息

	task   func() error
	ticker *time.Ticker
	stop   chan struct{}
}

// NewJobWithDuration 声明一个新的定时任务
func NewJobWithDuration(name string, task func() error, dur time.Duration) Job {
	return &Duration{
		name:     name,
		state:    Stoped,
		duration: dur,
		task:     task,
		stop:     make(chan struct{}, 1),
	}
}

// Name 任务名称
func (job *Duration) Name() string {
	return job.name
}

// State 状态，若状态值为 Failed，则可以使用 Err 函数获取具体的错误信息。
func (job *Duration) State() State {
	return job.state
}

// Err 错误信息
func (job *Duration) Err() error {
	return job.err
}

// Start 启动当前的定时任务
func (job *Duration) Start() error {
	job.state = Running
	job.err = nil
	job.ticker = time.NewTicker(job.duration)

	go func() {
		for {
			select {
			case <-job.ticker.C:
				if err := job.task(); err != nil {
					job.state |= Failed
					job.err = err
				} else {
					job.state = Running
					job.err = nil
				}
			case <-job.stop:
				job.ticker.Stop()
				return
			}
		}
	}()

	return nil
}

// Stop 停止执行
func (job *Duration) Stop() {
	job.stop <- struct{}{}
}
