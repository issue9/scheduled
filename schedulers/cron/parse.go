// SPDX-License-Identifier: MIT

package cron

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/issue9/scheduled/schedulers"
	"github.com/issue9/scheduled/schedulers/at"
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

func (b bound) valid(v int) bool {
	return v >= b.min && v <= b.max
}

// Parse 根据 spec 初始化 schedulers.Scheduler
//
// spec 的格式如下：
//  * * * * * *
//  | | | | | |
//  | | | | | --- 星期
//  | | | | ----- 月
//  | | | ------- 日
//  | | --------- 小时
//  | ----------- 分
//  ------------- 秒
//
// 星期与日若同时存在，则以或的形式组合。
//
// 支持以下符号：
//  - 表示范围
//  , 表示和
//
// 同时支持以下便捷指令：
//  @reboot:   启动时执行一次
//  @yearly:   0 0 0 1 1 *
//  @annually: 0 0 0 1 1 *
//  @monthly:  0 0 0 1 * *
//  @weekly:   0 0 0 * * 0
//  @daily:    0 0 0 * * *
//  @midnight: 0 0 0 * * *
//  @hourly:   0 0 * * * *
func Parse(spec string) (schedulers.Scheduler, error) {
	switch {
	case spec == "":
		return nil, errors.New("参数 spec 不能为空")
	case spec == "@reboot":
		return at.At(time.Time{}.Format(at.Layout))
	case spec[0] == '@':
		d, found := direct[spec]
		if !found {
			return nil, errors.New("未找到指令:" + spec)
		}
		spec = d
	}

	fs := strings.Fields(spec)
	if len(fs) != indexSize {
		return nil, errors.New("长度不正确")
	}

	c := &cron{
		title: spec,
		data:  make([]uint64, indexSize),
	}

	allAny := true // 是否所有字段都是 any
	for i, f := range fs {
		vals, err := parseField(i, f)
		if err != nil {
			return nil, err
		}

		if allAny && vals != any {
			allAny = false
		}

		if !allAny && vals == any {
			vals = step
		}

		c.data[i] = vals
	}

	if allAny { // 所有项都为 *
		return nil, errors.New("所有项都为 *")
	}

	return c, nil
}

// 分析单个数字域内容
//
// field 可以是以下格式：
//  *
//  n1-n2
//  n1,n2
//  n1-n2,n3-n4,n5
func parseField(typ int, field string) (uint64, error) {
	if field == "*" {
		return any, nil
	}

	fields := strings.FieldsFunc(field, func(r rune) bool { return r == ',' })
	list := make([]uint64, 0, len(fields))

	b := bounds[typ]
	for _, v := range fields {
		if len(v) <= 2 { // 少于 3 个字符，说明不可能有特殊字符。
			n, err := strconv.Atoi(v)
			if err != nil {
				return 0, err
			}

			if !b.valid(n) {
				return 0, fmt.Errorf("值 %d 超出范围：[%d,%d]", n, b.min, b.max)
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
				return 0, fmt.Errorf("值 %d 超出范围：[%d,%d]", n1, b.min, b.max)
			}

			if !b.valid(n2) {
				return 0, fmt.Errorf("值 %d 超出范围：[%d,%d]", n2, b.min, b.max)
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

func sortUint64(vals []uint64) {
	sort.SliceStable(vals, func(i, j int) bool {
		return vals[i] < vals[j]
	})
}
