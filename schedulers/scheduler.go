// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package schedulers 实现了部分时间调度的算法。
package schedulers

import "time"

// Scheduler 时间调度算法需要实现的接口
type Scheduler interface {
	// 生成下一次的时间。相对于 last 时间。
	//
	// 如果返回的时间值，已经小于当前时间，那么该任务会被安排在最先执行。
	// 如果返回是零值，表示该调度已经终结，后续都将返回零。
	//
	// 实现者应该继承 last 的时区信息，即返回值的时区应该和 last 相同，
	// 否则其结果是未定义的。
	//
	// 传递相同的 last 参数，其返回值应该也是相同，或是返回一个零值。
	Next(last time.Time) time.Time

	// Title 返回用于描述当前算法的一个简短介绍。
	Title() string
}
