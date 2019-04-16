// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import (
	"testing"
	"time"

	"github.com/issue9/assert"
)

var _ Nexter = &Expr{}

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
		curr  uint8
		list  []uint8
		carry bool

		// 输出
		v uint8
		c bool
	}

	var data = []*test{
		&test{
			curr:  0,
			list:  []uint8{1, 3, 5},
			carry: true,
			v:     1,
			c:     false,
		},

		&test{
			curr:  1,
			list:  []uint8{1, 3, 5},
			carry: false,
			v:     1,
			c:     false,
		},

		&test{
			curr:  1,
			list:  []uint8{1, 3, 5},
			carry: true,
			v:     3,
			c:     false,
		},

		&test{
			curr:  5,
			list:  []uint8{1, 3, 5},
			carry: false,
			v:     5,
			c:     false,
		},

		&test{
			curr:  5,
			list:  []uint8{1, 3, 5},
			carry: true,
			v:     1,
			c:     true,
		},

		&test{
			curr:  5,
			list:  nil,
			carry: true,
			v:     6,
			c:     false,
		},

		&test{
			curr:  5,
			list:  nil,
			carry: false,
			v:     5,
			c:     false,
		},
	}

	for _, item := range data {
		v, c := next(item.curr, item.list, item.carry)
		a.Equal(v, item.v).
			Equal(c, item.c)
	}
}
