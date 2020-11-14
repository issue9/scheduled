// SPDX-License-Identifier: MIT

package at

import (
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/scheduled/schedulers"
)

var _ schedulers.Scheduler = &scheduler{}

func TestAt(t *testing.T) {
	a := assert.New(t)

	// 早于当前时间
	now := time.Now()
	tt := now.Add(-time.Hour)
	s := At(tt)
	a.NotNil(s)
	a.True(s.Next(now).Before(now)).
		Equal(s.Next(now), zero) // 多次获取，返回零值
	a.Equal(s.Title(), tt.Format(layout))

	ttt, err := time.ParseInLocation(layout, tt.Format(layout), time.UTC)
	a.NotError(err)
	s = At(tt)
	a.NotNil(s)
	loc := time.FixedZone("UTC+8", 8*60*60)
	next := s.Next(time.Now().In(loc)) // 变成 8 时区，小于零时区的 loc
	a.True(next.Before(ttt))
}
