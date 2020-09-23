// SPDX-License-Identifier: MIT

// Package at 提供类似于 at 指令的定时器
package at

import (
	"time"

	"github.com/issue9/scheduled/schedulers"
)

// Layout Parse 解析时间的格式
//
// 同时也是 Title 返回的格式。
const Layout = "2006-01-02 15:04:05"

var zero = time.Time{}

type scheduler struct {
	title string

	month                           time.Month
	year, day, hour, minute, second int

	// 是否已经被使用，只要被调用过一次 Next
	// 表示该时间已经被使用，之后将返回零值。
	used bool
}

// At 返回只在指定时间执行一次的调度器
//
// t 为一个正常的时间字符串，在该时间执行一次 f。若时间早于当前时间，
// 则在启动之后立马执行，如果 t 的值为零，则不会被执行。
func At(t string) (schedulers.Scheduler, error) {
	tt, err := time.ParseInLocation(Layout, t, time.UTC)
	if err != nil {
		return nil, err
	}

	return at(tt), nil
}

func at(t time.Time) schedulers.Scheduler {
	year, month, day := t.Date()
	hour, minute, sec := t.Clock()
	return &scheduler{
		title:  t.Format(Layout),
		year:   year,
		month:  month,
		day:    day,
		hour:   hour,
		minute: minute,
		second: sec,
	}
}

func (s *scheduler) Title() string {
	return s.title
}

func (s *scheduler) Next(last time.Time) time.Time {
	if s.used {
		return zero
	}
	s.used = true
	return time.Date(s.year, s.month, s.day, s.hour, s.minute, s.second, 0, last.Location())
}
