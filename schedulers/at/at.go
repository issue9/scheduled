// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package at 提供类似于 at 指令的定时器
package at

import (
	"time"

	"github.com/issue9/scheduled/schedulers"
)

const layout = "2006-01-02 15:04:05"

type scheduler struct {
	title string
	at    time.Time
}

func Parse(t string) (schedulers.Scheduler, error) {
	at, err := time.Parse(layout, t)
	if err != nil {
		return nil, err
	}

	return At(at), nil
}

func At(t time.Time) schedulers.Scheduler {
	return &scheduler{
		title: t.Format(layout),
		at:    t,
	}
}

func (s *scheduler) Title() string {
	return s.title
}

func (s *scheduler) Next(last time.Time) time.Time {
	return s.at
}
