// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import (
	"testing"

	"github.com/issue9/assert"
)

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
