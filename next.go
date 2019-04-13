// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import "time"

// Nexter 用于生成下一次定时器的时间
type Nexter interface {
	// 生成下一次定时器需要的时间。
	// 相对于 last 时间。
	Next(last time.Time) time.Time
}

// 固定时间段的定时器
type duration struct {
	dur time.Duration
}

func newDuration(d time.Duration) Nexter {
	return &duration{
		dur: d,
	}
}

func (d *duration) Next(last time.Time) time.Time {
	return last.Add(d.dur)
}
