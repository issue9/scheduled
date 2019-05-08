// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package cron 实现了 cron 表达式的 Scheduler 接口
package cron

import "time"

// 表示 cron 语法表达式中的顺序
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
	//"@reboot":   "TODO",
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
	//
	// 最长的秒数，最多 60 位，正好可以使用一个 uint64 保存，
	// 其它类型也各自占一个字段长度。
	//
	// 其中每个字段中，从 0 位到高位，每一位表示一个值，比如在秒字段中，
	//  0,1,7 表示为 ...10000011
	// 如果是月份这种从 1 开始的，则其第一位永远是 0
	data []uint64

	next  time.Time
	title string
}

// Title 获取标题名称
func (c *cron) Title() string {
	return c.title
}
