// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package schedule 实现了时间部分时间调度的算法。
package schedule

import "time"

// Scheduler 为所有的时间调度算法指定一个统一的接口
type Scheduler interface {
	// 生成下一次的时间。相对于 last 时间。
	//
	// 如果不需要再执行了，则应该返回一个零值。
	Next(last time.Time) time.Time

	// Title 返回当前 Nexter 的一个名称。
	Title() string
}
