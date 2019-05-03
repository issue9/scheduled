// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package expr

import (
	"math"
	"testing"

	"github.com/issue9/assert"
)

// 2**y1 + 2**y2 + 2**y3 ...
func pow2(y ...uint64) uint64 {
	var p float64

	for _, yy := range y {
		p += math.Pow(2, float64(yy))
	}
	return uint64(p)
}

func TestField_next(t *testing.T) {
	a := assert.New(t)

	type test struct {
		// 输入
		typ   int
		curr  uint8
		list  uint64
		carry bool

		// 输出
		v uint8
		c bool
	}

	var data = []*test{
		&test{
			typ:   secondIndex,
			curr:  0,
			list:  pow2(1, 3, 5),
			carry: true,
			v:     1,
			c:     false,
		},

		&test{
			typ:   secondIndex,
			curr:  1,
			list:  pow2(1, 3, 5),
			carry: false,
			v:     1,
			c:     false,
		},

		&test{
			typ:   secondIndex,
			curr:  1,
			list:  pow2(1, 3, 5),
			carry: true,
			v:     3,
			c:     false,
		},

		&test{
			typ:   secondIndex,
			curr:  5,
			list:  pow2(1, 3, 5),
			carry: false,
			v:     5,
			c:     false,
		},

		&test{
			typ:   secondIndex,
			curr:  59,
			list:  any,
			carry: false,
			v:     59,
			c:     false,
		},

		&test{
			typ:   secondIndex,
			curr:  59,
			list:  step,
			carry: false,
			v:     59,
			c:     false,
		},

		&test{
			typ:   secondIndex,
			curr:  59,
			list:  step,
			carry: true,
			v:     0,
			c:     true,
		},

		&test{
			typ:   secondIndex,
			curr:  5,
			list:  pow2(1, 3, 5),
			carry: true,
			v:     1,
			c:     true,
		},

		&test{
			typ:   secondIndex,
			curr:  5,
			list:  any,
			carry: true,
			v:     5,
			c:     true,
		},

		&test{
			typ:   secondIndex,
			curr:  5,
			list:  any,
			carry: false,
			v:     5,
			c:     false,
		},

		&test{
			typ:   secondIndex,
			curr:  5,
			list:  step,
			carry: true,
			v:     6,
			c:     false,
		},

		&test{
			typ:   secondIndex,
			curr:  5,
			list:  step,
			carry: false,
			v:     5,
			c:     false,
		},

		&test{
			typ:   dayIndex,
			curr:  5,
			list:  step,
			carry: false,
			v:     5,
			c:     false,
		},

		&test{
			typ:   dayIndex,
			curr:  5,
			list:  step,
			carry: true,
			v:     6,
			c:     false,
		},

		&test{
			typ:   dayIndex,
			curr:  4,
			list:  pow2(1, 3, 4, 7),
			carry: true,
			v:     7,
			c:     false,
		},
	}

	for i, item := range data {
		f := fields[item.typ]
		v, c := f.next(item.curr, item.list, item.carry)
		a.Equal(v, item.v, "data[%d] 错误，实际返回:%d 期望值:%d", i, v, item.v).
			Equal(c, item.c, "data[%d] 错误，实际返回:%v 期望值:%v", i, c, item.c)
	}
}

func TestParse(t *testing.T) {
	a := assert.New(t)

	type expr struct {
		expr   string
		hasErr bool
		vals   []uint64
	}

	exprs := []*expr{
		&expr{
			expr: "1-3,10,9 * 3-7 * * 1",
			vals: []uint64{pow2(1, 2, 3, 10, 9), step, pow2(3, 4, 5, 6, 7), step, step, pow2(1)},
		},
		&expr{
			expr: "* * * * * 1",
			vals: []uint64{any, any, any, any, any, pow2(1)},
		},
		&expr{
			expr: "* * * * * 0",
			vals: []uint64{any, any, any, any, any, pow2(0)},
		},
		&expr{
			expr: "* * * * * 6",
			vals: []uint64{any, any, any, any, any, pow2(6)},
		},
		&expr{
			expr: "* 3 * * * 6",
			vals: []uint64{any, pow2(3), step, step, step, pow2(6)},
		},
		&expr{
			expr: "@daily",
			vals: []uint64{pow2(0), pow2(0), pow2(0), step, step, step},
		},
		&expr{ // 参数错误
			expr:   "",
			hasErr: true,
			vals:   nil,
		},
		&expr{ // 指令不存在
			expr:   "@not-exists",
			hasErr: true,
			vals:   nil,
		},
		&expr{ // 解析错误
			expr:   "* * * * * 7-a",
			hasErr: true,
			vals:   nil,
		},
		&expr{ // 值超出范围
			expr:   "* * * * * 7",
			hasErr: true,
			vals:   nil,
		},
		&expr{ // 表达式内容不够长
			expr:   "*",
			hasErr: true,
			vals:   nil,
		},
		&expr{ // 表达式内容太长
			expr:   "* * * * * * x",
			hasErr: true,
			vals:   nil,
		},
		&expr{ // 都为 *
			expr:   "* * * * * *",
			hasErr: true,
			vals:   nil,
		},
	}

	for _, v := range exprs {
		expr, err := Parse(v.expr)
		if v.hasErr {
			a.Error(err, "测试 %s 时出错", v.expr).
				Nil(expr)
			continue
		}

		a.NotError(err, "测试 %s 时出错 %s", v.expr, err)
		a.Equal(expr.data, v.vals, "测试 %s 时出错，期望值：%v，实际返回值：%v", v.expr, v.vals, expr.data)
	}
}

func TestParseField(t *testing.T) {
	a := assert.New(t)

	type field struct {
		typ    int
		field  string
		hasErr bool
		vals   uint64
	}

	fs := []*field{
		&field{
			typ:   secondIndex,
			field: "*",
			vals:  any,
		},
		&field{
			typ:   secondIndex,
			field: "2",
			vals:  pow2(2),
		},
		&field{
			typ:   secondIndex,
			field: "1,2",
			vals:  pow2(1, 2),
		},
		&field{
			typ:   secondIndex,
			field: "1,2,4,7,",
			vals:  pow2(1, 2, 4, 7),
		},
		&field{
			typ:   secondIndex,
			field: "0-4",
			vals:  pow2(0, 1, 2, 3, 4),
		},
		&field{
			typ:   monthIndex,
			field: "1-4",
			vals:  pow2(1, 2, 3, 4),
		},
		&field{
			typ:   secondIndex,
			field: "1-2",
			vals:  pow2(1, 2),
		},
		&field{
			typ:   secondIndex,
			field: "1-4,9",
			vals:  pow2(1, 2, 3, 4, 9),
		},
		&field{
			typ:   dayIndex,
			field: "31",
			vals:  pow2(31),
		},
		&field{
			typ:   secondIndex,
			field: "1-4,9,19-21",
			vals:  pow2(1, 2, 3, 4, 9, 19, 20, 21),
		},

		&field{ // 超出范围，月份从 1 开始
			typ:    monthIndex,
			field:  "0-4",
			hasErr: true,
		},
		&field{ // 超出范围，月份没有 13
			typ:    monthIndex,
			field:  "1-13",
			hasErr: true,
		},
		&field{ // 重复的值
			typ:    secondIndex,
			field:  "1-4,9,9-11",
			hasErr: true,
		},
		&field{ // 无效的数值
			typ:    secondIndex,
			field:  "a1",
			hasErr: true,
		},
		&field{ // 无效的数值
			typ:    secondIndex,
			field:  "a1-a3",
			hasErr: true,
		},
		&field{ // 无效的数值
			typ:    secondIndex,
			field:  "1-a3",
			hasErr: true,
		},
		&field{ // 无效的数值
			typ:    secondIndex,
			field:  "-3",
			hasErr: true,
		},
		&field{ // 无效的数值
			typ:    secondIndex,
			field:  "1-",
			hasErr: true,
		},
		&field{ // 无效的数值
			typ:    secondIndex,
			field:  "-a3",
			hasErr: true,
		},
	}

	for _, v := range fs {
		val, err := parseField(fields[v.typ], v.field)
		if v.hasErr {
			a.Error(err, "测试 %s 时出错", v.field).
				Equal(val, 0)
			continue
		}

		a.NotError(err)
		a.Equal(val, v.vals, "测试 %s 时出错 实际返回:%d，期望值：%d", v.field, val, v.vals)
	}
}

func TestIntRange(t *testing.T) {
	a := assert.New(t)

	a.Equal(intRange(1, 5), []uint8{1, 2, 3, 4, 5})
	a.Equal(intRange(1, 1), []uint8{1})
}
