// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package at

import (
	"testing"
	"time"

	"github.com/issue9/assert/v6"
)

func TestAt(t *testing.T) {
	a := assert.New(t, false)

	now := time.Now()

	s := At(now.Add(-time.Hour))
	a.NotNil(s)
	a.True(s.Next(now).IsZero()).
		True(s.Next(now).IsZero()) // 多次获取，返回相同值。

	s = At(now.Add(time.Hour))
	a.NotNil(s)
	a.True(s.Next(now).Equal(now.Add(time.Hour))).
		True(s.Next(now).Equal(now.Add(time.Hour))) // 多次获取，返回相同值。

	// 恰好等于参考时间
	s = At(now)
	a.True(s.Next(now).IsZero())

	// 零值：不执行
	s = At(time.Time{})
	a.True(s.Next(now).IsZero())

	// 同一个实例用不同的 now 交替调用，结果互不影响

	s = At(now.Add(time.Hour))
	a.Equal(s.Next(now), now.Add(time.Hour))
	a.True(s.Next(now.Add(2 * time.Hour)).IsZero())
	a.Equal(s.Next(now), now.Add(time.Hour)) // 再回到原来的 now，结果不变
}
