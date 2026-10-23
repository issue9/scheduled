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
