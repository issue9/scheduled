// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/issue9/assert"
)

var _ Nexter = &Expr{}

const day = 24 * time.Hour

func TestExpr_Next(t *testing.T) {
	a := assert.New(t)

	type test struct {
		expr  string
		times []time.Time // 前一个是后一个的参数
	}

	base := time.Date(2019, 1, 1, 0, 0, 0, 123, time.UTC) // 周二
	week := time.Wednesday
	wdays := week - base.Weekday()
	if wdays < 0 {
		wdays += 7
	}
	weekdur := time.Duration(wdays) * day
	var data = []*test{
		&test{
			expr: "1 * * * * *",
			times: []time.Time{
				base,
				base.Add(1 * time.Second),
				base.Add(1 * time.Second).Add(1 * time.Minute), // 进位
				base.Add(1 * time.Second).Add(2 * time.Minute),
				base.Add(1 * time.Second).Add(3 * time.Minute),
			},
		},

		&test{
			expr: "* 1 * * * *",
			times: []time.Time{
				base,
				base.Add(1 * time.Minute),
				base.Add(1 * time.Minute).Add(1 * time.Hour), // 进位
				base.Add(1 * time.Minute).Add(2 * time.Hour),
				base.Add(1 * time.Minute).Add(3 * time.Hour),
			},
		},

		&test{
			expr: "1 22 3 * * *",
			times: []time.Time{
				base,
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour),
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add(day),
			},
		},

		&test{ // 未指定日，只指定了星期
			expr: "1 22 3 * * " + strconv.Itoa(int(week)),
			times: []time.Time{
				base,
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add(weekdur),
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add(weekdur).Add(7 * day),
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add(weekdur).Add(2 * 7 * day),
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add(weekdur).Add(3 * 7 * day),
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add(weekdur).Add(4 * 7 * day),
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add(weekdur).Add(5 * 7 * day),
			},
		},

		&test{ // 指定了日和星期
			expr: "1 22 3 5 * " + strconv.Itoa(int(week)),
			times: []time.Time{
				base,
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add(weekdur),                                    // 周 3，2 号
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add(weekdur).Add(3 * day),                       // 周 6，5 号
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add(weekdur).Add((3 + 4) * day),                 // 周 3，9 号
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add(weekdur).Add((3 + 4 + 7) * day),             // 周 3，16
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add(weekdur).Add((3 + 4 + 7 + 7) * day),         // 3, 23
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add(weekdur).Add((3 + 4 + 7 + 7 + 7) * day),     // 3, 30
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add(weekdur).Add((3 + 4 + 7 + 7 + 7 + 6) * day), // 2, 2.6
			},
		},

		&test{
			expr: "1 22 3 31 * *",
			times: []time.Time{
				base,
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add(30 * day),
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add((30 + 28 + 31) * day),                                                             // 3 月
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add((30 + 28 + 31 + 30 + 31) * day),                                                   // 5 月
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add((30 + 28 + 31 + 30 + 31 + 30 + 31) * day),                                         // 7 月
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add((30 + 28 + 31 + 30 + 31 + 30 + 31 + 31) * day),                                    // 8 月
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add((30 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31) * day),                          // 10 月
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add((30 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30 + 31) * day),                // 12 月
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add((30 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30 + 31 + 31) * day),           // 2020.1 月
				base.Add(1 * time.Second).Add(22 * time.Minute).Add(3 * time.Hour).Add((30 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30 + 31 + 31 + 29 + 31) * day), // 2020.3 月
			},
		},
	}

	for i, t := range data {
		if len(t.times) < 2 {
			panic(fmt.Sprintf("%d times 最少两个元素", i))
		}

		next, err := parseExpr(t.expr)
		a.NotError(err).NotNil(next)

		for j := 1; j < len(t.times); j++ {
			last := t.times[j-1]
			n := next.Next(last)

			curr := t.times[j]

			a.Equal(n.Year(), curr.Year(), "year 不同在 %d.times[%d]，%d:%d 个元素", i, j, n.Year(), curr.Year()).
				Equal(n.Month(), curr.Month(), "month 不同在 %d.times[%d]，%d:%d 个元素", i, j, n.Month(), curr.Month()).
				Equal(n.Day(), curr.Day(), "day 不同在 %d.times[%d]，%d:%d 个元素", i, j, n.Day(), curr.Day()).
				Equal(n.Hour(), curr.Hour(), "hour 不同在 %d.times[%d]，%d:%d 个元素", i, j, n.Hour(), curr.Hour()).
				Equal(n.Minute(), curr.Minute(), "minute 不同在 %d.times[%d]，%d:%d 个元素", i, j, n.Minute(), curr.Minute()).
				Equal(n.Second(), curr.Second(), "second 不同在 %d.times[%d]，%d:%d 个元素", i, j, n.Second(), curr.Second())
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

func TestNext(t *testing.T) {
	a := assert.New(t)

	type test struct {
		// 输入
		typ   int
		curr  uint8
		list  []uint8
		carry bool

		// 输出
		v uint8
		c bool
	}

	var data = []*test{
		&test{
			typ:   secondIndex,
			curr:  0,
			list:  []uint8{1, 3, 5},
			carry: true,
			v:     1,
			c:     false,
		},

		&test{
			typ:   secondIndex,
			curr:  1,
			list:  []uint8{1, 3, 5},
			carry: false,
			v:     1,
			c:     false,
		},

		&test{
			typ:   secondIndex,
			curr:  1,
			list:  []uint8{1, 3, 5},
			carry: true,
			v:     3,
			c:     false,
		},

		&test{
			typ:   secondIndex,
			curr:  5,
			list:  []uint8{1, 3, 5},
			carry: false,
			v:     5,
			c:     false,
		},

		&test{
			typ:   secondIndex,
			curr:  59,
			carry: false,
			v:     59,
			c:     false,
		},

		&test{
			typ:   secondIndex,
			curr:  59,
			carry: true,
			v:     0,
			c:     true,
		},

		&test{
			typ:   secondIndex,
			curr:  5,
			list:  []uint8{1, 3, 5},
			carry: true,
			v:     1,
			c:     true,
		},

		&test{
			typ:   secondIndex,
			curr:  5,
			list:  nil,
			carry: true,
			v:     6,
			c:     false,
		},

		&test{
			typ:   secondIndex,
			curr:  5,
			list:  nil,
			carry: false,
			v:     5,
			c:     false,
		},
	}

	for _, item := range data {
		v, c := next(item.typ, item.curr, item.list, item.carry)
		a.Equal(v, item.v).
			Equal(c, item.c)
	}
}
