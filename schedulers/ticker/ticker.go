// SPDX-License-Identifier: MIT

// Package ticker 时间段固定的定时器
package ticker

import (
	"errors"
	"time"

	"github.com/issue9/scheduled/schedulers"
)

type ticker struct {
	dur time.Duration
	imm bool
}

// New 声明一个固定时间段的定时任务
//
// imm 是否立即执行一次任务，如果为 true，
// 则会在第一次调用 Next 时返回当前时间。
func New(d time.Duration, imm bool) (schedulers.Scheduler, error) {
	if d < time.Second {
		return nil, errors.New("参数 d 的值必须在 1 秒以上")
	}

	return &ticker{
		dur: d,
		imm: imm,
	}, nil
}

func (t *ticker) Next(last time.Time) time.Time {
	if t.imm {
		t.imm = false
		return time.Now().In(last.Location())
	}

	return last.Add(t.dur)
}
