// SPDX-License-Identifier: MIT

package scheduled

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestServer_Serve(t *testing.T) {
	a := assert.New(t)
	srv := NewServer(nil, nil, nil)
	a.NotNil(srv)

	tickers1 := make([]time.Time, 0, 20)
	tickers2 := make([]time.Time, 0, 20)
	tickers3 := make([]time.Time, 0, 20)

	srv.Tick("ticker1", func(t time.Time) error {
		tickers1 = append(tickers1, t)
		println("ticker1", t.String())
		return nil
	}, time.Second, false, true)

	srv.Tick("ticker2", func(t time.Time) error {
		tickers2 = append(tickers2, t)
		println("ticker2", t.String())
		return nil
	}, 2*time.Second, false, true)

	srv.Tick("ticker3", func(t time.Time) error {
		tickers3 = append(tickers3, t)
		println("ticker3-imm", t.String())
		return nil
	}, 2*time.Second, true, true)

	go func() {
		a.NotError(srv.Serve())
	}()
	time.Sleep(5 * time.Second)
	srv.Stop()

	a.NotEmpty(tickers1)
	for i := 1; i < len(tickers1); i++ {
		prev := tickers1[i-1].Unix()
		curr := tickers1[i].Unix()
		a.Equal(prev+1, curr, "%v != %v", prev, curr)
	}

	a.NotEmpty(tickers2)
	for i := 1; i < len(tickers2); i++ {
		prev := tickers2[i-1].Unix()
		curr := tickers2[i].Unix()
		a.Equal(prev+2, curr, "%v != %v", prev, curr)
	}

	a.NotEmpty(tickers3)
	for i := 1; i < len(tickers3); i++ {
		prev := tickers3[i-1].Unix()
		curr := tickers3[i].Unix()
		a.Equal(prev+2, curr, "%v != %v", prev, curr)
	}
}

// 初始为空，运行 Serve 之后动态添加任务
func TestServer_Serve_empty(t *testing.T) {
	a := assert.New(t)
	srv := NewServer(nil, nil, nil)
	a.NotNil(srv)
	a.Empty(srv.jobs)

	go func() {
		a.NotError(srv.Serve())
	}()
	time.Sleep(500 * time.Millisecond) // 等待 srv.Serve
	a.True(srv.running)

	tickers1 := make([]time.Time, 0, 20)

	srv.Tick("empty-ticker1", func(t time.Time) error {
		tickers1 = append(tickers1, t)
		println("empty-ticker1", t.String())
		return nil
	}, time.Second, true, false)

	time.Sleep(5 * time.Second)
	srv.Stop()

	a.NotEmpty(tickers1)
	for i := 1; i < len(tickers1); i++ {
		prev := tickers1[i-1].Unix()
		curr := tickers1[i].Unix()
		a.Equal(prev+1, curr, "%v != %v", prev, curr)
	}
}

type zero struct{}

func (z zero) Next(time.Time) time.Time { return time.Time{} }

// 初始为空，运行 Serve 之后动态添加任务
func TestServer_Serve_zero(t *testing.T) {
	a := assert.New(t)
	srv := NewServer(nil, nil, nil)
	a.NotNil(srv)

	tickers1 := make([]time.Time, 0, 20)
	tickers2 := make([]time.Time, 0, 20)

	srv.New("zero-ticker1", func(t time.Time) error {
		println("zero-ticker1", t.String())
		return nil
	}, zero{}, false)
	a.Equal(len(srv.Jobs()), 1)

	go func() {
		a.NotError(srv.Serve())
	}()
	time.Sleep(500 * time.Millisecond) // 等待 srv.Serve
	a.True(srv.running)

	srv.Tick("zero-ticker2", func(t time.Time) error {
		tickers2 = append(tickers2, t)
		println("zero-ticker2", t.String())
		return nil
	}, time.Second, true, false)

	time.Sleep(5 * time.Second)
	srv.Stop()

	a.Empty(tickers1)
	a.NotEmpty(tickers2)
	for i := 1; i < len(tickers1); i++ {
		prev := tickers2[i-1].Unix()
		curr := tickers2[i].Unix()
		a.Equal(prev+1, curr, "%v != %v", prev, curr)
	}
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
		buf.WriteString("job run at ")
		buf.WriteString(t.String())
		buf.WriteString("\n")
		return nil
	}

	now := time.Now().Add(2 * time.Second)
	_, m, d := now.Date()
	h, minute, s := now.Clock()
	spec := fmt.Sprintf("%d %d %d %d %d *", s, minute, h, d, m)

	srv.Cron("cron", job, spec, false)
	go srv.Serve()
	time.Sleep(4 * time.Second) // 等待 4 秒
	srv.Stop()
	a.Equal(0, buf.Len(), buf.String())
}
