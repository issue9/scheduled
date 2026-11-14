// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package ticker 时间段固定的定时器
package ticker

import (
	"time"

	"github.com/issue9/scheduled/schedulers"
)

// Tick 声明一个固定时间段的定时任务
func Tick(d time.Duration) schedulers.Scheduler {
	if d < time.Second {
		panic("参数 d 的值必须在 1 秒以上")
	}

	return schedulers.SchedulerFunc(func(last time.Time) time.Time { return last.Add(d) })
}
