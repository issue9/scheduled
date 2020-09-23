// SPDX-License-Identifier: MIT

package cron

import "time"

func (c *cron) Next(last time.Time) time.Time {
	return c.nextTime(last, true)
}

func (c *cron) nextTime(last time.Time, carry bool) time.Time {
	second, carry := bounds[secondIndex].next(last.Second(), c.data[secondIndex], carry)
	minute, carry := bounds[minuteIndex].next(last.Minute(), c.data[minuteIndex], carry)
	hour, carry := bounds[hourIndex].next(last.Hour(), c.data[hourIndex], carry)

	var year, month, day int
	if c.data[weekIndex] != any && c.data[weekIndex] != step {
		year, month, day = c.nextWeekDay(last, carry)
	} else {
		year, month, day = c.nextMonthDay(last, carry)
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, last.Location())
}

func (c *cron) nextMonthDay(last time.Time, carry bool) (year, month, day int) {
	day, carry = bounds[dayIndex].next(last.Day(), c.data[dayIndex], carry)
	month, carry = bounds[monthIndex].next(int(last.Month()), c.data[monthIndex], carry)
	year = last.Year()
	if carry {
		year++
	}

	for { // 由于月份中的天数不固定，还得计算该天数是否存在于当前月分
		days := getMonthDays(time.Month(month), year)
		if day <= days { // 天数存在于当前月，则退出循环
			return year, month, day
		}

		month, carry = bounds[monthIndex].next(month, c.data[monthIndex], true)
		if carry {
			year++
		}
	}
}

func (c *cron) nextWeekDay(last time.Time, carry bool) (year, month, day int) {
	// 计算 week day 在当前月份中的日期
	wday, ca := bounds[weekIndex].next(int(last.Weekday()), c.data[weekIndex], carry)
	dur := wday - int(last.Weekday()) // 相差的天数
	if (dur < 0) || (ca && dur == 0) {
		dur += 7
	}
	day = dur + last.Day() // wday 在当前月对应的天数
	year = last.Year()

	// 此处忽略返回的 c 参数。参数 carry 为 false，则肯定不会返回值为 true 的 carry
	month, _ = bounds[monthIndex].next(int(last.Month()), c.data[monthIndex], false)
	if time.Month(month) != last.Month() {
		day = getMonthWeekDay(time.Weekday(wday), time.Month(month), year)
	} else if days := getMonthDays(time.Month(month), year); day > days {
		// 跨月份，还有可能跨年份
		month, ca = bounds[monthIndex].next(month, c.data[monthIndex], true)
		if ca {
			year++
		}
		day = getMonthWeekDay(time.Weekday(wday), time.Month(month), year)
	}

	// 同时设置了 day，需要比较两个值哪个更近
	if c.data[dayIndex] != any && c.data[dayIndex] != step {
		y, m, d := c.nextMonthDay(last, carry)
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

// 从 list 中获取与 curr 最近的下一个值
//
// 如果 carry 为 false，且 curr 存在于 list 则有可能返回 curr 本身。
//
// curr 当前的时间值；
// list 可用的时间值；
// carry 前一个数值是否已经进位；
// val 返回计算后的最近一个时间值；
// c 是否需要下一个值进位。
func (b bound) next(curr int, list uint64, carry bool) (val int, c bool) {
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

	var min int
	var hasMin bool
	for i := b.min; i <= b.max; i++ {
		if ((uint64(1) << uint64(i)) & list) <= 0 { // 该位未被设置为 1
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
