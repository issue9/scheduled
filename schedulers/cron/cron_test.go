// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package cron

import (
	"testing"
	"time"

	"github.com/issue9/assert/v5"

	"github.com/issue9/scheduled/schedulers"
)

var _ schedulers.Scheduler = &cron{}

func TestReboot(t *testing.T) {
	a := assert.New(t, false)

	s, err := Parse("@reboot", time.Local)
	a.NotError(err).NotNil(s)
	a.False(s.Next(time.Now()).IsZero()).
		True(s.Next(time.Now()).IsZero()).
		True(s.Next(time.Now()).IsZero())
}

func TestParse(t *testing.T) {
	a := assert.New(t, false)

	type test struct {
		expr   string
		hasErr bool
		vals   []fields
	}

	exprs := []*test{
		{
			expr: "1-3,10,9 * 3-7 * * 1",
			vals: []fields{1<<1 + 1<<2 + 1<<3 + 1<<10 + 1<<9, step, 1<<3 + 1<<4 + 1<<5 + 1<<6 + 1<<7, step, step, 1 << 1},
		},
		{
			expr: "* * * * * 1",
			vals: []fields{asterisk, asterisk, asterisk, asterisk, asterisk, 1 << 1},
		},
		{
			expr: "* * * * * 0",
			vals: []fields{asterisk, asterisk, asterisk, asterisk, asterisk, 1 << 0},
		},
		{
			expr: "* * * * * 6",
			vals: []fields{asterisk, asterisk, asterisk, asterisk, asterisk, 1 << 6},
		},
		{
			expr: "* 3 * * * 6",
			vals: []fields{asterisk, 1 << 3, step, step, step, 1 << 6},
		},
		{
			expr: "@daily",
			vals: []fields{1 << 0, 1 << 0, 1 << 0, step, step, step},
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
		s, err := Parse(v.expr, time.UTC)
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
