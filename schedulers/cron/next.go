// SPDX-License-Identifier: MIT

package cron

import "time"

type datetime struct {
	year                 int
	month                time.Month
	day                  int
	weekday              time.Weekday
	hour, minute, second int
}

func (c *cron) Next(last time.Time) time.Time {
	last = last.In(c.loc)

	dt := &datetime{}
	dt.year, dt.month, dt.day = last.Date()
	dt.weekday = last.Weekday()
	dt.hour = last.Hour()
	dt.minute = last.Minute()
	dt.second = last.Second()

	second, carry := c.data[secondIndex].next(dt.second, bounds[secondIndex], true)
	minute, carry := c.data[minuteIndex].next(dt.minute, bounds[minuteIndex], carry)
	hour, carry := c.data[hourIndex].next(dt.hour, bounds[hourIndex], carry)

	var year, month, day int
	if c.data[weekIndex] != any && c.data[weekIndex] != step {
		year, month, day = c.nextWeekDay(dt, carry)
	} else {
		year, month, day = c.nextMonthDay(dt, carry)
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, last.Location())
}

func (c *cron) nextMonthDay(dt *datetime, carry bool) (year, month, day int) {
	dayBounds := bounds[dayIndex]
	dayBounds.max = getMonthDays(dt.month, dt.year) // 最大的天数根据月份不同而不同

	day, carry = c.data[dayIndex].next(dt.day, dayBounds, carry)
	month, carry = c.data[monthIndex].next(int(dt.month), bounds[monthIndex], carry)
	year = dt.year
	if carry {
		year++
	}

	for { // 由于月份中的天数不固定，还得计算该天数是否存在于当前月分
		days := getMonthDays(time.Month(month), year)
		if day <= days { // 天数存在于当前月，则退出循环
			return year, month, day
		}

		month, carry = c.data[monthIndex].next(month, bounds[monthIndex], true)
		if carry {
			year++
		}
	}
}

func (c *cron) nextWeekDay(dt *datetime, carry bool) (year, month, day int) {
	// 计算 week day 在当前月份中的日期
	wday, ca := c.data[weekIndex].next(int(dt.weekday), bounds[weekIndex], carry)
	dur := wday - int(dt.weekday) // 相差的天数
	if (dur < 0) || (ca && dur == 0) {
		dur += 7
	}
	day = dur + dt.day // wday 在当前月对应的天数
	year = dt.year

	// 此处忽略返回的 c 参数。参数 carry 为 false，则肯定不会返回值为 true 的 carry
	month, _ = c.data[monthIndex].next(int(dt.month), bounds[monthIndex], false)
	if time.Month(month) != dt.month {
		day = getMonthWeekDay(time.Weekday(wday), time.Month(month), year)
	} else if days := getMonthDays(time.Month(month), year); day > days {
		// 跨月份，还有可能跨年份
		month, ca = c.data[monthIndex].next(month, bounds[monthIndex], true)
		if ca {
			year++
		}
		day = getMonthWeekDay(time.Weekday(wday), time.Month(month), year)
	}

	// 同时设置了 day，需要比较两个值哪个更近
	if c.data[dayIndex] != any && c.data[dayIndex] != step {
		y, m, d := c.nextMonthDay(dt, carry)
		if !(year < y || month < m || day < d) {
			year = y
			month = m
			day = d
		}
	}

	return year, month, day
}

// 获取指定月份的天数
func getMonthDays(month time.Month, year int) int {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	return last.Day()
}

// 获取指定 year-month 下第一个符合 weekday 对应的天数
func getMonthWeekDay(weekday time.Weekday, month time.Month, year int) int {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	weekday -= first.Weekday()
	if weekday < 0 {
		weekday += 7
	}
	last := first.AddDate(0, 0, int(weekday))
	return last.Day()
}
