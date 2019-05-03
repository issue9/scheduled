// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import (
	"time"

	"github.com/issue9/cron/internal/expr"
	"github.com/issue9/cron/internal/ticker"
)

// Nexter 用于生成下一次定时器的时间
type Nexter interface {
	// 生成下一次定时器需要的时间。
	// 相对于 last 时间。
	Next(last time.Time) time.Time

	// Title 生成名称
	Title() string
}

// NewTicker 添加一个新的定时任务
func (c *Cron) NewTicker(name string, f JobFunc, dur time.Duration) {
	c.New(name, f, ticker.New(dur))
}

// NewExpr 使用 cron 表示式新建一个定时任务
//
// spec 的值可以是：
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
//  @yearly:   0 0 0 1 1 *
//  @annually: 0 0 0 1 1 *
//  @monthly:  0 0 0 1 * *
//  @weekly:   0 0 0 * * 0
//  @daily:    0 0 0 * * *
//  @midnight: 0 0 0 * * *
//  @hourly:   0 0 * * * *
func (c *Cron) NewExpr(name string, f JobFunc, spec string) error {
	next, err := expr.Parse(spec)
	if err != nil {
		return err
	}

	c.New(name, f, next)
	return nil
}
