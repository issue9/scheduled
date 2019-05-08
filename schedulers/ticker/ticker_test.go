// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

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

	s := New(5 * time.Minute)

	ticker, ok := s.(*ticker)
	a.True(ok).Equal(ticker.title, s.Title())

	now := time.Now()
	last := s.Next(now)
	a.Equal(last, now.Add(5*time.Minute))

	last2 := s.Next(last)
	a.Equal(last2, last.Add(5*time.Minute))

	last3 := s.Next(now)
	a.Equal(last3, now.Add(5*time.Minute))
}
