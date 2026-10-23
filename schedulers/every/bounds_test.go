// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package every

import (
	"testing"

	"github.com/issue9/assert/v6"
)

func TestBound_valid(t *testing.T) {
	a := assert.New(t, false)

	a.True(minuteBound.valid(0)).
		True(minuteBound.valid(1)).
		True(minuteBound.valid(30)).
		True(minuteBound.valid(59))

	a.False(minuteBound.valid(-1)).
		False(minuteBound.valid(60))

	a.True(hourBound.valid(0)).
		True(hourBound.valid(23))

	a.False(hourBound.valid(-1)).
		False(hourBound.valid(24))
}
