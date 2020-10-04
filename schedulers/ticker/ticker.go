// SPDX-License-Identifier: MIT

// Package ticker 时间段固定的定时器
package ticker

import (
	"fmt"
	"time"

	"github.com/issue9/scheduled/schedulers"
)

type ticker struct {
	dur   time.Duration
	title string
	imm   bool
}

// New 声明一个固定时间段的定时任务
//
// imm 是否立即执行一次任务，如果为 true，
// 则会在第一次调用 last 时返回当前时间。
func New(d time.Duration, imm bool) schedulers.Scheduler {
	if d < time.Second {
		panic("参数 d 的值必须在 1 秒以上")
	}

	return &ticker{
		dur:   d,
		title: fmt.Sprintf("每隔 %s", d),
		imm:   imm,
	}
}

func (t *ticker) Next(last time.Time) time.Time {
	if t.imm {
		t.imm = false
		return time.Now().In(last.Location())
	}

	return last.Add(t.dur)
}

func (t *ticker) Title() string {
	return t.title
}
