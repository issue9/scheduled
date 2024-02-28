// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package cron

import (
	"fmt"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
)

func TestCron_Next(t *testing.T) {
	a := assert.New(t, false)

	type test struct {
		loc  *time.Location
		expr string

		// 第一个元素表示起始值，
		// 之后的值均是计算 expr 之后的 next 返回值。
		times []string
	}

	var data = []*test{
		{
			expr: "1 * * * * *",
			times: []string{
				"2019-01-01 00:00:00+00:00",
				"2019-01-01 00:00:01+00:00",
				"2019-01-01 00:01:01+00:00",
				"2019-01-01 00:02:01+00:00",
				"2019-01-01 00:03:01+00:00",
			},
		},
		{
			loc:  time.FixedZone("utc+8", 8*3600),
			expr: "1 * * * * *",
			times: []string{
				"2019-01-01 00:00:00+08:00",
				"2019-01-01 00:00:01+08:00",
				"2019-01-01 01:01:01+09:00", // +09:00
				"2019-01-01 00:02:01+08:00",
				"2019-01-01 00:03:01+08:00",
			},
		},

		{
			expr: "* 1 * * * *",
			times: []string{
				"2019-01-01 00:00:00+00:00",
				"2019-01-01 00:01:00+00:00",
				"2019-01-01 01:01:00+00:00",
				"2019-01-01 02:01:00+00:00",
				"2019-01-01 03:01:00+00:00",
			},
		},

		{
			expr: "1 22 3 * * *",
			times: []string{
				"2019-01-01 00:00:00+00:00",
				"2019-01-01 03:22:01+00:00",
				"2019-01-02 03:22:01+00:00",
				"2019-01-03 03:22:01+00:00",
			},
		},
		{
			loc:  time.FixedZone("utc+8", 8*3600),
			expr: "1 22 3 * * *",
			times: []string{
				"2019-01-01 00:00:00+00:00", // 对应 2019-01-01 08:00:00+0800
				"2019-01-01 19:22:01+00:00", //      2019-01-02 03:22:01+0800
				"2019-01-03 03:22:01+08:00", //      2019-01-03 03:22:01+0800
				"2019-01-03 11:22:01-08:00", //      2019-01-04 03:22:01+0800
			},
		},

		{
			expr: "1 1 0 * * *",
			times: []string{
				"2019-06-30 12:00:00+00:00",
				"2019-07-01 00:01:01+00:00",
				"2019-07-02 00:01:01+00:00",
				"2019-07-03 00:01:01+00:00",
			},
		},

		{
			expr: "1 0 * * * *",
			times: []string{
				"2019-06-30 12:59:00+00:00",
				"2019-06-30 13:00:01+00:00",
				"2019-06-30 14:00:01+00:00",
			},
		},

		{ // 未指定日，只指定了星期
			expr: "1 22 3 * * 3",
			times: []string{
				"2019-01-01 00:00:00+00:00",
				"2019-01-02 03:22:01+00:00", // 周 3
				"2019-01-09 03:22:01+00:00",
				"2019-01-16 03:22:01+00:00",
				"2019-01-23 03:22:01+00:00",
				"2019-01-30 03:22:01+00:00",
				"2019-02-06 03:22:01+00:00", // 1 月份有 31
				"2019-02-13 03:22:01+00:00",
			},
		},

		{ // 指定了日和星期
			expr: "1 22 3 5 * 3",
			times: []string{
				"2019-01-01 00:00:00+00:00",
				"2019-01-02 03:22:01+00:00", // 周 3，1.2
				"2019-01-05 03:22:01+00:00", // 周 6，1.5
				"2019-01-09 03:22:01+00:00", // 周 3，1.9
				"2019-01-16 03:22:01+00:00", // 周 3，1.16
				"2019-01-23 03:22:01+00:00", // 周 3, 1.23
				"2019-01-30 03:22:01+00:00", // 周 3, 1.30
				"2019-02-05 03:22:01+00:00", // 周 2, 2.5
				"2019-02-06 03:22:01+00:00", // 周 3, 2.6
			},
		},

		{ // 未指定日，只指定了星期，以及跨月份
			expr: "1 22 3 * 3,7 3",
			times: []string{
				"2019-01-01 00:00:00+00:00",
				"2019-03-06 03:22:01+00:00", // 周 3
				"2019-03-13 03:22:01+00:00", // 周 3
				"2019-03-20 03:22:01+00:00", // 周 3
				"2019-03-27 03:22:01+00:00", // 周 3
				"2019-07-03 03:22:01+00:00", // 周 3
				"2019-07-10 03:22:01+00:00", // 周 3
				"2019-07-17 03:22:01+00:00", // 周 3
				"2019-07-24 03:22:01+00:00", // 周 3
				"2019-07-31 03:22:01+00:00", // 周 3
				"2020-03-04 03:22:01+00:00", // 周 3
				"2020-03-11 03:22:01+00:00", // 周 3
				"2020-03-18 03:22:01+00:00", // 周 3
				"2020-03-25 03:22:01+00:00", // 周 3
				"2020-07-01 03:22:01+00:00", // 周 3
				"2020-07-08 03:22:01+00:00", // 周 3
				"2020-07-15 03:22:01+00:00", // 周 3
				"2020-07-22 03:22:01+00:00", // 周 3
				"2020-07-29 03:22:01+00:00", // 周 3
				"2021-03-03 03:22:01+00:00", // 周 3
			},
		},

		{ // 未指定日，只指定了星期，以及跨月份
			expr: "1 22 3 * 3 3",
			times: []string{
				"2019-01-01 00:00:00+00:00",
				"2019-03-06 03:22:01+00:00", // 周 3
				"2019-03-13 03:22:01+00:00", // 周 3
				"2019-03-20 03:22:01+00:00", // 周 3
				"2019-03-27 03:22:01+00:00", // 周 3
				"2020-03-04 03:22:01+00:00", // 周 3
				"2020-03-11 03:22:01+00:00", // 周 3
				"2020-03-18 03:22:01+00:00", // 周 3
				"2020-03-25 03:22:01+00:00", // 周 3
				"2021-03-03 03:22:01+00:00", // 周 3
			},
		},

		{
			expr: "1,5 22 3 29 2 *", // 2.29 的相关测试
			times: []string{
				"2019-01-01 00:00:00+00:00",
				"2020-02-29 03:22:01+00:00",
				"2020-02-29 03:22:05+00:00",
				"2024-02-29 03:22:01+00:00",
				"2024-02-29 03:22:05+00:00",
			},
		},

		{
			expr: "1 22 3 31 * *", // 每个月 31 号
			times: []string{
				"2019-01-01 00:00:00+00:00",
				"2019-01-31 03:22:01+00:00",
				"2019-03-31 03:22:01+00:00",
				"2019-05-31 03:22:01+00:00",
				"2019-07-31 03:22:01+00:00",
				"2019-08-31 03:22:01+00:00",
				"2019-10-31 03:22:01+00:00",
				"2019-12-31 03:22:01+00:00",
				"2020-01-31 03:22:01+00:00",
				"2020-03-31 03:22:01+00:00",
			},
		},
	}

	const layout = "2006-01-02 15:04:05Z07:00"

	for i, t := range data {
		if len(t.times) < 2 {
			panic(fmt.Sprintf("%d times 最少两个元素", i))
		}

		loc := t.loc
		if loc == nil {
			loc = time.UTC
		}
		next, err := Parse(t.expr, loc)
		a.NotError(err).NotNil(next)

		for j := 1; j < len(t.times); j++ {
			last, err := time.Parse(layout, t.times[j-1])
			a.NotError(err)

			curr, err := time.Parse(layout, t.times[j])
			a.NotError(err)

			// 保持相同的时区再作比较
			n := next.Next(last)
			nn := n.In(time.UTC)
			cc := curr.In(time.UTC)

			a.Equal(nn.Year(), cc.Year(), "year 不同在 %d.times[%d]，返回值：%d，期望值：%d", i, j, n.Year(), curr.Year()).
				Equal(nn.Month(), cc.Month(), "month 不同在 %d.times[%d]，返回值：%d，期望值：%d", i, j, n.Month(), curr.Month()).
				Equal(nn.Day(), cc.Day(), "day 不同在 %d.times[%d]，返回值：%d，期望值：%d", i, j, n.Day(), curr.Day()).
				Equal(nn.Hour(), cc.Hour(), "hour 不同在 %d.times[%d]，返回值：%d，期望值：%d", i, j, n.Hour(), curr.Hour()).
				Equal(nn.Minute(), cc.Minute(), "minute 不同在 %d.times[%d]，返回值：%d，期望值：%d", i, j, n.Minute(), curr.Minute()).
				Equal(nn.Second(), cc.Second(), "second 不同在 %d.times[%d]，返回值：%d，期望值：%d", i, j, n.Second(), curr.Second())
		}
	}
}

func TestGetMonthDays(t *testing.T) {
	a := assert.New(t, false)

	var (
		leapDays = map[int]int{
			1:  31,
			3:  31,
			5:  31,
			7:  31,
			8:  31,
			10: 31,
			12: 31,
			2:  29,
			4:  30,
			6:  30,
			9:  30,
			11: 30,
		}
		days = map[int]int{
			1:  31,
			3:  31,
			5:  31,
			7:  31,
			8:  31,
			10: 31,
			12: 31,
			2:  28,
			4:  30,
			6:  30,
			9:  30,
			11: 30,
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
	a := assert.New(t, false)

	type test struct {
		// 输入值
		year    int
		month   time.Month
		weekday time.Weekday

		// 返回值
		day uint8
	}

	data := []*test{
		{
			year:    2019,
			month:   time.May,
			weekday: time.Wednesday,
			day:     1,
		},
		{
			year:    2019,
			month:   time.May,
			weekday: time.Saturday,
			day:     4,
		},
		{
			year:    2019,
			month:   time.May,
			weekday: time.Sunday,
			day:     5,
		},
		{
			year:    2020,
			month:   time.February,
			weekday: time.Saturday,
			day:     1,
		},
		{
			year:    2020,
			month:   time.February,
			weekday: time.Tuesday,
			day:     4,
		},
	}

	for index, item := range data {
		day := getMonthWeekDay(item.weekday, item.month, item.year)
		a.Equal(day, item.day, "%d 出错，返回值：%d，期望值：%d", index, day, item.day)
	}
}
