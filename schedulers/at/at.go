// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package at 提供类似于 at 指令的定时器
package at

import (
	"time"

	"github.com/issue9/scheduled/schedulers"
)

// At 返回只在指定时间执行一次的调度器
func At(t time.Time) schedulers.Scheduler {
	return schedulers.SchedulerFunc(func(last time.Time) time.Time {
		if !t.After(last) {
			return time.Time{}
		}
		return t
	})
}
