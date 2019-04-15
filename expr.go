// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import (
	"sort"
	"time"
)

// Expr 表达式的分析结果
type Expr struct {
	seconds []uint8 // 0-59
	minutes []uint8 // 0-59
	days    []uint8 // 1-31
	hours   []uint8 // 0-23
	months  []uint8 // 1-12
	weeks   []uint8 // 0-7

	next  time.Time
	title string
}

func sortUint8(vals []uint8) {
	sort.SliceStable(vals, func(i, j int) bool {
		return vals[i] < vals[j]
	})
}

// Title 获取标题名称
func (e *Expr) Title() string {
	if e.title == "" {
		// TODO
	}

	return e.title
}

// Next 计算下个时间点，相对于 last
func (e *Expr) Next(last time.Time) time.Time {
	if e.next.After(last) {
		return e.next
	}

	second, carry := next(uint8(last.Second()), e.seconds, false)
	minute, carry := next(uint8(last.Minute()), e.minutes, carry)
	hour, carry := next(uint8(last.Hour()), e.hours, carry)

	var day int
	if e.weeks != nil { // 除非指定了星期，否则永远按照日期来
		weekday, _ := next(uint8(last.Weekday()), e.weeks, carry)
		dur := weekday - int(last.Weekday()) // 相差的天数
		day = dur + last.Day()
	} else {
		day, carry = next(uint8(last.Day()), e.days, carry)
	}

	month, carry := next(uint8(last.Month()), e.months, carry)
	year := last.Year()
	if carry {
		year++
	}

	// 由于月份中的天数不固定，还得计算该天数是否存在于当前月分
	for {
		days := getMonthDays(time.Month(month), year, last.Location())
		if day <= days { // 天数存在于当前月，则退出循环
			break
		}

		month, carry = next(uint8(month), e.months, false)
		if carry {
			year++
		}
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, last.Location())
}

func getMonthDays(month time.Month, year int, loc *time.Location) int {
	first := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	last := first.AddDate(0, 1, -1)
	return last.Day()
}

// curr 当前的时间值；
// list 可用的时间值；
// carry 是否需要当前时间进位；
// val 返回计算后的最近一个时间值；
// c 是否已经进位。
func next(curr uint8, list []uint8, carry bool) (val int, c bool) {
	if list == nil {
		if carry {
			curr++
		}
		return int(curr), false
	}

	for _, item := range list {
		switch {
		case item == curr: // 存在与当前值相同的值
			if !carry {
				return int(item), false
			}
		case item > curr:
			return int(item), false
		}
	}

	// 大于当前列表的最大值，则返回列表中的最大值，则设置进位标记
	return int(list[0]), true
}

// Seconds 表示秒数，0-59
func (e *Expr) Seconds(s ...uint8) *Expr {
	e.seconds = s
	sortUint8(e.seconds)
	return e
}

// Minutes 指定分钟数，0-59
func (e *Expr) Minutes(m ...uint8) *Expr {
	e.minutes = m
	sortUint8(e.minutes)
	return e
}

// Hours 指定小时，0-23
func (e *Expr) Hours(h ...uint8) *Expr {
	e.hours = h
	sortUint8(e.hours)
	return e
}

// Days 指定天数，1-31
func (e *Expr) Days(d ...uint8) *Expr {
	e.days = d
	sortUint8(e.days)
	return e
}

// Months 指定月份，1-12
func (e *Expr) Months(m ...uint8) *Expr {
	e.months = m
	sortUint8(e.months)
	return e
}

// Weeks 指定星期，0-6，其中 0 表示周日
func (e *Expr) Weeks(w ...uint8) *Expr {
	e.weeks = w
	sortUint8(e.weeks)
	return e
}

// Seconds 表示秒数，0-59
func Seconds(s ...uint8) *Expr {
	return (&Expr{}).Seconds(s...)
}

// Minutes 指定分钟数，0-59
func Minutes(m ...uint8) *Expr {
	return (&Expr{}).Minutes(m...)
}

// Hours 指定小时，0-23
func Hours(h ...uint8) *Expr {
	return (&Expr{}).Hours(h...)
}

// Days 指定天数，1-31
func Days(d ...uint8) *Expr {
	return (&Expr{}).Days(d...)
}

// Months 指定月份，1-12
func Months(m ...uint8) *Expr {
	return (&Expr{}).Months(m...)
}

// Weeks 指定星期，0-6，其中 0 表示周日
func Weeks(w ...uint8) *Expr {
	return (&Expr{}).Weeks(w...)
}
