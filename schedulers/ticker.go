// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package schedulers

import "time"

// Tick 声明一个固定时间段的定时任务
//
// d 必须大于等于 1 秒，否则会 panic。
// 如果任务的执行时间超过 d，那么调度时间将顺延下一个时间点。
func Tick(d time.Duration) Scheduler {
	if d < time.Second {
		panic("参数 d 的值必须在 1 秒以上")
	}
	return SchedulerFunc(func(last time.Time) time.Time { return last.Add(d) })
}
