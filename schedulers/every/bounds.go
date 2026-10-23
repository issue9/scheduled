// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package every

import "time"

type integer interface{ ~int }

// 用于描述一个类型值的取值范围
type bound[T integer] struct {
	min T
	max T // 如果是 -1 表示没有上限
}

var (
	minYear = time.Now().Year()
	maxYear = minYear + 100
)

var (
	secondBound  = bound[int]{min: 0, max: 59}
	minuteBound  = bound[int]{min: 0, max: 59}
	hourBound    = bound[int]{min: 0, max: 23}
	weekdayBound = bound[time.Weekday]{min: time.Sunday, max: time.Saturday}
	monthBound   = bound[time.Month]{min: time.January, max: time.December}
	yearBound    = bound[int]{min: minYear, max: maxYear}
	dayBound     = bound[int]{min: 1, max: 31} // 每个月的日期范围
)

func buildMonthDayBound(month time.Month, year int) bound[int] {
	return bound[int]{min: 1, max: getMonthDays(month, year)}
}

func (b bound[T]) valid(v int) bool {
	t := T(v)
	return t >= b.min && (b.max == -1 || t <= b.max)
}
