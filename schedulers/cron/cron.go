// SPDX-License-Identifier: MIT

// Package cron 实现了 cron 表达式的 Scheduler 接口
package cron

import (
	"errors"
	"strings"
	"time"

	"github.com/issue9/scheduled/schedulers"
	"github.com/issue9/scheduled/schedulers/at"
)

// 表示 cron.data 中各个元素的索引值
const (
	secondIndex = iota
	minuteIndex
	hourIndex
	dayIndex
	monthIndex
	weekIndex
	indexSize
)

// 常用的便捷指令
var direct = map[string]string{
	"@yearly":   "0 0 0 1 1 *",
	"@annually": "0 0 0 1 1 *",
	"@monthly":  "0 0 0 1 * *",
	"@weekly":   "0 0 0 * * 0",
	"@daily":    "0 0 0 * * *",
	"@midnight": "0 0 0 * * *",
	"@hourly":   "0 0 * * * *",
}

type cron struct {
	// 依次保存着 cron 语法中各个字段解析后的内容。
	data []fields
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
		return at.At(time.Time{}), nil
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

	c := &cron{data: make([]fields, indexSize)}

	allAny := true // 是否所有字段都是 any
	for i, field := range fs {
		vals, err := parseField(i, field)
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
