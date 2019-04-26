// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import "time"

// 该值的顺序与 cron 中语法的顺序相同
const (
	secondIndex = iota
	minuteIndex
	hourIndex
	dayIndex
	monthIndex
	weekIndex
)

// Expr 表达式的分析结果
type Expr struct {
	data [][]uint8

	next  time.Time
	title string
}

// Title 获取标题名称
func (e *Expr) Title() string {
	return e.title
}

// Next 计算下个时间点，相对于 last
func (e *Expr) Next(last time.Time) time.Time {
	if e.next.After(last) {
		return e.next
	}

	next := e.nextTime(last, false)
	if next.Unix() == last.Unix() { // 相等（不考虑秒以下的内容），则需要重新计算
		next = e.nextTime(last, true)
	}

	e.next = next
	return next
}

func (e *Expr) nextTime(last time.Time, carry bool) time.Time {
	second, carry := next(secondIndex, uint8(last.Second()), e.data[secondIndex], carry)
	minute, carry := next(minuteIndex, uint8(last.Minute()), e.data[minuteIndex], carry)
	hour, carry := next(hourIndex, uint8(last.Hour()), e.data[hourIndex], carry)

	var month, day uint8
	var year int
	if e.data[weekIndex] != nil {
		year, month, day = e.nextWeekDay(last, carry)
	} else {
		year, month, day = e.nextMonthDay(last, carry)
	}

	return time.Date(year, time.Month(month), int(day), int(hour), int(minute), int(second), 0, last.Location())
}

func (e *Expr) nextMonthDay(last time.Time, carry bool) (year int, month, day uint8) {
	day, carry = next(dayIndex, uint8(last.Day()), e.data[dayIndex], carry)
	month, carry = next(monthIndex, uint8(last.Month()), e.data[monthIndex], carry)
	year = last.Year()
	if carry {
		year++
	}

	// 由于月份中的天数不固定，还得计算该天数是否存在于当前月分
	for {
		days := getMonthDays(time.Month(month), year)
		if day <= days { // 天数存在于当前月，则退出循环
			break
		}

		month, carry = next(monthIndex, uint8(month), e.data[monthIndex], true)
		if carry {
			year++
		}
	}

	return year, month, day
}

func (e *Expr) nextWeekDay(last time.Time, carry bool) (year int, month, day uint8) {
	// 计算 week day 在当前月份中的日期
	wday, _ := next(weekIndex, uint8(last.Weekday()), e.data[weekIndex], carry)
	dur := wday - uint8(last.Weekday()) // 相差的天数
	if dur <= 0 {
		dur += 7
	}
	day = dur + uint8(last.Day())

	// 如果 day 字段不为空，则还需要比较两者之间哪个值更接近
	if e.data[dayIndex] != nil {
		wday, carry = next(dayIndex, uint8(last.Day()), e.data[dayIndex], carry)
		if !carry && (wday < day) {
			day = wday
		}
	}

	month, _ = next(monthIndex, uint8(last.Month()), e.data[monthIndex], carry)
	year = last.Year()
	// 由于月份中的天数不固定，还得计算该天数是否存在于当前月分
	for {
		days := getMonthDays(time.Month(month), year)
		if day <= days { // 天数存在于当前月，则退出循环
			break
		}

		month, carry = next(monthIndex, uint8(month), e.data[monthIndex], true)
		if carry {
			year++
		}
		day -= days
	}

	return year, month, day
}

// 获取指定月份的天数
func getMonthDays(month time.Month, year int) uint8 {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	return uint8(last.Day())
}

// curr 当前的时间值；
// list 可用的时间值；
// carry 前一个数值是否已经进位；
// val 返回计算后的最近一个时间值；
// c 是否需要一个值进位。
func next(typ int, curr uint8, list []uint8, carry bool) (val uint8, c bool) {
	if list == nil {
		if carry {
			curr++
		}

		b := bounds[typ]
		if curr > b.max {
			return b.min, true
		}
		return curr, false
	}

	for _, item := range list {
		switch {
		case item == curr: // 存在与当前值相同的值
			if !carry {
				return item, false
			}
		case item > curr:
			return item, false
		}
	}

	// 大于当前列表的最大值，则返回列表中的最大值，则设置进位标记
	return list[0], true
}
