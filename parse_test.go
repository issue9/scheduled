// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import (
	"testing"

	"github.com/issue9/assert"
)

func TestParseExpr(t *testing.T) {
	a := assert.New(t)

	type expr struct {
		expr   string
		hasErr bool
		vals   [][]uint8
	}

	exprs := []*expr{
		&expr{ // 表达式内容不够长
			expr:   "*",
			hasErr: true,
			vals:   nil,
		},
		&expr{ // 表达式内容太长
			expr:   "* * * * * * *",
			hasErr: true,
			vals:   nil,
		},
	}

	for _, v := range exprs {
		expr, err := parseExpr(v.expr)
		if v.hasErr {
			a.Error(err, "测试 %s 时出错", v.expr).
				Nil(expr)
			continue
		}

		a.NotError(err)
		a.Equal(expr.data, v.vals, "测试 %s 时出错", v.expr)
	}
}

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

func TestIntRange(t *testing.T) {
	a := assert.New(t)

	a.Equal(intRange(1, 5), []uint8{1, 2, 3, 4, 5})
	a.Equal(intRange(1, 1), []uint8{1})
}
