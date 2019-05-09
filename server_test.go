// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package scheduled

import (
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestServer_Serve(t *testing.T) {
	a := assert.New(t)
	srv := NewServer()
	a.NotNil(srv)
	a.Empty(srv.jobs).
		Equal(srv.Serve(nil), ErrNoJobs)

	srv.NewTicker("tick1", succFunc, 1*time.Second)
	srv.NewTicker("tick2", erroFunc, 2*time.Second)
	go srv.Serve(nil)
	time.Sleep(3 * time.Second)
	srv.Stop()
}
