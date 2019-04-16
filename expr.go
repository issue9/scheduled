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

	second, carry := next(uint8(last.Second()), e.data[secondIndex], false)
	minute, carry := next(uint8(last.Minute()), e.data[minuteIndex], carry)
	hour, carry := next(uint8(last.Hour()), e.data[hourIndex], carry)

	var day int
	if e.data[weekIndex] != nil { // 除非指定了星期，否则永远按照日期来
		weekday, _ := next(uint8(last.Weekday()), e.data[weekIndex], carry)
		dur := weekday - int(last.Weekday()) // 相差的天数
		day = dur + last.Day()
	} else {
		day, carry = next(uint8(last.Day()), e.data[dayIndex], carry)
	}

	month, carry := next(uint8(last.Month()), e.data[monthIndex], carry)
	year := last.Year()
	if carry {
		year++
	}

	// 由于月份中的天数不固定，还得计算该天数是否存在于当前月分
	for {
		days := getMonthDays(time.Month(month), year)
		if day <= days { // 天数存在于当前月，则退出循环
			break
		}

		month, carry = next(uint8(month), e.data[monthIndex], false)
		if carry {
			year++
		}
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, last.Location())
}

// 获取指定月份的天数
func getMonthDays(month time.Month, year int) int {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
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
