// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package expr

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	// any 和 step 是两个特殊的标记位，需要大于 60（所有字段中，秒数越最大，
	// 但不会超过 60）

	// any 表示当前字段可以是任意值，即对值不做任意要求，
	// 甚至可以一直是相同的值，也不会做累加。
	any = 1 << 61

	// step 表示当前字段是允许范围内的所有值。
	// 每次计算时，按其当前值加 1 即可。
	step = 1 << 62
)

var fields = []field{
	field{min: 0, max: 59}, // secondIndex
	field{min: 0, max: 59}, // minuteIndex
	field{min: 0, max: 23}, // hourIndex
	field{min: 1, max: 31}, // dayIndex
	field{min: 1, max: 12}, // monthIndex
	field{min: 0, max: 6},  // weekIndex
}

type field struct{ min, max uint8 }

func (f field) valid(v uint8) bool {
	return v >= f.min && v <= f.max
}

func sortUint64(vals []uint64) {
	sort.SliceStable(vals, func(i, j int) bool {
		return vals[i] < vals[j]
	})
}

// 分析单个数字域内容
//
// field 可以是以下格式：
//  *
//  n1-n2
//  n1,n2
//  n1-n2,n3-n4,n5
func parseField(f field, field string) (uint64, error) {
	if field == "*" {
		return any, nil
	}

	fields := strings.FieldsFunc(field, func(r rune) bool { return r == ',' })
	list := make([]uint64, 0, len(fields))

	for _, v := range fields {
		if len(v) <= 2 {
			n, err := strconv.ParseUint(v, 10, 8)
			if err != nil {
				return 0, err
			}

			if !f.valid(uint8(n)) {
				return 0, fmt.Errorf("值 %d 超出范围", n)
			}

			list = append(list, n)
			continue
		}

		index := strings.IndexByte(v, '-')
		if index >= 0 {
			v1 := v[:index]
			v2 := v[index+1:]
			n1, err := strconv.ParseUint(v1, 10, 8)
			if err != nil {
				return 0, err
			}
			n2, err := strconv.ParseUint(v2, 10, 8)
			if err != nil {
				return 0, err
			}

			if !f.valid(uint8(n1)) {
				return 0, fmt.Errorf("值 %d 超出范围", n1)
			}

			if !f.valid(uint8(n2)) {
				return 0, fmt.Errorf("值 %d 超出范围", n2)
			}

			list = append(list, intRange(n1, n2)...)
		}
	}

	sortUint64(list)
	for i := 1; i < len(list); i++ {
		if list[i] == list[i-1] {
			return 0, fmt.Errorf("重复的值 %d", list[i])
		}
	}

	var ret uint64
	for _, v := range list {
		ret |= (1 << v)
	}
	return ret, nil
}

// 获取一个范围内的整数
func intRange(start, end uint64) []uint64 {
	r := make([]uint64, 0, end-start+1)
	for i := start; i <= end; i++ {
		r = append(r, i)
	}

	return r
}
