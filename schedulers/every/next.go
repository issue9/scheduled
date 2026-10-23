// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package every

import (
	"slices"
	"time"
)

// 计算 value 的下一次调度时间
//
// years、months、weekdays、days、hours、minutes 和 seconds 都是升序的。
func next(
	value time.Time, years []int, months []time.Month, weekdays []time.Weekday, days, hours, minutes, seconds []int,
) time.Time {
	second, minSecond, carrySecond := nextValue(seconds, secondBound, true, value.Second())
	minute, minMinute, carryMinute := nextValue(minutes, minuteBound, carrySecond, value.Minute())
	if carryMinute || minute != value.Minute() {
		second = minSecond
	}
	hour, _, carryHour := nextValue(hours, hourBound, carryMinute, value.Hour())
	if carryHour || hour != value.Hour() {
		minute = minMinute
	}

	var (
		year, day int
		month     time.Month
	)
	if len(weekdays) > 0 {
		year, month, day = nextWeekDay(years, months, weekdays, days, value, carryHour)
	} else {
		year, month, day = nextMonthDay(years, months, days, value, carryHour)
	}

	if (year == 0 && month == 0 && day == 0) || year > maxYear {
		return time.Time{}
	}

	return time.Date(year, month, day, hour, minute, second, 0, value.Location())
}

// nextMonthDay 计算 value 的下一次调度时间的年月日部分
//
// years、months 和 days 如果不为空，表示时间只能从这些列表中选择，
// carry 表示之前关于时间部分的内容是否有进位。
//
// 返回值中如果 year 为零，则表示该时间已经超出可计算的范围。
func nextMonthDay(
	years []int, months []time.Month, days []int, value time.Time, carry bool,
) (year int, month time.Month, day int) {
	dayBound := buildMonthDayBound(value.Month(), value.Year())

	day, minDay, carryDay := nextValue(days, dayBound, carry, value.Day())

	// 如果 day 没有进位，那么不需要重新计算 month 和 year，直接使用 value 中对应的值就行。
	if !carryDay &&
		(len(months) == 0 || slices.Contains(months, value.Month())) &&
		(len(years) == 0 || slices.Contains(years, value.Year())) {
		month = value.Month()
		year = value.Year()
	} else {
		var carryMonth bool
		var minMonth time.Month
		month, minMonth, carryMonth = nextValue(months, monthBound, carryDay, value.Month())
		if carryMonth || month != value.Month() {
			day = minDay
		}

		// 如果 month 没有进位，那么不需要重新计算 year，直接使用 value.Year 就行。
		if !carryMonth && (len(years) == 0 || slices.Contains(years, value.Year())) {
			year = value.Year()
		} else {
			var carryYear bool
			year, _, carryYear = nextValue(years, yearBound, carryMonth, value.Year())
			if carryYear { // 已经没有可用的年份了
				return 0, 0, 0
			}
			if year != value.Year() {
				month = minMonth
				day = minDay
			}
		}
	}

	// 由于月份中的天数不固定，还得计算该天数是否存在于当前月分，
	// 比如 days=[30,31]，碰到 4 月份的话，nextMonthDay 就需要跳过 31 号，并进入计算下个月，
	// 如果计算出来的下一个月是下一年的 2 月份，则 30 号依然不符合要求，需要再次计算。
	for {
		maxDays := getMonthDays(month, year)
		if day <= maxDays { // 天数存在于当前月，则返回
			return year, month, day
		}

		// 天数不存在于当前月，则需要进位到下一个月，day 要归零
		if len(days) > 0 {
			day = days[0]
		} else {
			day = 1
		}

		var minMonth time.Month
		var carryMonth, carryYear bool
		month, minMonth, carryMonth = nextValue(months, monthBound, true, month)
		if carryMonth { // 如果计算月份产生了进位，还需要重新计算年份。
			year, _, carryYear = nextValue(years, yearBound, carryMonth, year)
			if year != value.Year() || carryYear {
				month = minMonth
			}

			if carryYear { // 如果出来的年份还需要进位，则表示已经起出范围了，直接返回零值。
				return 0, 0, 0
			}
		}
	}
}

// nextWeekDay 计算 value 的下一次调度时间的年月日部分
//
// 与 nextMonthDay 的不同点在于此方法会同时考虑 days 和 weekdays 两个条件，
// 若同时指定了两者，则取最近的那个值时间点。
func nextWeekDay(
	years []int, months []time.Month, weekdays []time.Weekday, days []int, value time.Time, carry bool,
) (year int, month time.Month, day int) {
	// 计算 week day 在当前月份中的日期，getMonthWeekDay 是获取第一个相符的日期，与此功能稍有不同
	wday, _, ca := nextValue(weekdays, weekdayBound, carry, value.Weekday())
	dur := wday - value.Weekday() // 相差的天数
	if (dur < 0) || (ca && dur == 0) {
		dur += 7
	}
	day = int(dur) + int(value.Day()) // wday 在当前月对应的天数

	month, minMonth, carryMonth := nextValue(months, monthBound, false, value.Month())
	year, _, carryYear := nextValue(years, yearBound, carryMonth, value.Year())
	weekdayValid := !carryYear
	if weekdayValid {
		if year != value.Year() { // 年份改了，重置 month
			month = minMonth
		}

		// month 或 year 发生变化，要重新计算 day
		if month != value.Month() || year != value.Year() {
			day = firstWeekDay(weekdays, month, year)
		}

		// day 不存在于当前月份，比如 30 号不存在于 2 月分。
		for day > getMonthDays(month, year) {
			// 跨月份，还有可能跨年份
			month, minMonth, carryMonth = nextValue(months, monthBound, true, month)
			year, _, carryYear = nextValue(years, yearBound, carryMonth, year)
			if carryYear { // years 列表回到首页，表示没有新的年份可用了
				weekdayValid = false
				break
			} else if year != value.Year() { // 年份改了，重置 month
				month = minMonth
			}

			day = firstWeekDay(weekdays, month, year)
		}
	}

	// 同时设置了 days，需要比较两个值哪个更近
	if len(days) > 0 {
		y, m, d := nextMonthDay(years, months, days, value, carry)
		if y != 0 || m != 0 || d != 0 {
			if !weekdayValid || y < year || (y == year && m < month) || (y == year && m == month && d < day) {
				return y, m, d
			}
		}
	}

	if !weekdayValid {
		return 0, 0, 0
	}

	return year, month, day
}

// 从 values 或是 b 中查找 value 的下一个值
//
// carry 表示 value 的值至少需要加上 1 再从 values 或是 b 中查找匹配的下一值；
// b 只有在 values 为空时才有用。
//
// 返回找到的值以及是否已经进行进位的标记，如果已经进位，那么返回的值应该是 values[0] 或是 b.min。
func nextValue[T integer](values []T, b bound[T], carry bool, value T) (v, min T, c bool) {
	if carry { // 前一个值有进位
		value += 1
	}

	if len(values) > 0 {
		for _, v := range values {
			// carry 为 true，表示前面已经执行了 value+1，
			// 那么 v 可以等于 value，否则 v 必须大于 value。
			if v >= value {
				return v, values[0], false
			}
		}

		// 如果 values 已经没有符合要求的值了，则拿 values[0] 并返回该值已经进位的标记。
		return values[0], values[0], true
	}

	// 未指定 values，则从参数 b 中依次查找可用的值

	if (b.max == -1) || (value <= b.max) { // 不存在 max 或是 value 还未超出 max
		return value, b.min, false
	}

	return b.min, b.min, true

}

// 生成某个月份的日期取值范围
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

func firstWeekDay(weekdays []time.Weekday, month time.Month, year int) int {
	day := 0
	for _, w := range weekdays {
		if d := getMonthWeekDay(w, month, year); day == 0 || d < day {
			day = d
		}
	}
	return day
}
