// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

// Package every 可定制的定时任务
package every

import (
	"slices"
	"time"

	"github.com/issue9/scheduled/schedulers"
)

// Every 创建一个调度器
//
// 参数限定了 [schedulers.Scheduler.Next] 返回值的可用范围，
// 比如 months=[time.January, time.March]，那么 Next 返回的时间只能是一月或是三月的。
// 如果为空值，则会使用该类型元素的任意值，比如 hours 为空，实际上相当于该参数为 [0-23] 的切片列表。
//
// weekdays 与 days 同时非空时，只要满足其中之一即可，最终取更早的时间点。
//
// 参数中的所有切片必须按升序排列，且不能包含重复值、不能超出各自的取值范围
// （比如 hours 只能是 0-23），否则会直接 panic。
func Every(
	years []int, months []time.Month, weekdays []time.Weekday, days, hours, minutes, seconds []int,
) schedulers.Scheduler {
	return newScheduler(years, months, weekdays, days, hours, minutes, seconds)
}

// Yearly 创建在每年固定时间执行的高度器
func Yearly(
	months []time.Month, weekdays []time.Weekday, days, hours, minutes, seconds []int,
) schedulers.Scheduler {
	return newScheduler(nil, months, weekdays, days, hours, minutes, seconds)
}

// Hourly 创建在每个小时之内固定时间执行的高度器
//
// 这是一个特殊的 [Every]，只关注小时以内的定义，其它参数均为默认值。
func Hourly(minutes, seconds []int) schedulers.Scheduler {
	return newScheduler(nil, nil, nil, nil, nil, minutes, seconds)
}

// Daily 创建在每天固定时间执行的调度器
func Daily(hours, minutes, seconds []int) schedulers.Scheduler {
	return newScheduler(nil, nil, nil, nil, hours, minutes, seconds)
}

// Monthly 创建在每个月按固定时间执行的调度器
func Monthly(days, hours, minutes, seconds []int) schedulers.Scheduler {
	return newScheduler(nil, nil, nil, days, hours, minutes, seconds)
}

// EndOfMonthly 创建一个按月末日期运行的调度器
//
// 功能与 [Monthly] 相似，但是 days 改为从月末开始计算，比如 days 为 [0,2]，
// 表示每个月最后一天和倒数第三天的日期。days 的取值范围为 [0,30]，为空时匹配每一天；
// 其它参数与 [Monthly] 相同。
func EndOfMonthly(days, hours, minutes, seconds []int) schedulers.Scheduler {
	validate("days", days, bound[int]{min: 0, max: 30})
	validate("hours", hours, hourBound)
	validate("minutes", minutes, minuteBound)
	validate("seconds", seconds, secondBound)

	return schedulers.SchedulerFunc(func(last time.Time) time.Time {
		if last.Year() > maxYear {
			return time.Time{}
		}

		for year := last.Year(); year <= maxYear; year++ {
			firstMonth := time.January
			if year == last.Year() {
				firstMonth = last.Month()
			}

			for month := firstMonth; month <= time.December; month++ {
				monthDays := getMonthDays(month, year)
				actualDays := make([]int, 0, len(days))
				for _, day := range slices.Backward(days) {
					day := monthDays - day
					if day > 0 {
						actualDays = append(actualDays, day)
					}
				}

				if len(days) > 0 && len(actualDays) == 0 {
					continue
				}

				if next := next(last, []int{year}, []time.Month{month}, nil, actualDays, hours, minutes, seconds); !next.IsZero() {
					return next
				}
			}
		}

		return time.Time{}
	})
}

// Monthly 创建在每周按固定时间执行的调度器
func Weekly(weekdays []time.Weekday, hours, minutes, seconds []int) schedulers.Scheduler {
	return newScheduler(nil, nil, weekdays, nil, hours, minutes, seconds)
}

// newScheduler 校验参数并生成调度器
//
// 参数必须按升序排列，且不能超出各自的取值范围，否则会直接 panic；
func newScheduler(
	years []int, months []time.Month, weekdays []time.Weekday, days, hours, minutes, seconds []int,
) schedulers.Scheduler {
	validateSorted("years", years) // years 的上限由 maxYear 控制，超出时由 next 返回零值，所以此处不做范围检查。
	validate("months", months, monthBound)
	validate("weekdays", weekdays, weekdayBound)
	validate("days", days, dayBound)
	validate("hours", hours, hourBound)
	validate("minutes", minutes, minuteBound)
	validate("seconds", seconds, secondBound)

	return schedulers.SchedulerFunc(func(t time.Time) time.Time {
		return next(t, years, months, weekdays, days, hours, minutes, seconds)
	})
}

var weekdays = []time.Weekday{
	time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday,
	time.Sunday, time.Saturday, // 注意排序
}

// Weekday 创建一个只运行于工作日的调度器
func Weekday(hours, minutes, seconds []int) schedulers.Scheduler {
	return Weekly(weekdays[:5], hours, minutes, seconds)
}

// Weekend 创建一个只运行于周末的调度器
func Weekend(hours, minutes, seconds []int) schedulers.Scheduler {
	return Weekly(weekdays[5:], hours, minutes, seconds)
}

var hours = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23}

// AM 创建只运行于上午的调度器
func AM(minutes, seconds []int) schedulers.Scheduler {
	return Daily(hours[:12], minutes, seconds)
}

// PM 创建只运行于下午的调度器
func PM(minutes, seconds []int) schedulers.Scheduler {
	return Daily(hours[12:], minutes, seconds)
}
