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

func TestRange(t *testing.T) {
	a := assert.New(t)

	a.Equal(Range(1, 5), []uint8{1, 2, 3, 4, 5})
	a.Equal(Range(1, 1), []uint8{1})
}
