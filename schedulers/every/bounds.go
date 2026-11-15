// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package every

import (
	"fmt"
	"slices"
	"time"

	"github.com/issue9/sliceutil"
)

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
	secondBound      = bound[int]{min: 0, max: 59}
	minuteBound      = bound[int]{min: 0, max: 59}
	hourBound        = bound[int]{min: 0, max: 23}
	weekdayBound     = bound[time.Weekday]{min: time.Sunday, max: time.Saturday}
	cronWeekdayBound = bound[time.Weekday]{min: time.Sunday, max: 7} // cron 中支持用 7 表示星期日
	monthBound       = bound[time.Month]{min: time.January, max: time.December}
	yearBound        = bound[int]{min: minYear, max: maxYear}
	dayBound         = bound[int]{min: 1, max: 31} // 每个月的日期范围
)

func buildMonthDayBound(month time.Month, year int) bound[int] {
	return bound[int]{min: 1, max: getMonthDays(month, year)}
}

func (b bound[T]) valid(v int) bool {
	t := T(v)
	return t >= b.min && (b.max == -1 || t <= b.max)
}

// validate 检查 values 是否按升序排列，且元素都在 b 的范围内，不满足时直接 panic。
//
// name 仅用于生成 panic 信息。
func validate[T integer](name string, values []T, b bound[T]) {
	validateSorted(name, values)

	for _, v := range values {
		if !b.valid(int(v)) {
			panic(fmt.Sprintf("参数 %s 的值 %d 超出范围 [%d,%d]", name, int(v), int(b.min), int(b.max)))
		}
	}
}

// validateSorted 检查 values 是否按升序排列且没有重复值，不满足时直接 panic。
func validateSorted[T integer](name string, values []T) {
	if !slices.IsSorted(values) {
		panic(fmt.Sprintf("参数 %s 的值必须按升序排列: %v", name, values))
	}

	if indexes := sliceutil.Dup(values, func(i, j T) bool { return i == j }); len(indexes) > 0 {
		panic(fmt.Sprintf("参数 %s 的值存在重复项: %d", name, values[indexes[0]]))
	}
}
