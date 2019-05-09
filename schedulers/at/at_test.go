// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package at

import (
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/scheduled/schedulers"
)

var _ schedulers.Scheduler = &scheduler{}

func TestParse(t *testing.T) {
	a := assert.New(t)

	// 格式错误
	s, err := Parse("2019-01-02 13:14:")
	a.Error(err).Nil(s)

	// 早于当前时间
	now := time.Now()
	tt := "2019-01-02 13:14:15"
	s, err = Parse(tt)
	a.NotError(err).NotNil(s)
	a.True(s.Next(now).Before(now)).
		Equal(s.Next(now), s.Next(now.Add(10*time.Hour))) // 多次获取，值是相同的
	a.Equal(s.Title(), tt)

}

func TestAt(t *testing.T) {
	a := assert.New(t)

	zero := time.Time{}
	s := At(zero)
	a.NotNil(s)

	// 零值，永远返回零值。
	a.True(s.Next(time.Now()).IsZero())
	a.True(s.Next(time.Now()).IsZero())
}
