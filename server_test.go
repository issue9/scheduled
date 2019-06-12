// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package scheduled

import (
	"bytes"
	"sync/atomic"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/scheduled/schedulers/at"
)

type incr struct {
	count time.Duration
}

func (i *incr) Next(t time.Time) time.Time {
	i.count += 2
	return t.Add(i.count * time.Second)
}

func (i *incr) Title() string {
	return "递增"
}

func TestServer_Serve1(t *testing.T) {
	a := assert.New(t)
	srv := NewServer(nil)
	a.NotNil(srv)

	var ticker1 int64
	var ticker2 int64

	a.NotError(srv.NewTicker("ticker1", func(t time.Time) error {
		atomic.AddInt64(&ticker1, 1)
		return nil
	}, time.Second, false))

	a.NotError(srv.New("ticker2", func(t time.Time) error {
		atomic.AddInt64(&ticker2, 1)
		return nil
	}, &incr{}, false))

	go func() {
		a.NotError(srv.Serve(nil, nil))
	}()

	<-time.NewTimer(6 * time.Second).C
	srv.Stop()
	println(ticker1, ticker2)
	a.True(ticker1 > ticker2, ticker1, ticker2)
}

func TestServer_Serve(t *testing.T) {
	a := assert.New(t)
	srv := NewServer(nil)
	a.NotNil(srv)
	a.Empty(srv.jobs).
		Equal(srv.Serve(nil, nil), ErrNoJobs)

	a.NotError(srv.NewTicker("tick1", succFunc, 1*time.Second, false))
	a.NotError(srv.NewTicker("tick2", erroFunc, 2*time.Second, false))
	a.NotError(srv.NewTicker("delay", delayFunc, 1*time.Second, false))
	go srv.Serve(nil, nil)
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
	srv.NewAt("xxx", job, now, false)
	go srv.Serve(errlog, nil)
	time.Sleep(3 * time.Second)
	a.Equal(0, buf.Len(), buf.String())
}
