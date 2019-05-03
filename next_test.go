// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/cron/internal/expr"
	"github.com/issue9/cron/internal/ticker"
)

var (
	_ Nexter = &ticker.Ticker{}
	_ Nexter = &expr.Expr{}
)

func TestCron_NewExpr(t *testing.T) {
	a := assert.New(t)

	c := New()
	a.NotError(c.NewExpr("test", nil, "* * * 3-7 * *"))
	a.Error(c.NewExpr("test", nil, "* * * 3-7a * *"))
}
