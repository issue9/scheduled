// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import (
	"testing"

	"github.com/issue9/assert"
)

func TestCron_NewExpr(t *testing.T) {
	a := assert.New(t)

	c := New()
	a.NotError(c.NewExpr("test", nil, "* * * 3-7 * *"))
	a.Error(c.NewExpr("test", nil, "* * * 3-7a * *"))
}
