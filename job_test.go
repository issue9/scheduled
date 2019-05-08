// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package scheduled

import (
	"testing"

	"github.com/issue9/assert"
)

func TestServer_NewExpr(t *testing.T) {
	a := assert.New(t)

	srv := NewServer()
	a.NotError(srv.NewCron("test", nil, "* * * 3-7 * *"))
	a.Error(srv.NewCron("test", nil, "* * * 3-7a * *"))
}
