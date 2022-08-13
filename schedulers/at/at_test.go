// SPDX-License-Identifier: MIT

package at

import (
	"testing"
	"time"

	"github.com/issue9/assert/v3"

	"github.com/issue9/scheduled/schedulers"
)

var _ schedulers.Scheduler = &scheduler{}

func TestAt(t *testing.T) {
	a := assert.New(t, false)

	now := time.Now()

	s := At(now.Add(-time.Hour))
	a.NotNil(s)
	a.True(s.Next(now).Before(now)).
		True(s.Next(now).IsZero()) // 多次获取，返回零值

	s = At(now.Add(time.Hour))
	a.NotNil(s)
	a.True(s.Next(now).After(now)).
		True(s.Next(now).IsZero()) // 多次获取，返回零值
}
