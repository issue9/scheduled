// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import "time"

// Expr 表达式的分析结果
type Expr struct {
	seconds uint64 // 0-59
	minutes uint64 // 0-59
	days    uint32 // 1-31
	hours   uint32 // 0-23
	months  uint16 // 1-12
	weeks   uint8  // 0-7

	next  time.Time
	title string
}

// Title 获取标题名称
func (e *Expr) Title() string {
	if e.title == "" {
		// TODO
	}

	return e.title
}

// Next 下个时间点
func (e *Expr) Next(last time.Time) time.Time {
	if e.next.After(last) {
		return e.next
	}

	var carry bool

	year := last.Year()
	month := last.Month()
	//
}

// Seconds 表示秒数，0-59
func (e *Expr) Seconds(s ...uint64) *Expr {
	for _, ss := range s {
		e.seconds |= 1 << ss
	}
	return e
}

// Minutes 指定分钟数，0-59
func (e *Expr) Minutes(m ...uint64) *Expr {
	for _, ss := range m {
		e.minutes |= 1 << ss
	}
	return e
}

// Hours 指定小时，0-23
func (e *Expr) Hours(h ...uint32) *Expr {
	for _, ss := range h {
		e.hours |= 1 << ss
	}
	return e
}

// Days 指定天数，1-31
func (e *Expr) Days(d ...uint32) *Expr {
	for _, ss := range d {
		e.days |= 1 << (ss - 1)
	}
	return e
}

// Months 指定月份，1-12
func (e *Expr) Months(m ...uint16) *Expr {
	for _, ss := range m {
		e.months |= 1 << (ss - 1)
	}
	return e
}

// Weeks 指定星期，0-6，其中 0 表示周日
func (e *Expr) Weeks(w ...uint8) *Expr {
	for _, ss := range w {
		e.weeks |= 1 << ss
	}
	return e
}

// Seconds 表示秒数，0-59
func Seconds(s ...uint64) *Expr {
	return (&Expr{}).Seconds(s...)
}

// Minutes 指定分钟数，0-59
func Minutes(m ...uint64) *Expr {
	return (&Expr{}).Minutes(m...)
}

// Hours 指定小时，0-23
func Hours(h ...uint32) *Expr {
	return (&Expr{}).Hours(h...)
}

// Days 指定天数，1-31
func Days(d ...uint32) *Expr {
	return (&Expr{}).Days(d...)
}

// Months 指定月份，1-12
func Months(m ...uint16) *Expr {
	return (&Expr{}).Months(m...)
}

// Weeks 指定星期，0-6，其中 0 表示周日
func Weeks(w ...uint8) *Expr {
	return (&Expr{}).Weeks(w...)
}
