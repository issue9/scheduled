// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package every

import (
	"testing"
	"time"

	"github.com/issue9/assert/v6"

	"github.com/issue9/scheduled/schedulers"
)

func TestEvery(t *testing.T) {
	data := []struct {
		name string

		value    string
		years    []int
		months   []time.Month
		weekdays []time.Weekday
		days     []int
		hours    []int
		minutes  []int
		seconds  []int

		want string // 空表示期望返回零值
	}{
		{
			name: "星期",

			value:    "2024-05-01T00:00:00Z",
			weekdays: []time.Weekday{time.Monday},
			hours:    []int{0},
			minutes:  []int{0},
			seconds:  []int{0},

			want: "2024-05-06T00:00:00Z",
		},
		{
			name: "没有匹配的时间",

			value:   "2024-01-31T00:00:00Z",
			years:   []int{2024},
			months:  []time.Month{time.February},
			days:    []int{31},
			hours:   []int{0},
			minutes: []int{0},
			seconds: []int{0},

			want: "",
		},
		{
			name: "年份超出上限",

			value: time.Date(maxYear, time.December, 31, 23, 59, 59, 0, time.UTC).Format(time.RFC3339),

			want: "",
		},
		{
			name: "value 超过年份上限",

			value: time.Date(maxYear+1, time.January, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),

			want: "",
		},
		{
			name: "years 超过年份上限",

			value:   time.Date(maxYear, time.January, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			years:   []int{maxYear + 1},
			hours:   []int{0},
			minutes: []int{0},
			seconds: []int{0},

			want: "",
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			s := Every(item.years, item.months, item.weekdays, item.days,
				item.hours, item.minutes, item.seconds)
			checkNext(t, s, item.value, item.want)
		})
	}
}

func TestHourly(t *testing.T) {
	data := []struct {
		name string

		value            string
		minutes, seconds []int

		want string
	}{
		{
			name: "秒进位，分钟跳到下一项",

			value:   "2024-05-01T12:05:10Z",
			minutes: []int{0, 30},
			seconds: []int{0},

			want: "2024-05-01T12:30:00Z",
		},
		{
			name: "分钟进位到下一小时",

			value:   "2024-05-01T12:30:00Z",
			minutes: []int{0, 30},
			seconds: []int{0},

			want: "2024-05-01T13:00:00Z",
		},
		{
			name: "不限制分钟和秒",

			value: "2024-05-01T12:00:00Z",

			want: "2024-05-01T12:00:01Z",
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			checkNext(t, Hourly(item.minutes, item.seconds), item.value, item.want)
		})
	}
}

func TestDaily(t *testing.T) {
	data := []struct {
		name string

		value                   string
		hours, minutes, seconds []int

		want string
	}{
		{
			name: "当天稍后",

			value:   "2024-05-01T10:00:00Z",
			hours:   []int{9, 18},
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-01T18:00:00Z",
		},
		{
			name: "次日",

			value:   "2024-05-01T18:00:00Z",
			hours:   []int{9, 18},
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-02T09:00:00Z",
		},
		{
			name: "跨月",

			value:   "2024-05-31T23:59:59Z",
			hours:   []int{0},
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-06-01T00:00:00Z",
		},
		{
			name: "秒未进位，分钟保持不变",

			value:   "2024-05-01T09:00:10Z",
			hours:   []int{9, 18},
			minutes: []int{0},
			seconds: []int{30},

			want: "2024-05-01T09:00:30Z",
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			checkNext(t, Daily(item.hours, item.minutes, item.seconds), item.value, item.want)
		})
	}
}

func TestMonthly(t *testing.T) {
	data := []struct {
		name string

		value                   string
		days                    []int
		hours, minutes, seconds []int

		want string
	}{
		{
			name: "当月稍后",

			value:   "2024-05-10T00:00:00Z",
			days:    []int{1, 15},
			hours:   []int{0},
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-15T00:00:00Z",
		},
		{
			name: "下个月",

			value:   "2024-05-15T00:00:00Z",
			days:    []int{1, 15},
			hours:   []int{0},
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-06-01T00:00:00Z",
		},
		{
			name: "跨年",

			value:   "2024-12-01T00:00:00Z",
			days:    []int{1},
			hours:   []int{0},
			minutes: []int{0},
			seconds: []int{0},

			want: "2025-01-01T00:00:00Z",
		},
		{
			name: "跳过没有该日期的月份",

			value:   "2024-05-31T00:00:00Z",
			days:    []int{31},
			hours:   []int{0},
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-07-31T00:00:00Z",
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			checkNext(t, Monthly(item.days, item.hours, item.minutes, item.seconds), item.value, item.want)
		})
	}
}

func TestEndOfMonthly(t *testing.T) {
	data := []struct {
		name string

		value                   string
		days                    []int
		hours, minutes, seconds []int

		want string
	}{
		{
			name:    "月末",
			value:   "2024-05-30T12:00:00Z",
			days:    []int{0},
			hours:   []int{0},
			minutes: []int{0},
			seconds: []int{0},
			want:    "2024-05-31T00:00:00Z",
		},
		{
			name:    "倒数第三天",
			value:   "2024-05-28T00:00:00Z",
			days:    []int{2},
			hours:   []int{0},
			minutes: []int{0},
			seconds: []int{0},
			want:    "2024-05-29T00:00:00Z",
		},
		{
			name:    "指定日期时间已过",
			value:   "2024-05-31T10:00:00Z",
			days:    []int{0},
			hours:   []int{9},
			minutes: []int{0},
			seconds: []int{0},
			want:    "2024-06-30T09:00:00Z",
		},
		{
			name:    "多个日期和时间",
			value:   "2024-05-29T09:30:30Z",
			days:    []int{0, 2},
			hours:   []int{9, 15},
			minutes: []int{0, 30},
			seconds: []int{0, 30},
			want:    "2024-05-29T15:00:00Z",
		},
		{
			name:    "多个日期和时间均已过",
			value:   "2024-05-29T15:30:30Z",
			days:    []int{0, 2},
			hours:   []int{9, 15},
			minutes: []int{0, 30},
			seconds: []int{0, 30},
			want:    "2024-05-31T09:00:00Z",
		},
		{
			name:    "当月不存在偏移日期",
			value:   "2024-02-28T00:00:00Z",
			days:    []int{30},
			hours:   []int{0},
			minutes: []int{0},
			seconds: []int{0},
			want:    "2024-03-01T00:00:00Z",
		},
		{
			name:  "未指定日期",
			value: "2024-05-15T00:00:00Z",
			want:  "2024-05-15T00:00:01Z",
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			checkNext(t, EndOfMonthly(item.days, item.hours, item.minutes, item.seconds), item.value, item.want)
		})
	}

	a := assert.New(t, false)
	a.Panic(func() { EndOfMonthly([]int{-1}, nil, nil, nil) })
	a.Panic(func() { EndOfMonthly([]int{31}, nil, nil, nil) })
	a.NotPanic(func() { EndOfMonthly([]int{0, 30}, nil, nil, nil) })
}

func TestWeekly(t *testing.T) {
	data := []struct {
		name string

		value                   string
		weekdays                []time.Weekday
		hours, minutes, seconds []int

		want string
	}{
		{
			name: "本周稍后",

			value:    "2024-04-30T10:00:00Z",
			weekdays: []time.Weekday{time.Monday, time.Wednesday, time.Friday},
			hours:    []int{9},
			minutes:  []int{0},
			seconds:  []int{0},

			want: "2024-05-01T09:00:00Z",
		},
		{
			name: "本周下一个",

			value:    "2024-05-01T09:00:00Z",
			weekdays: []time.Weekday{time.Monday, time.Wednesday, time.Friday},
			hours:    []int{9},
			minutes:  []int{0},
			seconds:  []int{0},

			want: "2024-05-03T09:00:00Z",
		},
		{
			name: "跨月",

			value:    "2024-05-31T00:00:00Z",
			weekdays: []time.Weekday{time.Monday},
			hours:    []int{0},
			minutes:  []int{0},
			seconds:  []int{0},

			want: "2024-06-03T00:00:00Z",
		},
		{
			name: "空列表表示每天都执行",

			value:   "2024-05-01T00:00:00Z",
			hours:   []int{0},
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-02T00:00:00Z",
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			checkNext(t, Weekly(item.weekdays, item.hours, item.minutes, item.seconds), item.value, item.want)
		})
	}
}

func TestWeekday(t *testing.T) {
	data := []struct {
		name string

		value                   string
		hours, minutes, seconds []int

		want string
	}{
		{
			name: "周一之后",

			value:   "2024-05-06T10:00:00Z",
			hours:   []int{9},
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-07T09:00:00Z",
		},
		{
			name: "周五之后跳到下周一",

			value:   "2024-05-03T10:00:00Z",
			hours:   []int{9},
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-06T09:00:00Z",
		},
		{
			name: "周日之后跳到周一",

			value:   "2024-05-05T10:00:00Z",
			hours:   []int{9},
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-06T09:00:00Z",
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			checkNext(t, Weekday(item.hours, item.minutes, item.seconds), item.value, item.want)
		})
	}
}

func TestWeekend(t *testing.T) {
	data := []struct {
		name string

		value                   string
		hours, minutes, seconds []int

		want string
	}{
		{
			name: "周五之后跳到周六",

			value:   "2024-05-03T10:00:00Z",
			hours:   []int{9},
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-04T09:00:00Z",
		},
		{
			name: "周六之后跳到周日",

			value:   "2024-05-04T10:00:00Z",
			hours:   []int{9},
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-05T09:00:00Z",
		},
		{
			name: "周日之后跳到下周六",

			value:   "2024-05-05T10:00:00Z",
			hours:   []int{9},
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-11T09:00:00Z",
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			checkNext(t, Weekend(item.hours, item.minutes, item.seconds), item.value, item.want)
		})
	}
}

func TestYearly(t *testing.T) {
	data := []struct {
		name string

		value                   string
		months                  []time.Month
		weekdays                []time.Weekday
		days                    []int
		hours, minutes, seconds []int

		want string
	}{
		{
			name: "闰年 2 月 29 日",

			value:   "2024-03-01T00:00:00Z",
			months:  []time.Month{time.February},
			days:    []int{29},
			hours:   []int{0},
			minutes: []int{0},
			seconds: []int{0},

			want: "2028-02-29T00:00:00Z",
		},
		{
			name: "年内日期",

			value:   "2024-05-01T00:00:00Z",
			months:  []time.Month{time.December},
			days:    []int{25},
			hours:   []int{0},
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-12-25T00:00:00Z",
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			s := Yearly(item.months, item.weekdays, item.days, item.hours, item.minutes, item.seconds)
			checkNext(t, s, item.value, item.want)
		})
	}
}

func TestAM(t *testing.T) {
	data := []struct {
		name string

		value            string
		minutes, seconds []int

		want string
	}{
		{
			name: "当天稍后",

			value:   "2024-05-01T10:00:00Z",
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-01T11:00:00Z",
		},
		{
			name: "次日",

			value:   "2024-05-01T13:00:00Z",
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-02T00:00:00Z",
		},
		{
			name: "上午结束",

			value:   "2024-05-01T11:59:59Z",
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-02T00:00:00Z",
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			checkNext(t, AM(item.minutes, item.seconds), item.value, item.want)
		})
	}
}

func TestPM(t *testing.T) {
	data := []struct {
		name string

		value            string
		minutes, seconds []int

		want string
	}{
		{
			name: "当天稍后",

			value:   "2024-05-01T00:00:00Z",
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-01T12:00:00Z",
		},
		{
			name: "下午内",

			value:   "2024-05-01T12:00:00Z",
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-01T13:00:00Z",
		},
		{
			name: "次日",

			value:   "2024-05-01T23:00:00Z",
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-02T12:00:00Z",
		},
		{
			name: "下午结束",

			value:   "2024-05-01T23:59:59Z",
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-05-02T12:00:00Z",
		},
		{
			name: "秒＋1",

			value: "2024-05-01T23:59:58Z",

			want: "2024-05-01T23:59:59Z",
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			checkNext(t, PM(item.minutes, item.seconds), item.value, item.want)
		})
	}
}

// checkNext 校验 s.Next(value) 的返回值。
//
// want 为空时表示期望返回零值。
func checkNext(t *testing.T, s schedulers.Scheduler, value, want string) {
	t.Helper()
	a := assert.New(t, false)

	v, err := time.Parse(time.RFC3339, value)
	a.NotError(err)
	got := s.Next(v)

	if want == "" {
		a.True(got.IsZero(), "期望零值，实际 %v", got)
		return
	}

	w, err := time.Parse(time.RFC3339, want)
	a.NotError(err).Equal(w, got)
}

// 参数值超出有效范围时应该 panic
func TestEvery_outOfRange(t *testing.T) {
	a := assert.New(t, false)

	// 越界值：每一处都应该 panic
	a.PanicString(func() {
		Daily(nil, []int{60}, nil)
	}, "参数 minutes 的值 60 超出范围 [0,59]")

	a.Panic(func() { Daily([]int{24}, nil, nil) })        // hours 上限
	a.Panic(func() { Daily([]int{-1}, nil, nil) })        // hours 下限
	a.Panic(func() { Daily(nil, []int{-1}, nil) })        // minutes 下限
	a.Panic(func() { Hourly(nil, []int{60}) })            // seconds 上限
	a.Panic(func() { Monthly([]int{0}, nil, nil, nil) })  // days 下限
	a.Panic(func() { Monthly([]int{32}, nil, nil, nil) }) // days 上限

	a.Panic(func() { Yearly([]time.Month{0}, nil, nil, nil, nil, nil) }) // months 下限
	a.PanicString(func() {
		Yearly([]time.Month{13}, nil, nil, nil, nil, nil)
	}, "参数 months 的值 13 超出范围 [1,12]") // months 上限，同时锁定输出为数字

	a.Panic(func() { Weekly([]time.Weekday{-1}, nil, nil, nil) }) // weekdays 下限
	a.Panic(func() { Weekly([]time.Weekday{7}, nil, nil, nil) })  // weekdays 上限

	a.Panic(func() { Every(nil, nil, nil, nil, []int{24}, nil, nil) }) // 通过 Every 传入

	// 边界值都是合法的
	a.NotPanic(func() { Daily([]int{0, 23}, []int{0, 59}, []int{0, 59}) })
	a.NotPanic(func() { Monthly([]int{1, 31}, nil, nil, nil) })
	a.NotPanic(func() { Yearly([]time.Month{time.January, time.December}, nil, nil, nil, nil, nil) })
	a.NotPanic(func() { Weekly([]time.Weekday{time.Sunday, time.Saturday}, nil, nil, nil) })
}

// 未排序
func TestEvery_unsorted(t *testing.T) {
	a := assert.New(t, false)

	a.Panic(func() { Daily([]int{12, 5}, nil, nil) })        // hours
	a.Panic(func() { Daily(nil, []int{30, 10}, nil) })       // minutes
	a.Panic(func() { Hourly(nil, []int{30, 10}) })           // seconds
	a.Panic(func() { Monthly([]int{31, 1}, nil, nil, nil) }) // days

	a.PanicString(func() { // weekdays
		Weekly([]time.Weekday{time.Saturday, time.Sunday}, nil, nil, nil)
	}, "参数 weekdays 的值必须按升序排列")

	a.Panic(func() { // months
		Yearly([]time.Month{time.March, time.January}, nil, nil, nil, nil, nil)
	})

	a.Panic(func() { // years
		Every([]int{minYear + 1, minYear}, nil, nil, nil, nil, nil, nil)
	})

	// 升序且无重复值才是合法的
	a.NotPanic(func() { Daily([]int{5, 12}, nil, nil) })
	a.NotPanic(func() { Daily(nil, []int{10, 30}, nil) })
	a.NotPanic(func() { Every([]int{minYear, minYear + 1}, nil, nil, nil, nil, nil, nil) })
}

// 切片中存在重复值
func TestEvery_duplicate(t *testing.T) {
	a := assert.New(t, false)

	a.Panic(func() { Daily([]int{5, 5}, nil, nil) }) // hours
	a.PanicString(func() {
		Daily(nil, []int{10, 10, 30}, nil)
	}, "参数 minutes 的值存在重复项: 10") // minutes

	a.Panic(func() { Hourly(nil, []int{30, 30}) })                                                // seconds
	a.Panic(func() { Monthly([]int{1, 5, 5, 31}, nil, nil, nil) })                                // days
	a.Panic(func() { Weekly([]time.Weekday{time.Sunday, time.Sunday}, nil, nil, nil) })           // weekdays
	a.Panic(func() { Yearly([]time.Month{time.January, time.January}, nil, nil, nil, nil, nil) }) // months
	a.Panic(func() { Every([]int{minYear, minYear}, nil, nil, nil, nil, nil, nil) })              // years
}
