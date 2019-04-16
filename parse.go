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

type bound struct{ min, max uint8 }

// 每种类型的值取值范围
var bounds = []bound{
	bound{min: 0, max: 59}, // secondIndex
	bound{min: 0, max: 59}, // minuteIndex
	bound{min: 0, max: 23}, // hourIndex
	bound{min: 1, max: 31}, // dayIndex
	bound{min: 1, max: 12}, // monthIndex
	bound{min: 0, max: 6},  // wwekIndex
}

func (b bound) valid(v uint8) bool {
	return v >= b.min && v <= b.max
}

func sortUint8(vals []uint8) {
	sort.SliceStable(vals, func(i, j int) bool {
		return vals[i] < vals[j]
	})
}

// NewExpr 新建表达式定时器
//
// expr 的值可以是：
//  * * * * * *  cmd
//  | | | | | |
//  | | | | | --- 星期
//  | | | | ----- 月
//  | | | ------- 日
//  | | --------- 小时
//  | ----------- 分
//  ------------- 秒
//
// 支持以下符号：
//  - 表示范围
//  , 表示和
func (c *Cron) NewExpr(name string, f JobFunc, expr string) error {
	next, err := parseExpr(expr)
	if err != nil {
		return err
	}

	c.New(name, f, next)
	return nil
}

func parseExpr(spec string) (*Expr, error) {
	fields := strings.Fields(spec)
	if len(fields) != 6 {
		return nil, errors.New("长度不正确")
	}

	e := &Expr{
		title: spec,
		data:  make([][]uint8, 6, 6),
	}

	for i, f := range fields {
		vals, err := parseField(f)
		if err != nil {
			return nil, err
		}

		if vals != nil {
			bound := bounds[i]
			if !bound.valid(vals[0]) || !bound.valid(vals[len(vals)-1]) {
				return nil, errors.New("值超出范围")
			}
		}

		e.data[i] = vals
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
func parseField(field string) ([]uint8, error) {
	if field == "*" {
		return nil, nil
	}

	fields := strings.FieldsFunc(field, func(r rune) bool { return r == ',' })
	ret := make([]uint8, 0, len(fields))

	for _, v := range fields {
		if len(v) <= 2 {
			n, err := strconv.ParseUint(v, 10, 8)
			if err != nil {
				return nil, err
			}
			ret = append(ret, uint8(n))
			continue
		}

		index := strings.IndexByte(v, '-')
		if index >= 0 {
			v1 := v[:index]
			v2 := v[index+1:]
			n1, err := strconv.ParseUint(v1, 10, 8)
			if err != nil {
				return nil, err
			}
			n2, err := strconv.ParseUint(v2, 10, 8)
			if err != nil {
				return nil, err
			}

			ret = append(ret, intRange(uint8(n1), uint8(n2))...)
		}
	}

	// 排序，查重
	sortUint8(ret)
	for i := 1; i < len(ret); i++ {
		if ret[i] == ret[i-1] {
			return nil, fmt.Errorf("存在相同的值 %d", ret[i])
		}
	}
	return ret, nil
}

// 获取一个范围内的整数
func intRange(start, end uint8) []uint8 {
	r := make([]uint8, 0, end-start+1)
	for i := start; i <= end; i++ {
		r = append(r, i)
	}

	return r
}
