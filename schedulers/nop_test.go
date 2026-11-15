// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package schedulers

import (
	"testing"
	"time"

	"github.com/issue9/assert/v6"
)

func TestNop(t *testing.T) {
	a := assert.New(t, false)

	now := time.Now()

	a.True(Nop().Next(now).IsZero()).
		True(Nop().Next(now.Add(time.Hour)).IsZero()).
		True(Nop().Next(now.Add(-time.Hour)).IsZero())
}
