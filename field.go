// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import (
	"errors"
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

// 常用的便捷指令
var direct = map[string]string{
	//"@reboot":   "TODO",
	"@yearly":   "0 0 0 1 1 *",
	"@annually": "0 0 0 1 1 *",
	"@monthly":  "0 0 0 1 * *",
	"@weekly":   "0 0 0 * * 0",
	"@daily":    "0 0 0 * * *",
	"@midnight": "0 0 0 * * *",
	"@hourly":   "0 0 * * * *",
}

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

// curr 当前的时间值；
// list 可用的时间值；
// carry 前一个数值是否已经进位；
// val 返回计算后的最近一个时间值；
// c 是否需要一个值进位。
func (f field) next(curr uint8, list uint64, carry bool) (val uint8, c bool) {
	if list == any {
		return curr, carry
	}

	if list == step {
		if carry {
			curr++
		}

		if curr > f.max {
			return f.min, true
		}
		return curr, false
	}

	var min uint8
	var hasMin bool
	for i := f.min; i <= f.max; i++ {
		if ((uint64(1) << i) & list) <= 0 { // 该位未被设置为 1
			continue
		}

		if i > curr {
			return i, false
		}

		if !hasMin {
			min = i
			hasMin = true
		}

		if i == curr {
			if !carry {
				return i, false
			}
			carry = false
		}
	} // end for

	// 大于当前列表的最大值，则返回列表中的最小值，并设置进位标记
	return min, true
}

func sortUint64(vals []uint64) {
	sort.SliceStable(vals, func(i, j int) bool {
		return vals[i] < vals[j]
	})
}

func parseExpr(spec string) (*expr, error) {
	if spec == "" {
		return nil, errors.New("参数 spec 错误")
	}

	if spec[0] == '@' {
		d, found := direct[spec]
		if !found {
			return nil, errors.New("款找到指令" + spec)
		}
		spec = d
	}

	fs := strings.Fields(spec)
	if len(fs) != typeSize {
		return nil, errors.New("长度不正确")
	}

	e := &expr{
		title: spec,
		data:  make([]uint64, typeSize),
	}

	allAny := true
	for i, f := range fs {
		vals, err := parseField(fields[i], f)
		if err != nil {
			return nil, err
		}

		if allAny && vals != any {
			allAny = false
		}

		if !allAny && vals == any {
			vals = step
		}

		e.data[i] = vals
	}

	if allAny { // 所有项都为 *
		return nil, errors.New("所有项都为 *")
	}

	return e, nil
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
