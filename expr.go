// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import "time"

// 表示 cron 语法表达式中的顺序
const (
	secondIndex = iota
	minuteIndex
	hourIndex
	dayIndex
	monthIndex
	weekIndex
	indexSize
)

type expr struct {
	// 依次保存着 cron 语法中各个字段解析后的内容。
	//
	// 最长的秒数，最多 60 位，正好可以使用一个 uint64 保存，
	// 其它类型也各自占一个字段长度。
	//
	// 其中每个字段中，从 0 位到高位，每一位表示一个值，比如在秒字段中，
	//  0,1,7 表示为 ...10000011
	// 如果是月份这种从 1 开始的，则其第一位永远是 0
	data []uint64

	next  time.Time
	title string
}

// NewExpr 使用 cron 表示式新建一个定时任务
//
// expr 的值可以是：
//  * * * * * *
//  | | | | | |
//  | | | | | --- 星期
//  | | | | ----- 月
//  | | | ------- 日
//  | | --------- 小时
//  | ----------- 分
//  ------------- 秒
//
// 星期与日若同时存在，则以或的形式组合。
//
// 支持以下符号：
//  - 表示范围
//  , 表示和
//
// 同时支持以下便捷指令：
//  @yearly:   0 0 0 1 1 *
//  @annually: 0 0 0 1 1 *
//  @monthly:  0 0 0 1 * *
//  @weekly:   0 0 0 * * 0
//  @daily:    0 0 0 * * *
//  @midnight: 0 0 0 * * *
//  @hourly:   0 0 * * * *
func (c *Cron) NewExpr(name string, f JobFunc, expr string) error {
	next, err := parseExpr(expr)
	if err != nil {
		return err
	}

	c.New(name, f, next)
	return nil
}

// Title 获取标题名称
func (e *expr) Title() string {
	return e.title
}

// Next 计算下个时间点，相对于 last
func (e *expr) Next(last time.Time) time.Time {
	if e.next.After(last) {
		return e.next
	}

	e.next = e.nextTime(last, true)
	return e.next
}

func (e *expr) nextTime(last time.Time, carry bool) time.Time {
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

func (e *expr) nextMonthDay(last time.Time, carry bool) (year int, month, day uint8) {
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

func (e *expr) nextWeekDay(last time.Time, carry bool) (year int, month, day uint8) {
	// 计算 week day 在当前月份中的日期
	wday, c := fields[weekIndex].next(uint8(last.Weekday()), e.data[weekIndex], carry)
	dur := int(wday) - int(last.Weekday()) // 相差的天数
	if (dur < 0) || (c && dur == 0) {
		dur += 7
	}
	day = uint8(dur) + uint8(last.Day()) // wday 在当前月对应的天数
	year = last.Year()

	// 此处忽略返回的 c 参数。参数 carry 为 false，则肯定不会返回值为 true 的 carry
	month, _ = fields[monthIndex].next(uint8(last.Month()), e.data[monthIndex], false)
	if time.Month(month) != last.Month() {
		day = getMonthWeekDay(time.Weekday(wday), time.Month(month), year)
	} else if days := getMonthDays(time.Month(month), year); day > days {
		// 跨月份，还有可能跨年份
		month, c = fields[monthIndex].next(uint8(month), e.data[monthIndex], true)
		if c {
			year++
		}
		day = getMonthWeekDay(time.Weekday(wday), time.Month(month), year)
	}

	// 同时设置了 day，需要比较两个值哪个更近
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
