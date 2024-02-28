// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package schedulers 实现了部分时间调度的算法
package schedulers

import "time"

// Scheduler 时间调度算法需要实现的接口
type Scheduler interface {
	// Next 生成相对于 last 的下一次时间。
	//
	// 如果返回的时间值，已经小于当前时间，那么该任务会被安排在最先执行。
	// 如果返回是零值，表示该调度已经终结，不会再执行该任务，后续也都应返回零值。
	//
	// 传递相同的 last 参数，其返回值应该也相同，或是返回一个零值。
	Next(last time.Time) time.Time
}

type SchedulerFunc func(time.Time) time.Time

func (f SchedulerFunc) Next(last time.Time) time.Time { return f(last) }
