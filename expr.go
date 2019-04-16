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
	"time"
)

// 该值的顺序与 cron 中语法的顺序相同
const (
	secondIndex = iota
	minuteIndex
	hourIndex
	dayIndex
	monthIndex
	weekIndex
)

type bound struct{ min, max uint8 }

// Expr 表达式的分析结果
type Expr struct {
	data [][]uint8

	next  time.Time
	title string
}

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

// Title 获取标题名称
func (e *Expr) Title() string {
	return e.title
}

// Next 计算下个时间点，相对于 last
func (e *Expr) Next(last time.Time) time.Time {
	if e.next.After(last) {
		return e.next
	}

	second, carry := next(uint8(last.Second()), e.data[secondIndex], false)
	minute, carry := next(uint8(last.Minute()), e.data[minuteIndex], carry)
	hour, carry := next(uint8(last.Hour()), e.data[hourIndex], carry)

	var day int
	if e.data[weekIndex] != nil { // 除非指定了星期，否则永远按照日期来
		weekday, _ := next(uint8(last.Weekday()), e.data[weekIndex], carry)
		dur := weekday - int(last.Weekday()) // 相差的天数
		day = dur + last.Day()
	} else {
		day, carry = next(uint8(last.Day()), e.data[dayIndex], carry)
	}

	month, carry := next(uint8(last.Month()), e.data[monthIndex], carry)
	year := last.Year()
	if carry {
		year++
	}

	// 由于月份中的天数不固定，还得计算该天数是否存在于当前月分
	for {
		days := getMonthDays(time.Month(month), year)
		if day <= days { // 天数存在于当前月，则退出循环
			break
		}

		month, carry = next(uint8(month), e.data[monthIndex], false)
		if carry {
			year++
		}
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, last.Location())
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

// 获取指定月份的天数
func getMonthDays(month time.Month, year int) int {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	return last.Day()
}

// curr 当前的时间值；
// list 可用的时间值；
// carry 是否需要当前时间进位；
// val 返回计算后的最近一个时间值；
// c 是否已经进位。
func next(curr uint8, list []uint8, carry bool) (val int, c bool) {
	if list == nil {
		if carry {
			curr++
		}
		return int(curr), false
	}

	for _, item := range list {
		switch {
		case item == curr: // 存在与当前值相同的值
			if !carry {
				return int(item), false
			}
		case item > curr:
			return int(item), false
		}
	}

	// 大于当前列表的最大值，则返回列表中的最大值，则设置进位标记
	return int(list[0]), true
}
