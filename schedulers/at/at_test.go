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
		True(s.Next(now).IsZero()) // 多次获取，返回零值
}
