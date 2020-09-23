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

	// 格式错误
	s, err := At("2019-01-02 13:14:")
	a.Error(err).Nil(s)

	// 早于当前时间
	now := time.Now()
	tt := "2019-01-02 13:14:15"
	s, err = At(tt)
	a.NotError(err).NotNil(s)
	a.True(s.Next(now).Before(now)).
		Equal(s.Next(now), zero) // 多次获取，返回零值
	a.Equal(s.Title(), tt)

	loc := time.FixedZone("UTC+8", 8*60*60)
	ttt, err := time.ParseInLocation(Layout, tt, time.UTC)
	a.NotError(err)
	s, err = At(tt)
	a.NotError(err).NotNil(s)
	next := s.Next(time.Now().In(loc)) // 变成 8 时区，小于零时区的 loc
	a.True(next.Before(ttt))
}
