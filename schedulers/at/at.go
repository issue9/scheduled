// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package at 提供类似于 at 指令的定时器
package at

import (
	"time"

	"github.com/issue9/scheduled/schedulers"
)

// Layout Parse 解析时间的格式。同时也是 Title 返回的格式。
const Layout = "2006-01-02 15:04:05"

type scheduler struct {
	title string
	at    time.Time
}

// Parse 返回只在指定时间执行一次的调度器
//
// 时间从 Parse 中获取。
func Parse(t string) (schedulers.Scheduler, error) {
	at, err := time.Parse(Layout, t)
	if err != nil {
		return nil, err
	}

	return At(at), nil
}

// At 返回只在指定时间执行一次的调度器
func At(t time.Time) schedulers.Scheduler {
	return &scheduler{
		title: t.Format(Layout),
		at:    t,
	}
}

func (s *scheduler) Title() string {
	return s.title
}

func (s *scheduler) Next(last time.Time) time.Time {
	return s.at
}
