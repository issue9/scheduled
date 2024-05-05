// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package cron

import (
	"math/bits"
	"strconv"
	"strings"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
)

// 表示 cron 语法中每一个字段的数据
//
// 最长的秒数，最多 60 位，正好可以使用一个 uint64 保存，
// 其它类型也各自占一个字段长度。
//
// 其中每个字段中，从 0 位到高位，每一位表示一个值，比如在秒字段中，
//
//	0,1,7 表示为 ...10000011
//
// 如果是月份这种从 1 开始的，则其第一位永远是 0
type fields uint64

const (
	// asterisk 和 step 是两个特殊的标记位，需要大于 60（所有字段中，秒数最大，
	// 但不会超过 60）

	// asterisk 表示当前字段可以是任意值，即对值不做任意要求，
	// 甚至可以一直是相同的值，也不会做累加，即 * 符号表示的值。
	asterisk fields = 1 << 61

	// step 表示当前字段是允许范围内的所有值。
	// 每次计算时，按其当前值加 1 即可。
	step fields = 1 << 62
)

var bounds = []bound{
	{min: 0, max: 59}, // secondIndex
	{min: 0, max: 59}, // minuteIndex
	{min: 0, max: 23}, // hourIndex
	{min: 1, max: 31}, // dayIndex
	{min: 1, max: 12}, // monthIndex
	{min: 0, max: 7},  // weekIndex
}

type bound struct{ min, max int }

func (b bound) valid(v int) bool { return v >= b.min && v <= b.max }

// 获取 fields 中与 curr 最近的下一个值
//
// curr 当前的时间值；
// typ 字段类型；
// greater 是否必须要大于 curr 这个值；
// val 返回计算后的最近一个时间值；
// c 是否需要下一个值进位。
func (fs fields) next(curr int, b bound, greater bool) (val int, c bool) {
	if fs == asterisk { // asterisk 表示对当前值没有要求，不需要增加值。
		return curr, greater
	} else if fs == step {
		if greater {
			curr++
		}

		if curr > b.max {
			return b.min, true
		}
		return curr, false
	}

	for i := curr; i <= b.max; i++ {
		if ((uint64(1) << uint64(i)) & uint64(fs)) <= 0 { // 该位未被设置为 1
			continue
		}

		if i > curr {
			return i, false
		} else if i == curr && !greater {
			return i, false
		}
	}

	// 大于当前列表的最大值，则返回列表中的最小值，并设置进位标记
	return bits.TrailingZeros64(uint64(fs)), true
}

// 分析单个数字域内容
//
// field 可以是以下格式：
//
//	*
//	n1-n2
//	n1,n2
//	n1-n2,n3-n4,n5
func parseField(typ int, field string) (fields, error) {
	if field == "*" {
		return asterisk, nil
	}

	fs := strings.FieldsFunc(field, func(r rune) bool { return r == ',' })
	list := make([]uint64, 0, len(fs))

	b := bounds[typ]
	for _, v := range fs {
		if len(v) <= 2 { // 少于 3 个字符，说明不可能有特殊字符。
			n, err := strconv.Atoi(v)
			if err != nil {
				return 0, err
			}

			if !b.valid(n) {
				return 0, syntaxError(localeutil.Phrase("the value %d out of range [%d,%d]", n, b.min, b.max))
			}

			// 星期中的 7 替换成 0
			if typ == weekIndex && n == b.max {
				n = b.min
			}

			list = append(list, uint64(n))
			continue
		}

		index := strings.IndexByte(v, '-')
		if index >= 0 {
			n1, err := strconv.Atoi(v[:index])
			if err != nil {
				return 0, err
			}
			n2, err := strconv.Atoi(v[index+1:])
			if err != nil {
				return 0, err
			}

			if !b.valid(n1) {
				return 0, syntaxError(localeutil.Phrase("the value %d out of range [%d,%d]", n1, b.min, b.max))
			}

			if !b.valid(n2) {
				return 0, syntaxError(localeutil.Phrase("the value %d out of range [%d,%d]", n2, b.min, b.max))
			}

			for i := n1; i <= n2; i++ {
				if typ == weekIndex && i == b.max {
					list = append(list, uint64(b.min))
				} else {
					list = append(list, uint64(i))
				}
			}
		}
	}

	if indexes := sliceutil.Dup(list, func(i, j uint64) bool { return i == j }); len(indexes) > 0 {
		return 0, syntaxError(localeutil.Phrase("duplicate value %d", list[indexes[0]]))
	}

	var ret fields
	for _, v := range list {
		ret |= (1 << v)
	}
	return ret, nil
}
