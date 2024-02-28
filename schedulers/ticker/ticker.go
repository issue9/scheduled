// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package ticker 时间段固定的定时器
package ticker

import (
	"time"

	"github.com/issue9/scheduled/schedulers"
)

// Tick 声明一个固定时间段的定时任务
//
// imm 是否立即执行一次任务，如果为 true，
// 则会在第一次调用 Next 时返回当前时间。
func Tick(d time.Duration, imm bool) schedulers.Scheduler {
	if d < time.Second {
		panic("参数 d 的值必须在 1 秒以上")
	}

	return schedulers.SchedulerFunc(func(last time.Time) time.Time {
		if imm {
			imm = false
			return time.Now()
		}

		return last.Add(d)
	})
}
