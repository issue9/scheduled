// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package expr

import (
	"fmt"
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestExpr_Next(t *testing.T) {
	a := assert.New(t)

	layout := "2006-01-02 15:04:05"
	type test struct {
		expr  string
		times []string // 前一个是后一个的参数
	}

	var data = []*test{
		&test{
			expr: "1 * * * * *",
			times: []string{
				"2019-01-01 00:00:00",
				"2019-01-01 00:00:01",
				"2019-01-01 00:01:01",
				"2019-01-01 00:02:01",
				"2019-01-01 00:03:01",
			},
		},

		&test{
			expr: "* 1 * * * *",
			times: []string{
				"2019-01-01 00:00:00",
				"2019-01-01 00:01:00",
				"2019-01-01 01:01:00",
				"2019-01-01 02:01:00",
				"2019-01-01 03:01:00",
			},
		},

		&test{
			expr: "1 22 3 * * *",
			times: []string{
				"2019-01-01 00:00:00",
				"2019-01-01 03:22:01",
				"2019-01-02 03:22:01",
				"2019-01-03 03:22:01",
			},
		},

		&test{ // 未指定日，只指定了星期
			expr: "1 22 3 * * 3",
			times: []string{
				"2019-01-01 00:00:00",
				"2019-01-02 03:22:01", // 周 3
				"2019-01-09 03:22:01",
				"2019-01-16 03:22:01",
				"2019-01-23 03:22:01",
				"2019-01-30 03:22:01",
				"2019-02-06 03:22:01", // 1 月份有 31
				"2019-02-13 03:22:01",
			},
		},

		&test{ // 指定了日和星期
			expr: "1 22 3 5 * 3",
			times: []string{
				"2019-01-01 00:00:00",
				"2019-01-02 03:22:01", // 周 3，1.2
				"2019-01-05 03:22:01", // 周 6，1.5
				"2019-01-09 03:22:01", // 周 3，1.9
				"2019-01-16 03:22:01", // 周 3，1.16
				"2019-01-23 03:22:01", // 周 3, 1.23
				"2019-01-30 03:22:01", // 周 3, 1.30
				"2019-02-05 03:22:01", // 周 2, 2.5
				"2019-02-06 03:22:01", // 周 3, 2.6
			},
		},

		&test{ // 未指定日，只指定了星期，以及跨月份
			expr: "1 22 3 * 3,7 3",
			times: []string{
				"2019-01-01 00:00:00",
				"2019-03-06 03:22:01", // 周 3
				"2019-03-13 03:22:01", // 周 3
				"2019-03-20 03:22:01", // 周 3
				"2019-03-27 03:22:01", // 周 3
				"2019-07-03 03:22:01", // 周 3
				"2019-07-10 03:22:01", // 周 3
				"2019-07-17 03:22:01", // 周 3
				"2019-07-24 03:22:01", // 周 3
				"2019-07-31 03:22:01", // 周 3
				"2020-03-04 03:22:01", // 周 3
				"2020-03-11 03:22:01", // 周 3
				"2020-03-18 03:22:01", // 周 3
				"2020-03-25 03:22:01", // 周 3
				"2020-07-01 03:22:01", // 周 3
				"2020-07-08 03:22:01", // 周 3
				"2020-07-15 03:22:01", // 周 3
				"2020-07-22 03:22:01", // 周 3
				"2020-07-29 03:22:01", // 周 3
				"2021-03-03 03:22:01", // 周 3
			},
		},

		&test{ // 未指定日，只指定了星期，以及跨月份
			expr: "1 22 3 * 3 3",
			times: []string{
				"2019-01-01 00:00:00",
				"2019-03-06 03:22:01", // 周 3
				"2019-03-13 03:22:01", // 周 3
				"2019-03-20 03:22:01", // 周 3
				"2019-03-27 03:22:01", // 周 3
				"2020-03-04 03:22:01", // 周 3
				"2020-03-11 03:22:01", // 周 3
				"2020-03-18 03:22:01", // 周 3
				"2020-03-25 03:22:01", // 周 3
				"2021-03-03 03:22:01", // 周 3
			},
		},

		&test{
			expr: "1,5 22 3 29 2 *", // 2.29 的相关测试
			times: []string{
				"2019-01-01 00:00:00",
				"2020-02-29 03:22:01",
				"2020-02-29 03:22:05",
				"2024-02-29 03:22:01",
				"2024-02-29 03:22:05",
			},
		},

		&test{
			expr: "1 22 3 31 * *", // 每个月 31 号
			times: []string{
				"2019-01-01 00:00:00",
				"2019-01-31 03:22:01",
				"2019-03-31 03:22:01",
				"2019-05-31 03:22:01",
				"2019-07-31 03:22:01",
				"2019-08-31 03:22:01",
				"2019-10-31 03:22:01",
				"2019-12-31 03:22:01",
				"2020-01-31 03:22:01",
				"2020-03-31 03:22:01",
			},
		},
	}

	for i, t := range data {
		if len(t.times) < 2 {
			panic(fmt.Sprintf("%d times 最少两个元素", i))
		}

		next, err := Parse(t.expr)
		a.NotError(err).NotNil(next)

		for j := 1; j < len(t.times); j++ {
			last, err := time.Parse(layout, t.times[j-1])
			a.NotError(err)
			n := next.Next(last)

			curr, err := time.Parse(layout, t.times[j])
			a.NotError(err)

			a.Equal(n.Year(), curr.Year(), "year 不同在 %d.times[%d]，返回值：%d，期望值：%d", i, j, n.Year(), curr.Year()).
				Equal(n.Month(), curr.Month(), "month 不同在 %d.times[%d]，返回值：%d，期望值：%d", i, j, n.Month(), curr.Month()).
				Equal(n.Day(), curr.Day(), "day 不同在 %d.times[%d]，返回值：%d，期望值：%d", i, j, n.Day(), curr.Day()).
				Equal(n.Hour(), curr.Hour(), "hour 不同在 %d.times[%d]，返回值：%d，期望值：%d", i, j, n.Hour(), curr.Hour()).
				Equal(n.Minute(), curr.Minute(), "minute 不同在 %d.times[%d]，返回值：%d，期望值：%d", i, j, n.Minute(), curr.Minute()).
				Equal(n.Second(), curr.Second(), "second 不同在 %d.times[%d]，返回值：%d，期望值：%d", i, j, n.Second(), curr.Second())
		}
	}
}

func TestGetMonthDays(t *testing.T) {
	a := assert.New(t)

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
	a := assert.New(t)

	type test struct {
		// 输入值
		year    int
		month   time.Month
		weekday time.Weekday

		// 返回值
		day uint8
	}

	data := []*test{
		&test{
			year:    2019,
			month:   time.May,
			weekday: time.Wednesday,
			day:     1,
		},
		&test{
			year:    2019,
			month:   time.May,
			weekday: time.Saturday,
			day:     4,
		},
		&test{
			year:    2019,
			month:   time.May,
			weekday: time.Sunday,
			day:     5,
		},
		&test{
			year:    2020,
			month:   time.February,
			weekday: time.Saturday,
			day:     1,
		},
		&test{
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
