// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package ticker 时间段固定的定时器，功能与 time.Ticker 相同。
package ticker

import (
	"fmt"
	"time"

	"github.com/issue9/cron/schedule"
)

type ticker struct {
	dur   time.Duration
	title string
}

// New 声明一个固定时间段的定时任务
func New(d time.Duration) schedule.Scheduler {
	return &ticker{
		dur:   d,
		title: fmt.Sprintf("每隔 %s", d),
	}
}

func (t *ticker) Next(last time.Time) time.Time {
	return last.Add(t.dur)
}

func (t *ticker) Title() string {
	return t.title
}
