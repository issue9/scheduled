// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import (
	"strconv"
	"time"
)

// 该值的顺序与 cron 中语法的顺序相同
const (
	secondIndex = iota
	minuteIndex
	hourIndex
	dayIndex
	monthIndex
	weekIndex
	typeSize
)

// Expr 表达式的分析结果
type Expr struct {
	data       []uint64
	startIndex int

	next  time.Time
	title string
}

// NewExpr 新建表达式定时器
//
// expr 的值可以是：
//  * * * * * *  cmd
//  | | | | | |
//  | | | | | --- 星期
//  | | | | ----- 月
//  | | | ------- 日
//  | | --------- 小时
//  | ----------- 分
//  ------------- 秒
//
// 支持以下符号：
//  - 表示范围
//  , 表示和
func (c *Cron) NewExpr(name string, f JobFunc, expr string) error {
	next, err := parseExpr(expr)
	if err != nil {
		return err
	}

	c.New(name, f, next)
	return nil
}

// Title 获取标题名称
func (e *Expr) Title() string {
	return e.title
}

// 当前的时间是否与 t 相等
func (e *Expr) equal(t1, t2 time.Time) bool {
	return true
	for i := e.startIndex; i < typeSize; i++ {
		switch i {
		case secondIndex:
			if t1.Second() != t2.Second() {
				return false
			}
		case minuteIndex:
			if t1.Minute() != t2.Minute() {
				return false
			}
		case hourIndex:
			if t1.Hour() != t2.Hour() {
				return false
			}
		case monthIndex:
			if t1.Month() != t2.Month() {
				return false
			}
		case dayIndex:
			if (e.data[dayIndex] != step) && (t1.Day() != t2.Day()) {
				return false
			}
		case weekIndex:
			if (e.data[weekIndex] != step) && (t1.Weekday() != t2.Weekday()) {
				return false
			}
		default:
			panic("超出了范围" + strconv.Itoa(e.startIndex))
		}
	}

	return true
}

// Next 计算下个时间点，相对于 last
func (e *Expr) Next(last time.Time) time.Time {
	if e.next.After(last) {
		return e.next
	}

	carry := e.equal(e.next, last)
	e.next = e.nextTime(last, carry)
	return e.next
}

func (e *Expr) nextTime(last time.Time, carry bool) time.Time {
	second, carry := fields[secondIndex].next(uint8(last.Second()), e.data[secondIndex], carry)
	minute, carry := fields[minuteIndex].next(uint8(last.Minute()), e.data[minuteIndex], carry)
	hour, carry := fields[hourIndex].next(uint8(last.Hour()), e.data[hourIndex], carry)

	var year int
	var month, day uint8
	if e.data[weekIndex] != any && e.data[weekIndex] != step {
		year, month, day = e.nextWeekDay(last, carry)
	} else {
		year, month, day = e.nextMonthDay(last, carry)
	}

	return time.Date(year, time.Month(month), int(day), int(hour), int(minute), int(second), 0, last.Location())
}

func (e *Expr) nextMonthDay(last time.Time, carry bool) (year int, month, day uint8) {
	day, carry = fields[dayIndex].next(uint8(last.Day()), e.data[dayIndex], carry)
	month, carry = fields[monthIndex].next(uint8(last.Month()), e.data[monthIndex], carry)
	year = last.Year()
	if carry {
		year++
	}

	for { // 由于月份中的天数不固定，还得计算该天数是否存在于当前月分
		days := getMonthDays(time.Month(month), year)
		if day <= days { // 天数存在于当前月，则退出循环
			return year, month, day
		}

		month, carry = fields[monthIndex].next(uint8(month), e.data[monthIndex], true)
		if carry {
			year++
		}
	}
}

func (e *Expr) nextWeekDay(last time.Time, carry bool) (year int, month, day uint8) {
	// 计算 week day 在当前月份中的日期
	wday, c := fields[weekIndex].next(uint8(last.Weekday()), e.data[weekIndex], carry)
	dur := int(wday) - int(last.Weekday()) // 相差的天数
	if (dur < 0) || (c && dur == 0) {
		dur += 7
	}
	day = uint8(dur) + uint8(last.Day()) // wday 在当前月对应的天数
	year = last.Year()
	month, _ = fields[monthIndex].next(uint8(last.Month()), e.data[monthIndex], false)

	if days := getMonthDays(time.Month(month), year); day > days { // 跨月份，还有可能跨年份
		month, c = fields[monthIndex].next(uint8(month), e.data[monthIndex], true)
		if c {
			year++
		}
		day = getMonthWeekDay(time.Weekday(wday), time.Month(month), year)
	}

	if e.data[dayIndex] != any && e.data[dayIndex] != step {
		y, m, d := e.nextMonthDay(last, carry)
		if !(year < y || month < m || day < d) {
			year = y
			month = m
			day = d
		}
	}

	return year, month, day
}

// 获取指定月份的天数
func getMonthDays(month time.Month, year int) uint8 {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	return uint8(last.Day())
}

// 获取指定 year-month 下第一个符合 weekday 对应的天数
func getMonthWeekDay(weekday time.Weekday, month time.Month, year int) uint8 {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	weekday -= first.Weekday()
	if weekday < 0 {
		weekday += 7
	}
	last := first.AddDate(0, 0, int(weekday))
	return uint8(last.Day())
}
