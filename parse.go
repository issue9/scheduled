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

// TODO 可以解析以下内容
//
// @reboot        Run once, at startup.
// @yearly         Run once a year, "0 0 1 1 *".
// @annually      (same as @yearly)
// @monthly       Run once a month, "0 0 1 * *".
// @weekly        Run once a week, "0 0 * * 0".
// @daily           Run once a day, "0 0 * * *".
// @midnight      (same as @daily)
// @hourly         Run once an hour, "0 * * * *".

func sortUint64(vals []uint64) {
	sort.SliceStable(vals, func(i, j int) bool {
		return vals[i] < vals[j]
	})
}

func parseExpr(spec string) (*Expr, error) {
	fs := strings.Fields(spec)
	if len(fs) != typeSize {
		return nil, errors.New("长度不正确")
	}

	e := &Expr{
		title:      spec,
		data:       make([]uint64, typeSize),
		startIndex: -1,
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

		if e.startIndex == -1 && !allAny {
			e.startIndex = i
		}

		e.data[i] = vals
	}

	if e.startIndex == -1 {
		e.startIndex = 0
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
