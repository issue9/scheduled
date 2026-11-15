// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package schedulers

import (
	"testing"
	"time"

	"github.com/issue9/assert/v6"
)

// fixed 返回一个总是给出相同结果的调度器
func fixed(t time.Time) Scheduler {
	return SchedulerFunc(func(time.Time) time.Time { return t })
}

func TestEarliest(t *testing.T) {
	a := assert.New(t, false)

	now := time.Date(2026, time.March, 1, 5, 0, 0, 0, time.Local)

	a.PanicString(func() {
		Earliest()
	}, "参数 s 不能为空")

	t.Run("单个", func(t *testing.T) {
		a := assert.New(t, false)
		a.Equal(Earliest(Tick(time.Hour)).Next(now), now.Add(time.Hour))
	})

	t.Run("取最早", func(t *testing.T) {
		a := assert.New(t, false)
		s := Earliest(Tick(2*time.Hour), Tick(time.Hour), Tick(90*time.Minute))
		a.Equal(s.Next(now), now.Add(time.Hour))
	})

	t.Run("最早项不是第一个", func(t *testing.T) {
		a := assert.New(t, false)
		s := Earliest(Tick(time.Hour), Tick(time.Minute))
		a.Equal(s.Next(now), now.Add(time.Minute))
	})

	t.Run("忽略小于 last 的项", func(t *testing.T) {
		a := assert.New(t, false)
		s := Earliest(fixed(now.Add(-time.Hour)), Tick(time.Hour))
		a.Equal(s.Next(now), now.Add(time.Hour))
	})

	t.Run("忽略零值", func(t *testing.T) {
		a := assert.New(t, false)
		s := Earliest(fixed(time.Time{}), Tick(time.Hour))
		a.Equal(s.Next(now), now.Add(time.Hour))
	})

	t.Run("等于 last 的项", func(t *testing.T) {
		a := assert.New(t, false)
		s := Earliest(fixed(now), Tick(time.Hour))
		a.Equal(s.Next(now), now)
	})

	t.Run("全部为零值", func(t *testing.T) {
		a := assert.New(t, false)
		s := Earliest(fixed(time.Time{}), fixed(time.Time{}))
		a.True(s.Next(now).IsZero())
	})

	t.Run("全部小于 last", func(t *testing.T) {
		a := assert.New(t, false)
		s := Earliest(fixed(now.Add(-time.Hour)), fixed(now.Add(-2*time.Hour)))
		a.True(s.Next(now).IsZero())
	})

	t.Run("纯函数", func(t *testing.T) {
		a := assert.New(t, false)
		s := Earliest(Tick(2*time.Hour), Tick(time.Hour))

		first := s.Next(now)
		a.Equal(first, now.Add(time.Hour))
		a.Equal(s.Next(now), first, "相同的 last 应该返回相同的值")

		// 不同的 last 互不影响
		a.Equal(s.Next(now.Add(time.Hour)), now.Add(2*time.Hour))
		a.Equal(s.Next(now), first, "再回到原来的 last，结果不变")
	})
}
