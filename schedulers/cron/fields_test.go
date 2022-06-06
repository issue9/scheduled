// SPDX-License-Identifier: MIT

package cron

import (
	"testing"

	"github.com/issue9/assert/v2"
)

func TestParseField(t *testing.T) {
	a := assert.New(t, false)

	type field struct {
		typ    int
		field  string
		hasErr bool
		vals   fields
	}

	fs := []*field{
		{
			typ:   secondIndex,
			field: "*",
			vals:  asterisk,
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

func TestBits_next(t *testing.T) {
	a := assert.New(t, false)

	type test struct {
		// 输入
		typ   int
		curr  int
		bits  fields
		carry bool

		// 输出
		v uint8
		c bool
	}

	var data = []*test{
		{
			typ:   secondIndex,
			curr:  0,
			bits:  pow2(1, 3, 5),
			carry: true,
			v:     1,
			c:     false,
		},

		{
			typ:   secondIndex,
			curr:  1,
			bits:  pow2(1, 3, 5),
			carry: false,
			v:     1,
			c:     false,
		},

		{
			typ:   secondIndex,
			curr:  1,
			bits:  pow2(1, 3, 5),
			carry: true,
			v:     3,
			c:     false,
		},

		{
			typ:   secondIndex,
			curr:  5,
			bits:  pow2(1, 3, 5),
			carry: false,
			v:     5,
			c:     false,
		},

		{
			typ:   secondIndex,
			curr:  59,
			bits:  asterisk,
			carry: false,
			v:     59,
			c:     false,
		},

		{
			typ:   secondIndex,
			curr:  59,
			bits:  step,
			carry: false,
			v:     59,
			c:     false,
		},

		{
			typ:   secondIndex,
			curr:  59,
			bits:  step,
			carry: true,
			v:     0,
			c:     true,
		},

		{
			typ:   secondIndex,
			curr:  5,
			bits:  pow2(1, 3, 5),
			carry: true,
			v:     1,
			c:     true,
		},

		{
			typ:   secondIndex,
			curr:  5,
			bits:  asterisk,
			carry: true,
			v:     5,
			c:     true,
		},

		{
			typ:   secondIndex,
			curr:  5,
			bits:  asterisk,
			carry: false,
			v:     5,
			c:     false,
		},

		{
			typ:   secondIndex,
			curr:  5,
			bits:  step,
			carry: true,
			v:     6,
			c:     false,
		},

		{
			typ:   secondIndex,
			curr:  5,
			bits:  step,
			carry: false,
			v:     5,
			c:     false,
		},

		{
			typ:   dayIndex,
			curr:  5,
			bits:  step,
			carry: false,
			v:     5,
			c:     false,
		},

		{
			typ:   dayIndex,
			curr:  5,
			bits:  step,
			carry: true,
			v:     6,
			c:     false,
		},

		{
			typ:   dayIndex,
			curr:  4,
			bits:  pow2(1, 3, 4, 7),
			carry: true,
			v:     7,
			c:     false,
		},
	}

	for i, item := range data {
		v, c := item.bits.next(item.curr, bounds[item.typ], item.carry)
		a.Equal(v, item.v, "data[%d] 错误，实际返回:%d 期望值:%d", i, v, item.v).
			Equal(c, item.c, "data[%d] 错误，实际返回:%v 期望值:%v", i, c, item.c)
	}
}
