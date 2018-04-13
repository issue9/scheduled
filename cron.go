// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package cron 执行定时任务的包
package cron

// 表示任务状态
const (
	Running State = 1 << iota
	Stoped
	Failed
)

// State 状态值类型
type State int8

// Job 一个定时任务的基本接口
type Job interface {
	Name() string
	State() State
	Err() error // 状态为 failed 时的错误信息
	Start() error
	Stop()
}

// Cron 管理定时任务
type Cron struct {
	jobs []Job
}

// New 声明一个 Cron 实例
func New() *Cron {
	return &Cron{
		jobs: make([]Job, 0, 10),
	}
}

// Start 启动所有的任务
func (cron *Cron) Start() error {
	for _, job := range cron.jobs {
		if err := job.Start(); err != nil {
			return err
		}
	}

	return nil
}

// Stop 停止所有的任务
func (cron *Cron) Stop() {
	for _, job := range cron.jobs {
		job.Stop()
	}
}

// Jobs 返回所有的任务列表
func (cron *Cron) Jobs() []Job {
	return cron.jobs
}
