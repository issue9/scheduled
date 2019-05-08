// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ticker

import (
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestTicker(t *testing.T) {
	a := assert.New(t)

	ticker := New(5 * time.Minute)
	a.Equal(ticker.title, ticker.Title())

	now := time.Now()
	last := ticker.Next(now)
	a.Equal(last, now.Add(5*time.Minute))

	last2 := ticker.Next(last)
	a.Equal(last2, last.Add(5*time.Minute))

	last3 := ticker.Next(now)
	a.Equal(last3, now.Add(5*time.Minute))
}
