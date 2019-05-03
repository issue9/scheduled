// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package expr

import "time"

// Next 计算下个时间点，相对于 last
func (e *Expr) Next(last time.Time) time.Time {
	if e.next.After(last) {
		return e.next
	}

	e.next = e.nextTime(last, true)
	return e.next
}

func (e *Expr) nextTime(last time.Time, carry bool) time.Time {
	second, carry := bounds[secondIndex].next(uint8(last.Second()), e.data[secondIndex], carry)
	minute, carry := bounds[minuteIndex].next(uint8(last.Minute()), e.data[minuteIndex], carry)
	hour, carry := bounds[hourIndex].next(uint8(last.Hour()), e.data[hourIndex], carry)

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
	day, carry = bounds[dayIndex].next(uint8(last.Day()), e.data[dayIndex], carry)
	month, carry = bounds[monthIndex].next(uint8(last.Month()), e.data[monthIndex], carry)
	year = last.Year()
	if carry {
		year++
	}

	for { // 由于月份中的天数不固定，还得计算该天数是否存在于当前月分
		days := getMonthDays(time.Month(month), year)
		if day <= days { // 天数存在于当前月，则退出循环
			return year, month, day
		}

		month, carry = bounds[monthIndex].next(uint8(month), e.data[monthIndex], true)
		if carry {
			year++
		}
	}
}

func (e *Expr) nextWeekDay(last time.Time, carry bool) (year int, month, day uint8) {
	// 计算 week day 在当前月份中的日期
	wday, c := bounds[weekIndex].next(uint8(last.Weekday()), e.data[weekIndex], carry)
	dur := int(wday) - int(last.Weekday()) // 相差的天数
	if (dur < 0) || (c && dur == 0) {
		dur += 7
	}
	day = uint8(dur) + uint8(last.Day()) // wday 在当前月对应的天数
	year = last.Year()

	// 此处忽略返回的 c 参数。参数 carry 为 false，则肯定不会返回值为 true 的 carry
	month, _ = bounds[monthIndex].next(uint8(last.Month()), e.data[monthIndex], false)
	if time.Month(month) != last.Month() {
		day = getMonthWeekDay(time.Weekday(wday), time.Month(month), year)
	} else if days := getMonthDays(time.Month(month), year); day > days {
		// 跨月份，还有可能跨年份
		month, c = bounds[monthIndex].next(uint8(month), e.data[monthIndex], true)
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

// curr 当前的时间值；
// list 可用的时间值；
// carry 前一个数值是否已经进位；
// val 返回计算后的最近一个时间值；
// c 是否需要一个值进位。
func (b bound) next(curr uint8, list uint64, carry bool) (val uint8, c bool) {
	if list == any {
		return curr, carry
	}

	if list == step {
		if carry {
			curr++
		}

		if curr > b.max {
			return b.min, true
		}
		return curr, false
	}

	var min uint8
	var hasMin bool
	for i := b.min; i <= b.max; i++ {
		if ((uint64(1) << i) & list) <= 0 { // 该位未被设置为 1
			continue
		}

		if i > curr {
			return i, false
		}

		if !hasMin {
			min = i
			hasMin = true
		}

		if i == curr {
			if !carry {
				return i, false
			}
			carry = false
		}
	} // end for

	// 大于当前列表的最大值，则返回列表中的最小值，并设置进位标记
	return min, true
}
