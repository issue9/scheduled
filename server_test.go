// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package scheduled

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func TestServer_Serve(t *testing.T) {
	a := assert.New(t, false)
	term := NewTermLogger(message.NewPrinter(language.SimplifiedChinese))
	srv := NewServer(nil, term, term)
	a.NotNil(srv)

	tickers1 := make([]time.Time, 0, 20)
	tickers2 := make([]time.Time, 0, 20)
	tickers3 := make([]time.Time, 0, 20)

	srv.Tick(localeutil.StringPhrase("ticker1-delay"), func(t time.Time) error {
		tickers1 = append(tickers1, t)
		return nil
	}, time.Second, false, true)

	srv.Tick(localeutil.StringPhrase("ticker2-delay"), func(t time.Time) error {
		tickers2 = append(tickers2, t)
		return nil
	}, 2*time.Second, false, true)

	srv.Tick(localeutil.StringPhrase("ticker3-imm-delay"), func(t time.Time) error {
		tickers3 = append(tickers3, t)
		return nil
	}, 2*time.Second, true, true)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		a.ErrorIs(srv.Serve(ctx), context.Canceled)
	}()
	time.Sleep(5 * time.Second)
	cancel()

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
	a := assert.New(t, false)
	srv := NewServer(nil, nil, nil)
	a.NotNil(srv)
	a.Empty(srv.jobs)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		srv.Serve(ctx)
	}()
	time.Sleep(500 * time.Millisecond) // 等待 srv.Serve
	a.True(srv.running)

	tickers1 := make([]time.Time, 0, 20)

	srv.Tick(localeutil.StringPhrase("empty-ticker1"), func(t time.Time) error {
		tickers1 = append(tickers1, t)
		println("empty-ticker1", t.String())
		return nil
	}, time.Second, true, false)

	time.Sleep(5 * time.Second)
	cancel()

	a.NotEmpty(tickers1)
	for i := 1; i < len(tickers1); i++ {
		prev := tickers1[i-1].Unix()
		curr := tickers1[i].Unix()
		a.Equal(prev+1, curr, "%v != %v", prev, curr)
	}
}

type zero struct{}

func (z zero) Next(time.Time) time.Time { return time.Time{} }

// 附带一个 next 永远为 0 的任务
func TestServer_Serve_zero(t *testing.T) {
	a := assert.New(t, false)
	srv := NewServer(nil, nil, nil)
	a.NotNil(srv)

	// zero 应该永远不会被执行。
	tickers1 := make([]time.Time, 0, 20)
	srv.New(localeutil.StringPhrase("zero-ticker1"), func(t time.Time) error {
		tickers1 = append(tickers1, t)
		println("zero-ticker1", t.String())
		return nil
	}, zero{}, false)
	a.Equal(len(srv.Jobs()), 1)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		srv.Serve(ctx)
	}()
	time.Sleep(500 * time.Millisecond) // 等待 srv.Serve
	a.True(srv.running)

	tickers2 := make([]time.Time, 0, 20)
	srv.Tick(localeutil.StringPhrase("zero-ticker2"), func(t time.Time) error {
		tickers2 = append(tickers2, t)
		println("zero-ticker2", t.String())
		return nil
	}, time.Second, true, false)

	time.Sleep(5 * time.Second)
	cancel()

	a.Empty(tickers1)
	a.NotEmpty(tickers2)
	for i := 1; i < len(tickers2); i++ {
		prev := tickers2[i-1].Unix()
		curr := tickers2[i].Unix()
		a.Equal(prev+1, curr, "%v != %v", prev, curr)
	}
}

// 运行时间超过一个时间间隔的任务
func TestServer_Serve_delay(t *testing.T) {
	a := assert.New(t, false)
	srv := NewServer(nil, nil, nil)
	a.NotNil(srv)

	tickers1 := make([]time.Time, 0, 20)
	srv.Tick(localeutil.StringPhrase("delay-ticker1"), func(t time.Time) error {
		tickers1 = append(tickers1, t)
		println("delay-ticker1", t.String())
		time.Sleep(2 * time.Second)
		return nil
	}, time.Second, true, true)

	tickers2 := make([]time.Time, 0, 20)
	srv.Tick(localeutil.StringPhrase("delay-ticker2"), func(t time.Time) error {
		tickers2 = append(tickers2, t)
		println("delay-ticker2", t.String())
		time.Sleep(2 * time.Second)
		return nil
	}, time.Second, false, true)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		srv.Serve(ctx)
	}()
	time.Sleep(500 * time.Millisecond) // 等待 srv.Serve
	a.True(srv.running)

	time.Sleep(5 * time.Second)
	cancel()

	a.NotEmpty(tickers1).
		NotEmpty(tickers2)
	for i := 1; i < len(tickers1); i++ {
		prev := tickers1[i-1].Unix()
		curr := tickers1[i].Unix()
		delta := math.Abs(float64(curr - prev)) // 缺失一次执行，应该介于 4-6 之间？
		a.True(delta >= 4 && delta < 6, "v1=%d, v2=%d", prev, curr)
	}
	for i := 1; i < len(tickers2); i++ {
		prev := tickers2[i-1].Unix()
		curr := tickers2[i].Unix()
		delta := math.Abs(float64(curr - prev)) // 缺失一次执行，应该介于 4-6 之间？
		a.True(delta >= 4 && delta < 6, "v1=%d, v2=%d", prev, curr)
	}
}

func TestServer_Serve_loc(t *testing.T) {
	a := assert.New(t, false)

	// 将 srv 的时区调到 15 小时前，保证 job 还没到时间
	loc := time.FixedZone("UTC-15", -15*60*60)
	srv := NewServer(loc, errlog, nil)
	a.NotNil(srv)

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

	srv.Cron(localeutil.StringPhrase("cron"), job, spec, false)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		srv.Serve(ctx)
	}()
	time.Sleep(4 * time.Second) // 等待 4 秒
	cancel()
	a.Equal(0, buf.Len(), buf.String())
}
