// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import (
	"fmt"
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

// JobFunc 每一个定时任务实际上执行的函数签名
type JobFunc func() error

// Job 一个定时任务的基本接口
type Job struct {
	name  string
	f     JobFunc
	n     Nexter
	state State
	err   error // 出错时的错误内容

	prev, next time.Time
}

// New 添加一个新的定时任务
func (c *Cron) New(name string, f JobFunc, n Nexter) {
	c.jobs = append(c.jobs, &Job{
		name:  name,
		f:     f,
		n:     n,
		state: Stoped,
	})
}

// Name 任务的名称
func (j *Job) Name() string { return j.name }

//
func (j *Job) Description() string { return j.n.Title() }

// State 获取当前的状态
func (j *Job) State() State { return j.state }

// 运行当前的任务
func (j *Job) run() {
	defer func() {
		if msg := recover(); msg != nil {
			if err, ok := msg.(error); ok {
				j.err = err
			} else {
				j.err = fmt.Errorf("job error: %v", msg)
			}

			j.state = Failed
		}
	}()

	j.state = Running
	j.err = j.f()

	if j.err != nil {
		j.state = Failed
	} else {
		j.state = Stoped
		j.err = nil
	}

	j.prev = j.next
	j.next = j.n.Next(j.next)
}

// 初始化当前任务，获取其下次执行时间。
func (j *Job) init(now time.Time) {
	j.next = j.n.Next(now)
}
