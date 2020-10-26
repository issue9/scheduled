// SPDX-License-Identifier: MIT

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
	srv := NewServer(nil, nil, nil)
	a.NotNil(srv)

	var ticker1 int64
	var ticker2 int64

	a.NotError(srv.Tick("ticker1", func(t time.Time) error {
		atomic.AddInt64(&ticker1, 1)
		return nil
	}, time.Second, false, false))

	srv.New("ticker2", func(t time.Time) error {
		atomic.AddInt64(&ticker2, 1)
		return nil
	}, &incr{}, false)

	go func() {
		a.NotError(srv.Serve())
	}()

	time.Sleep(3 * time.Second)
	srv.Stop()
	a.True(ticker1 > ticker2, ticker1, ticker2)
}

func TestServer_Serve(t *testing.T) {
	a := assert.New(t)
	srv := NewServer(nil, nil, nil)
	a.NotNil(srv)
	a.Empty(srv.jobs).
		Equal(srv.Serve(), ErrNoJobs)

	a.NotError(srv.Tick("tick1", succFunc, 1*time.Second, false, false))
	a.NotError(srv.Tick("tick2", erroFunc, 2*time.Second, false, false))
	go srv.Serve()
	time.Sleep(3 * time.Second)
	a.NotError(srv.Tick("delay", delayFunc, 1*time.Second, false, false))
	time.Sleep(2 * time.Second)
	srv.Stop()
}

func TestServer_Serve_loc(t *testing.T) {
	a := assert.New(t)

	// 将 srv 的时区调到 15 小时前，保证 job 还没到时间
	loc := time.FixedZone("UTC-15", -15*60*60)
	srv := NewServer(loc, errlog, nil)
	a.NotError(srv)

	buf := new(bytes.Buffer)
	a.Equal(0, buf.Len())
	job := func(t time.Time) error {
		buf.WriteString(t.String())
		buf.WriteString("\n")
		return nil
	}

	now := time.Now().Format(at.Layout)
	srv.At("xxx", job, now, false)
	go srv.Serve()
	time.Sleep(3 * time.Second)
	a.Equal(0, buf.Len(), buf.String())
}
