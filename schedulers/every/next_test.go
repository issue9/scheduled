// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package every

import (
	"testing"
	"time"

	"github.com/issue9/assert/v5"
)

func TestNext(t *testing.T) {
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

		want string
	}{
		{
			name: "下一秒",

			value: "2024-05-01T12:34:56Z",
			want:  "2024-05-01T12:34:57Z",
		},
		{
			name: "仅限制秒",

			value:   "2024-05-01T12:34:56Z",
			seconds: []int{0, 15, 30, 45},

			want: "2024-05-01T12:35:00Z",
		},
		{
			name: "年月日",

			value:   "2024-01-31T00:00:00Z",
			years:   []int{2024, 2028},
			months:  []time.Month{time.February},
			days:    []int{29},
			hours:   []int{0},
			minutes: []int{0},
			seconds: []int{0},

			want: "2024-02-29T00:00:00Z",
		},
		{
			name: "时分秒级联进位",

			value:   "2024-05-01T12:34:56Z",
			hours:   []int{13, 15},
			minutes: []int{0, 30},
			seconds: []int{0, 15},

			want: "2024-05-01T13:00:00Z",
		},
		{
			name: "月底进位",

			value: "2024-05-31T23:59:59Z",
			want:  "2024-06-01T00:00:00Z",
		},
		{
			name: "星期及时间约束",

			value:    "2024-05-05T10:30:00Z",
			weekdays: []time.Weekday{time.Monday},
			hours:    []int{9},
			minutes:  []int{0},
			seconds:  []int{0},

			want: "2024-05-06T09:00:00Z",
		},
		{
			name: "月份和日期约束",

			value:  "2024-05-10T12:00:00Z",
			months: []time.Month{time.June},
			days:   []int{15},

			want: "2024-06-15T12:00:01Z",
		},
		{
			name: "秒未进位，分钟保持不变",

			value:   "2024-05-01T12:00:05Z",
			minutes: []int{0, 30},
			seconds: []int{10},

			want: "2024-05-01T12:00:10Z",
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			a := assert.New(t, false)
			value, err := time.Parse(time.RFC3339, item.value)
			a.NotError(err)
			want, err := time.Parse(time.RFC3339, item.want)
			a.NotError(err)

			got := next(value, item.years, item.months, item.weekdays, item.days, item.hours, item.minutes, item.seconds)
			a.Equal(want, got)
		})
	}
}

func TestNextMonthDay(t *testing.T) {
	data := []struct {
		name string

		// 输入

		value  string
		years  []int
		months []time.Month
		days   []int
		carry  bool

		// 输出

		year  int
		month time.Month
		day   int
	}{
		{
			name: "!carry day+1 month+1",

			value:  "2024-01-31T00:00:00Z",
			years:  []int{2024},
			months: []time.Month{time.January, time.February, time.March},
			days:   []int{31},

			year:  2024,
			month: time.January,
			day:   31,
		},
		{
			name: "!carry day+1",

			value:  "2024-01-15T00:00:00Z",
			years:  []int{2024},
			months: []time.Month{time.January, time.February},
			days:   []int{15, 20, 31},

			year:  2024,
			month: time.January,
			day:   15,
		},
		{
			name: "闰年 2-29",

			value:  "2024-01-31T00:00:00Z",
			years:  []int{2024, 2028},
			months: []time.Month{time.February},
			days:   []int{29},

			year:  2024,
			month: time.February,
			day:   29,
		},
		{
			name: "非闰年跳过 2-29",

			value:  "2023-01-31T00:00:00Z",
			years:  []int{2023},
			months: []time.Month{time.February, time.March},
			days:   []int{29},

			year:  2023,
			month: time.March,
			day:   29,
		},
		{
			name: "carry year+1",

			value:  "2024-12-31T00:00:00Z",
			years:  []int{2024, 2025},
			months: []time.Month{time.January},
			days:   []int{31},
			carry:  true,

			year:  2025,
			month: time.January,
			day:   31,
		},
		{
			name: "月份前进时日期从头计算",

			value:  "2024-05-01T00:00:00Z",
			months: []time.Month{time.December},
			days:   []int{1, 15},

			year:  2024,
			month: time.December,
			day:   1,
		},
		{
			name: "年份约束",

			value: "2024-05-01T00:00:00Z",
			years: []int{2025},

			year:  2025,
			month: time.January,
			day:   1,
		},
		{
			name: "年份和日期约束",

			value: "2024-05-10T00:00:00Z",
			years: []int{2025},
			days:  []int{1, 15},

			year:  2025,
			month: time.January,
			day:   1,
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			a := assert.New(t, false)
			val, err := time.Parse(time.RFC3339, item.value)
			a.NotError(err)
			year, month, day := nextMonthDay(item.years, item.months, item.days, val, item.carry)
			a.Equal(item.year, year).Equal(item.month, month).Equal(item.day, day)
		})
	}
}

func TestNextWeekDay(t *testing.T) {
	data := []struct {
		name string

		value    string
		years    []int
		months   []time.Month
		weekdays []time.Weekday
		days     []int
		carry    bool

		year  int
		month time.Month
		day   int
	}{
		{
			name: "限定星期",

			value:    "2024-05-01T00:00:00Z",
			weekdays: []time.Weekday{time.Friday},

			year:  2024,
			month: time.May,
			day:   3,
		},
		{
			name: "同一天",

			value:    "2024-05-01T00:00:00Z",
			weekdays: []time.Weekday{time.Wednesday},

			year:  2024,
			month: time.May,
			day:   1,
		},
		{
			name: "星期跨月",

			value:    "2024-05-31T00:00:00Z",
			weekdays: []time.Weekday{time.Monday},

			year:  2024,
			month: time.June,
			day:   3,
		},
		{
			name: "星期跨年",

			value:    "2024-12-31T00:00:00Z",
			weekdays: []time.Weekday{time.Thursday},

			year:  2025,
			month: time.January,
			day:   2,
		},
		{
			name: "同时拥有月份和星期",

			value:    "2024-05-10T00:00:00Z",
			months:   []time.Month{time.June, time.August},
			weekdays: []time.Weekday{time.Monday},

			year:  2024,
			month: time.June,
			day:   3,
		},
		{
			name: "跳月后取最早的星期",

			value:    "2024-05-01T00:00:00Z",
			months:   []time.Month{time.June},
			weekdays: []time.Weekday{time.Monday, time.Friday},

			year:  2024,
			month: time.June,
			day:   3,
		},
		{
			name: "同时拥有月份星期及日期",

			value:    "2024-05-10T00:00:00Z",
			months:   []time.Month{time.June, time.August},
			weekdays: []time.Weekday{time.Monday},
			days:     []int{10, 11},

			year:  2024,
			month: time.June,
			day:   3,
		},
		{
			name: "星期早于指定日期",

			value:    "2024-05-01T00:00:00Z",
			weekdays: []time.Weekday{time.Friday},
			days:     []int{4},

			year:  2024,
			month: time.May,
			day:   3,
		},
		{
			name: "星期晚于指定日期",

			value:    "2024-05-01T00:00:00Z",
			weekdays: []time.Weekday{time.Friday},
			days:     []int{2},

			year:  2024,
			month: time.May,
			day:   2,
		},
		{
			name: "星期受年份约束",

			value:    "2024-05-01T00:00:00Z",
			years:    []int{2025},
			weekdays: []time.Weekday{time.Monday},

			year:  2025,
			month: time.January,
			day:   6,
		},
		{
			name: "跨月时星期与日期取更早的",

			value:    "2024-05-30T00:00:00Z",
			weekdays: []time.Weekday{time.Monday},
			days:     []int{31},

			year:  2024,
			month: time.May,
			day:   31,
		},
		{
			name: "days 无匹配时仍使用星期",

			value:    "2024-01-01T00:00:00Z",
			months:   []time.Month{time.February},
			weekdays: []time.Weekday{time.Monday},
			days:     []int{31},

			year:  2024,
			month: time.February,
			day:   5,
		},
		{
			name: "星期年份耗尽时仍使用日期",

			value:    "2024-12-30T00:00:00Z",
			years:    []int{2024},
			weekdays: []time.Weekday{time.Monday},
			days:     []int{31},
			carry:    true,

			year:  2024,
			month: time.December,
			day:   31,
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			a := assert.New(t, false)
			value, err := time.Parse(time.RFC3339, item.value)
			a.NotError(err)
			year, month, day := nextWeekDay(item.years, item.months, item.weekdays, item.days, value, item.carry)
			a.Equal(item.year, year).Equal(item.month, month).Equal(item.day, day)
		})
	}
}

func TestNextValue(t *testing.T) {
	// 注意：carry 表示更低位的时间是否发生了进位。
	// carry 为 false 时，当前值本身可以是合法值；只有 carry 为 true 时，才必须严格大于当前值。

	t.Run("values+!carry", func(t *testing.T) {
		a := assert.New(t, false)

		val, _, carry := nextValue([]int{1, 2, 3}, bound[int]{}, false, 1)
		a.False(carry).Equal(val, 1)

		// 当前值小于列表的最小值，返回最小值
		val, _, carry = nextValue([]int{1, 2, 3}, bound[int]{}, false, 0)
		a.False(carry).Equal(val, 1)

		// 当前值是列表的最大值，且不需要进位，不能回绕
		val, _, carry = nextValue([]int{1, 2, 3}, bound[int]{}, false, 3)
		a.False(carry).Equal(val, 3)

		// 当前值大于列表的最大值，才回绕并进位
		val, _, carry = nextValue([]int{1, 2, 3}, bound[int]{}, false, 4)
		a.True(carry).Equal(val, 1)
	})

	t.Run("values+carry", func(t *testing.T) {
		a := assert.New(t, false)

		val, _, carry := nextValue([]int{1, 2, 3}, bound[int]{}, true, 2)
		a.False(carry).Equal(val, 3)

		val, _, carry = nextValue([]int{1, 2, 3}, bound[int]{}, true, 1)
		a.False(carry).Equal(val, 2)

		val, _, carry = nextValue([]int{1, 2, 3}, bound[int]{}, true, 3)
		a.True(carry).Equal(val, 1)
	})

	t.Run("bound+!carry", func(t *testing.T) {
		a := assert.New(t, false)

		// 未进位：允许返回当前值
		wday, _, carry := nextValue(nil, weekdayBound, false, time.Wednesday)
		a.False(carry).Equal(wday, time.Wednesday)

		val, _, carry := nextValue(nil, minuteBound, false, 30)
		a.False(carry).Equal(val, 30)

		// 当前值已经是上限，且不需要进位，不能回绕
		val, _, carry = nextValue(nil, minuteBound, false, 59)
		a.False(carry).Equal(val, 59)

		// 超出上限才回绕并进位
		val, _, carry = nextValue(nil, minuteBound, false, 60)
		a.True(carry).Equal(val, 0)
	})

	t.Run("bound+carry", func(t *testing.T) {
		a := assert.New(t, false)

		wday, _, carry := nextValue(nil, weekdayBound, true, time.Wednesday)
		a.False(carry).Equal(wday, time.Thursday)

		val, _, carry := nextValue(nil, minuteBound, true, 30)
		a.False(carry).Equal(val, 31)

		wday, _, carry = nextValue(nil, weekdayBound, true, time.Saturday)
		a.True(carry).Equal(wday, time.Sunday)

		val, _, carry = nextValue(nil, minuteBound, true, 59)
		a.True(carry).Equal(val, 0)
	})

	t.Run("bound max=-1", func(t *testing.T) {
		a := assert.New(t, false)

		// max 为 -1 表示没有上限
		val, _, carry := nextValue(nil, bound[int]{min: 1, max: -1}, false, 100)
		a.False(carry).Equal(val, 100)

		val, _, carry = nextValue(nil, bound[int]{min: 1, max: -1}, true, 100)
		a.False(carry).Equal(val, 101)

		val, _, carry = nextValue(nil, yearBound, true, maxYear)
		a.True(carry).Equal(val, minYear)
	})
}

func TestGetMonthDays(t *testing.T) {
	a := assert.New(t, false)

	var (
		leapDays = map[int]int{ // 闰年对应的天数
			1: 31, 3: 31, 5: 31, 7: 31, 8: 31, 10: 31, 12: 31,
			2: 29, 4: 30, 6: 30, 9: 30, 11: 30,
		}
		days = map[int]int{
			1: 31, 3: 31, 5: 31, 7: 31, 8: 31, 10: 31, 12: 31,
			2: 28, 4: 30, 6: 30, 9: 30, 11: 30,
		}
	)

	// 非闰年
	for k, v := range days {
		a.Equal(v, getMonthDays(time.Month(k), 2019))
	}

	// 闰年：2020
	for k, v := range leapDays {
		a.Equal(v, getMonthDays(time.Month(k), 2020))
	}
}

func TestGetMonthWeekDay(t *testing.T) {
	type test struct {
		name string
		// 输入值
		year    int
		month   time.Month
		weekday time.Weekday

		// 返回值
		day uint8
	}

	data := []*test{
		{
			name:    "2019-05 月的第一个周三是几号",
			year:    2019,
			month:   time.May,
			weekday: time.Wednesday,
			day:     1,
		},
		{
			name:    "2019-05 月的第一个周六是几号",
			year:    2019,
			month:   time.May,
			weekday: time.Saturday,
			day:     4,
		},
		{
			name:    "2019-05 月的第一个周日是几号",
			year:    2019,
			month:   time.May,
			weekday: time.Sunday,
			day:     5,
		},
		{
			name:    "2020-02 月的第一个周六是几号",
			year:    2020,
			month:   time.February,
			weekday: time.Saturday,
			day:     1,
		},
		{
			name:    "2020-02 月的第一个周二是几号",
			year:    2020,
			month:   time.February,
			weekday: time.Tuesday,
			day:     4,
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			a := assert.New(t, false)
			day := getMonthWeekDay(item.weekday, item.month, item.year)
			a.Equal(day, item.day)
		})
	}
}
