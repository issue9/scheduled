// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package scheduled

import (
	"bytes"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/scheduled/schedulers/at"
)

func TestServer_Serve(t *testing.T) {
	a := assert.New(t)
	srv := NewServer(nil)
	a.NotNil(srv)
	a.Empty(srv.jobs).
		Equal(srv.Serve(nil), ErrNoJobs)

	srv.NewTicker("tick1", succFunc, 1*time.Second)
	srv.NewTicker("tick2", erroFunc, 2*time.Second)
	go srv.Serve(nil)
	time.Sleep(3 * time.Second)
	srv.Stop()
}

func TestServer_Serve_loc(t *testing.T) {
	a := assert.New(t)

	// 将 srv 的时区调到 15 小时前，保证 job 还没到时间
	loc := time.FixedZone("UTC-15", -15*60*60)
	srv := NewServer(loc)
	a.NotError(srv)

	buf := new(bytes.Buffer)
	a.Equal(0, buf.Len())
	job := func(t time.Time) error {
		buf.WriteString(t.String())
		buf.WriteString("\n")
		return nil
	}

	now := time.Now().Format(at.Layout)
	srv.NewAt("xxx", job, now)
	go srv.Serve(errlog)
	time.Sleep(3 * time.Second)
	a.Equal(0, buf.Len(), buf.String())
}
