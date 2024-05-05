// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package cron 实现了 [cron] 表达式的 [schedulers.Scheduler] 接口
//
// [cron]: https://zh.wikipedia.org/wiki/Cron
package cron

import (
	"strings"
	"time"

	"github.com/issue9/localeutil"
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
	// 依次保存着 cron 语法中各个字段解析后的内容
	data []fields
	loc  *time.Location
}

// Parse 根据 spec 初始化 [schedulers.Scheduler]
//
// spec 表示 crontab 的格式
//
// 区分大小写，支持秒，其格式如下：
//
//	! * * * * * *
//	  | | | | | |
//	  | | | | | --- 星期
//	  | | | | ----- 月
//	  | | | ------- 日
//	  | | --------- 小时
//	  | ----------- 分
//	  ------------- 秒
//
// 星期与日若同时存在，则以或的形式组合。！用于使 go fmt 不会自动格式化内容，无实际意义。
//
// 支持以下符号：
//   - - 表示范围
//   - , 表示和
//
// 同时支持以下便捷指令：
//
//	@reboot:   启动时执行一次
//	@yearly:   0 0 0 1 1 *
//	@annually: 0 0 0 1 1 *
//	@monthly:  0 0 0 1 * *
//	@weekly:   0 0 0 * * 0
//	@daily:    0 0 0 * * *
//	@midnight: 0 0 0 * * *
//	@hourly:   0 0 * * * *
func Parse(spec string, loc *time.Location) (schedulers.Scheduler, error) {
	switch {
	case spec == "":
		return nil, syntaxError(localeutil.Phrase("can not be empty"))
	case spec == "@reboot":
		return at.At(time.Now()), nil
	case spec[0] == '@':
		d, found := direct[spec]
		if !found {
			return nil, syntaxError(localeutil.Phrase("invalid direct %s", spec))
		}
		spec = d
	}

	fs := strings.Fields(spec)
	if len(fs) != indexSize {
		return nil, syntaxError(localeutil.Phrase("incorrect length"))
	}

	c := &cron{
		data: make([]fields, indexSize),
		loc:  loc,
	}

	allAny := true // 是否所有字段都是 asterisk
	for i, field := range fs {
		vals, err := parseField(i, field)
		if err != nil {
			return nil, err
		}

		if allAny && vals != asterisk {
			allAny = false
		}

		if !allAny && vals == asterisk {
			vals = step
		}

		c.data[i] = vals
	}

	if allAny { // 所有项都为 *
		return nil, syntaxError(localeutil.Phrase("all items are asterisk"))
	}

	return c, nil
}

func syntaxError(s localeutil.Stringer) error {
	return localeutil.Error("cron syntax error %s", s)
}
