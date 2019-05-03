// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package expr

import (
	"testing"

	"github.com/issue9/assert"
)

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
