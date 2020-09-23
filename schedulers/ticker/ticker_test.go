// SPDX-License-Identifier: MIT

package ticker

import (
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/scheduled/schedulers"
)

var _ schedulers.Scheduler = &ticker{}

func TestTicker(t *testing.T) {
	a := assert.New(t)

	a.Panic(func() {
		New(300*time.Microsecond, false)
	})

	s := New(5*time.Minute, false)
	a.NotNil(s)

	ticker, ok := s.(*ticker)
	a.True(ok).Equal(ticker.title, s.Title())

	now := time.Now()
	next1 := s.Next(now)
	a.Equal(next1.Unix(), now.Add(5*time.Minute).Unix())

	next2 := s.Next(next1)
	a.Equal(next2.Unix(), next1.Add(5*time.Minute).Unix())

	// 与 next1 相同的值调用，返回值也相同
	next3 := s.Next(now)
	a.Equal(next3.Unix(), next1.Unix())

	// imm == false

	s = New(5*time.Minute, true)
	a.NotNil(s)
	now = time.Now()
	next1 = s.Next(now)
	a.Equal(next1.Unix(), now.Unix())

	next2 = s.Next(next1)
	a.Equal(next2.Unix(), now.Add(5*time.Minute).Unix())
}
