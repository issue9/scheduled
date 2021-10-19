// SPDX-License-Identifier: MIT

// Package at 提供类似于 at 指令的定时器
package at

import (
	"time"

	"github.com/issue9/scheduled/schedulers"
)

const layout = "2006-01-02 15:04:05"

var zero = time.Time{}

type scheduler struct {
	month                           time.Month
	year, day, hour, minute, second int

	// 是否已经被使用，只要被调用过一次 Next
	// 表示该时间已经被使用，之后将返回零值。
	used bool
}

// At 返回只在指定时间执行一次的调度器
func At(t time.Time) schedulers.Scheduler {
	year, month, day := t.Date()
	hour, minute, sec := t.Clock()
	return &scheduler{
		year:   year,
		month:  month,
		day:    day,
		hour:   hour,
		minute: minute,
		second: sec,
	}
}

func (s *scheduler) Next(last time.Time) time.Time {
	if s.used {
		return zero
	}
	s.used = true
	return time.Date(s.year, s.month, s.day, s.hour, s.minute, s.second, 0, last.Location())
}
