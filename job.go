// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import "time"

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
	last  time.Time
	state State
	next  Nexter
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
