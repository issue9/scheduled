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

func TestParseField(t *testing.T) {
	a := assert.New(t)

	type field struct {
		field  string
		hasErr bool
		vals   []uint8
	}

	fields := []*field{
		&field{
			field:  "*",
			hasErr: false,
			vals:   nil,
		},
		&field{
			field:  "2",
			hasErr: false,
			vals:   []uint8{2},
		},
		&field{
			field:  "1,2",
			hasErr: false,
			vals:   []uint8{1, 2},
		},
		&field{
			field:  "1,2,4,7,",
			hasErr: false,
			vals:   []uint8{1, 2, 4, 7},
		},
		&field{
			field:  "1-4",
			hasErr: false,
			vals:   []uint8{1, 2, 3, 4},
		},
		&field{
			field:  "1-2",
			hasErr: false,
			vals:   []uint8{1, 2},
		},
		&field{
			field:  "1-4,9",
			hasErr: false,
			vals:   []uint8{1, 2, 3, 4, 9},
		},
		&field{
			field:  "1-4,9,19-21",
			hasErr: false,
			vals:   []uint8{1, 2, 3, 4, 9, 19, 20, 21},
		},
		&field{ // 重复的值
			field:  "1-4,9,9-11",
			hasErr: true,
			vals:   nil,
		},
		&field{ // 无效的数值
			field:  "a1",
			hasErr: true,
			vals:   nil,
		},
		&field{ // 无效的数值
			field:  "a1-a3",
			hasErr: true,
			vals:   nil,
		},
		&field{ // 无效的数值
			field:  "1-a3",
			hasErr: true,
			vals:   nil,
		},
		&field{ // 无效的数值
			field:  "-3",
			hasErr: true,
			vals:   nil,
		},
		&field{ // 无效的数值
			field:  "1-",
			hasErr: true,
			vals:   nil,
		},
		&field{ // 无效的数值
			field:  "-a3",
			hasErr: true,
			vals:   nil,
		},
	}

	for _, v := range fields {
		val, err := parseField(v.field)
		if v.hasErr {
			a.Error(err, "测试 %s 时出错", v.field).
				Nil(val)
			continue
		}

		a.NotError(err)
		a.Equal(val, v.vals, "测试 %s 时出错", v.field)
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

func TestIntRange(t *testing.T) {
	a := assert.New(t)

	a.Equal(intRange(1, 5), []uint8{1, 2, 3, 4, 5})
	a.Equal(intRange(1, 1), []uint8{1})
}
