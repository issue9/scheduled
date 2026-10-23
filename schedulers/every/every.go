// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

// Package every 可定制的定时任务
package every

import (
	"time"

	"github.com/issue9/scheduled/schedulers"
)

// Every 创建一个调度器
//
// 参数限定了 [schedulers.Scheduler.Next] 返回值的可用范围，
// 比如 months=[time.January, time.March]，那么 Next 返回的时间只能是一月或是三月的。
// 如果为空值，则会使用该类型元素的任意值，比如 hours 为空，实际上相当于该参数为 [0-59] 的切片列表。
//
// weekdays 与 days 同时非空时，只要满足其中之一即可，最终取更早的时间点。
//
// 参数中的所有切片必须按升序排列。
func Every(
	years []int, months []time.Month, weekdays []time.Weekday, days, hours, minutes, seconds []int,
) schedulers.Scheduler {
	return schedulers.SchedulerFunc(func(t time.Time) time.Time {
		return next(t, years, months, weekdays, days, hours, minutes, seconds)
	})
}

// Yearly 与 [Every] 相同但是 years 不允许自定义
func Yearly(
	months []time.Month, weekdays []time.Weekday, days, hours, minutes, seconds []int,
) schedulers.Scheduler {
	return schedulers.SchedulerFunc(func(t time.Time) time.Time {
		return next(t, nil, months, weekdays, days, hours, minutes, seconds)
	})
}

// Hourly 创建一个只可自定义分和秒的调度器
//
// 这是一个特殊的 [Every]，只关注小时以内的定义，其它参数均为默认值。
func Hourly(minutes, seconds []int) schedulers.Scheduler {
	return schedulers.SchedulerFunc(func(t time.Time) time.Time {
		return next(t, nil, nil, nil, nil, nil, minutes, seconds)
	})
}

func Daily(hours, minutes, seconds []int) schedulers.Scheduler {
	return schedulers.SchedulerFunc(func(t time.Time) time.Time {
		return next(t, nil, nil, nil, nil, hours, minutes, seconds)
	})
}

func Monthly(days, hours, minutes, seconds []int) schedulers.Scheduler {
	return schedulers.SchedulerFunc(func(t time.Time) time.Time {
		return next(t, nil, nil, nil, days, hours, minutes, seconds)
	})
}

func Weekly(weekdays []time.Weekday, hours, minutes, seconds []int) schedulers.Scheduler {
	return schedulers.SchedulerFunc(func(t time.Time) time.Time {
		return next(t, nil, nil, weekdays, nil, hours, minutes, seconds)
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

// AM 创建只运行于下午的调度器
func PM(minutes, seconds []int) schedulers.Scheduler {
	return Daily(hours[12:], minutes, seconds)
}
