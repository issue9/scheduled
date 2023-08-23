// SPDX-License-Identifier: MIT

package ticker

import (
	"testing"
	"time"

	"github.com/issue9/assert/v3"
)

func TestTicker(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		Tick(300*time.Microsecond, false)
	}, "参数 d 的值必须在 1 秒以上")

	s := Tick(5*time.Minute, false)
	a.NotNil(s)

	now := time.Now()
	next1 := s.Next(now)
	a.Equal(next1.Unix(), now.Add(5*time.Minute).Unix())

	next2 := s.Next(next1)
	a.Equal(next2.Unix(), next1.Add(5*time.Minute).Unix())

	// 与 next1 相同的值调用，返回值也相同
	next3 := s.Next(now)
	a.Equal(next3.Unix(), next1.Unix())

	// imm == false

	s = Tick(5*time.Minute, true)
	a.NotNil(s)
	now = time.Now()
	next1 = s.Next(now)
	a.Equal(next1.Unix(), now.Unix())

	next2 = s.Next(next1)
	a.Equal(next2.Unix(), now.Add(5*time.Minute).Unix())
}
