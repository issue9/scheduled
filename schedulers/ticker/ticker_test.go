// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package ticker

import (
	"testing"
	"time"

	"github.com/issue9/assert/v6"
)

func TestTick(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		Tick(300 * time.Microsecond)
	}, "参数 d 的值必须在 1 秒以上")

	s := Tick(5 * time.Minute)
	a.NotNil(s)

	now := time.Now()
	next1 := s.Next(now)
	a.Equal(next1.Unix(), now.Add(5*time.Minute).Unix())

	next2 := s.Next(next1)
	a.Equal(next2.Unix(), next1.Add(5*time.Minute).Unix())

	// 纯函数：相同的输入总是得到相同的返回值。
	a.Equal(s.Next(now).Unix(), next1.Unix()).
		Equal(s.Next(now).Unix(), next1.Unix())

	// 一个实例可以被多个任务共享。
	other := Tick(5 * time.Minute)
	a.Equal(other.Next(now).Unix(), next1.Unix())
}
