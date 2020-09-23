// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

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

func TestParse(t *testing.T) {
	a := assert.New(t)

	type test struct {
		expr   string
		hasErr bool
		vals   []uint64
	}

	exprs := []*test{
		{
			expr: "1-3,10,9 * 3-7 * * 1",
			vals: []uint64{pow2(1, 2, 3, 10, 9), step, pow2(3, 4, 5, 6, 7), step, step, pow2(1)},
		},
		{
			expr: "* * * * * 1",
			vals: []uint64{any, any, any, any, any, pow2(1)},
		},
		{
			expr: "* * * * * 0",
			vals: []uint64{any, any, any, any, any, pow2(0)},
		},
		{
			expr: "* * * * * 6",
			vals: []uint64{any, any, any, any, any, pow2(6)},
		},
		{
			expr: "* 3 * * * 6",
			vals: []uint64{any, pow2(3), step, step, step, pow2(6)},
		},
		{
			expr: "@daily",
			vals: []uint64{pow2(0), pow2(0), pow2(0), step, step, step},
		},
		{ // 参数错误
			expr:   "",
			hasErr: true,
			vals:   nil,
		},
		{ // 指令不存在
			expr:   "@not-exists",
			hasErr: true,
			vals:   nil,
		},
		{ // 解析错误
			expr:   "* * * * * 7-a",
			hasErr: true,
			vals:   nil,
		},
		{ // 值超出范围
			expr:   "* * * * * 8",
			hasErr: true,
			vals:   nil,
		},
		{ // 表达式内容不够长
			expr:   "*",
			hasErr: true,
			vals:   nil,
		},
		{ // 表达式内容太长
			expr:   "* * * * * * x",
			hasErr: true,
			vals:   nil,
		},
		{ // 都为 *
			expr:   "* * * * * *",
			hasErr: true,
			vals:   nil,
		},
	}

	for _, v := range exprs {
		s, err := Parse(v.expr)
		if v.hasErr {
			a.Error(err, "测试 %s 时出错", v.expr).
				Nil(s)
			continue
		}

		c, ok := s.(*cron)
		a.True(ok).NotNil(c)
		a.NotError(err, "测试 %s 时出错 %s", v.expr, err)
		a.Equal(c.data, v.vals, "测试 %s 时出错，期望值：%v，实际返回值：%v", v.expr, v.vals, c.data)
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
		{
			typ:   secondIndex,
			field: "*",
			vals:  any,
		},
		{
			typ:   secondIndex,
			field: "2",
			vals:  pow2(2),
		},
		{
			typ:   secondIndex,
			field: "1,2",
			vals:  pow2(1, 2),
		},
		{
			typ:   secondIndex,
			field: "1,2,4,7,",
			vals:  pow2(1, 2, 4, 7),
		},
		{
			typ:   secondIndex,
			field: "0-4",
			vals:  pow2(0, 1, 2, 3, 4),
		},
		{
			typ:   monthIndex,
			field: "1-4",
			vals:  pow2(1, 2, 3, 4),
		},
		{
			typ:   secondIndex,
			field: "1-2",
			vals:  pow2(1, 2),
		},
		{
			typ:   secondIndex,
			field: "1-4,9",
			vals:  pow2(1, 2, 3, 4, 9),
		},
		{
			typ:   dayIndex,
			field: "31",
			vals:  pow2(31),
		},

		// week 相关的测试
		{
			typ:   weekIndex,
			field: "7",
			vals:  pow2(0),
		},
		{ // 0 与 7 是相同的值
			typ:    weekIndex,
			field:  "0-7",
			hasErr: true,
		},
		{ // 超出范围
			typ:    weekIndex,
			field:  "0-8",
			hasErr: true,
		},
		{
			typ:   weekIndex,
			field: "5-7",
			vals:  pow2(0, 5, 6),
		},

		{
			typ:   secondIndex,
			field: "1-4,9,19-21",
			vals:  pow2(1, 2, 3, 4, 9, 19, 20, 21),
		},

		{ // 超出范围，月份从 1 开始
			typ:    monthIndex,
			field:  "0-4",
			hasErr: true,
		},
		{ // 超出范围，月份没有 13
			typ:    monthIndex,
			field:  "1-13",
			hasErr: true,
		},
		{ // 重复的值
			typ:    secondIndex,
			field:  "1-4,9,9-11",
			hasErr: true,
		},
		{ // 无效的数值
			typ:    secondIndex,
			field:  "a1",
			hasErr: true,
		},
		{ // 无效的数值
			typ:    secondIndex,
			field:  "a1-a3",
			hasErr: true,
		},
		{ // 无效的数值
			typ:    secondIndex,
			field:  "1-a3",
			hasErr: true,
		},
		{ // 无效的数值
			typ:    secondIndex,
			field:  "-3",
			hasErr: true,
		},
		{ // 无效的数值
			typ:    secondIndex,
			field:  "1-",
			hasErr: true,
		},
		{ // 无效的数值
			typ:    secondIndex,
			field:  "-a3",
			hasErr: true,
		},
	}

	for _, v := range fs {
		val, err := parseField(v.typ, v.field)
		if v.hasErr {
			a.Error(err, "测试 %s 时出错", v.field).
				Equal(val, 0)
			continue
		}

		a.NotError(err)
		a.Equal(val, v.vals, "测试 %s 时出错 实际返回:%d，期望值：%d", v.field, val, v.vals)
	}
}
