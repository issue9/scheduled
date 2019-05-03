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
