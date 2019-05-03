// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

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
