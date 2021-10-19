// SPDX-License-Identifier: MIT

// Package at 提供类似于 at 指令的定时器
package at

import (
	"time"

	"github.com/issue9/scheduled/schedulers"
)

type scheduler struct {
	t time.Time
}

// At 返回只在指定时间执行一次的调度器
func At(t time.Time) schedulers.Scheduler { return &scheduler{t: t} }

func (s *scheduler) Next(last time.Time) time.Time {
	if s.t.IsZero() {
		return s.t
	}

	ret := s.t
	s.t = time.Time{}
	return ret.In(last.Location())
}
