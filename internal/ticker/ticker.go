// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package ticker 时间段固定的定时器，功能与 time.Ticker 相同。
package ticker

import (
	"fmt"
	"time"
)

// Ticker 固定时间段的定时器
type Ticker struct {
	dur   time.Duration
	title string
}

// New 声明一个固定时间段的定时任务
func New(d time.Duration) *Ticker {
	return &Ticker{
		dur:   d,
		title: fmt.Sprintf("每隔 %s", d),
	}
}

// Next 实现 Nexter.Next 接口函数
func (t *Ticker) Next(last time.Time) time.Time {
	return last.Add(t.dur)
}

// Title 实现 Nexter.Title 接口函数
func (t *Ticker) Title() string {
	return t.title
}
