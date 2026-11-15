// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package scheduled

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/issue9/assert/v6"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/scheduled/schedulers"
)

func TestServer_Serve(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
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
		}, time.Second, false)

		srv.Tick(localeutil.StringPhrase("ticker2-delay"), func(t time.Time) error {
			tickers2 = append(tickers2, t)
			return nil
		}, 2*time.Second, false)

		srv.Tick(localeutil.StringPhrase("ticker3-imm-delay"), func(t time.Time) error {
			tickers3 = append(tickers3, t)
			return nil
		}, 2*time.Second, true)

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() {
			defer close(done)
			a.ErrorIs(srv.Serve(ctx), context.Canceled)
		}()
		synctest.Wait()

		synctest.Sleep(5 * time.Second)                                   // 等待任务运行
		a.Equal(srv.Serve(context.Background()), errServerState(Running)) // 多次调用，返回的错误。
		a.Equal(srv.Serve(context.Background()), errServerState(Running)) // 多次调用，返回的错误。
		cancel()
		<-done

		synctest.Sleep(5 * time.Second)
		a.Equal(srv.Serve(context.Background()), errServerState(Cancelled)) // 多次调用，返回的错误。
		a.Equal(srv.Serve(context.Background()), errServerState(Cancelled)) // 多次调用，返回的错误。
		a.PanicValue(func() {
			srv.New(localeutil.Phrase("test"), succFunc, schedulers.At(time.Now()), false)
		}, errServerState(Cancelled))

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
		for i := 2; i < len(tickers3); i++ {
			prev := tickers3[i-1].Unix()
			curr := tickers3[i].Unix()
			a.Equal(prev+2, curr, "%v != %v", prev, curr)
		}
	})
}

// 初始为空，运行 Serve 之后动态添加任务
func TestServer_Serve_empty(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a := assert.New(t, false)
		srv := NewServer(nil, nil, nil)
		a.NotNil(srv)
		a.Empty(srv.jobs)

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() {
			defer close(done)
			srv.Serve(ctx)
		}()
		synctest.Wait()

		a.Equal(srv.State(), Running)

		tickers1 := make([]time.Time, 0, 20)

		srv.Tick(localeutil.StringPhrase("empty-ticker1"), func(t time.Time) error {
			tickers1 = append(tickers1, t)
			println("empty-ticker1", t.String())
			return nil
		}, 2*time.Second, true)

		synctest.Sleep(5 * time.Second)
		cancel()
		<-done

		a.NotEmpty(tickers1)
		for i := 1; i < len(tickers1); i++ {
			prev := tickers1[i-1].Unix()
			curr := tickers1[i].Unix()
			a.Equal(prev+2, curr, "%v != %v at %d", prev, curr, i)
		}
	})
}

type zero struct{}

func (z zero) Next(time.Time) time.Time { return time.Time{} }

// 附带一个 next 永远为 0 的任务
func TestServer_Serve_zero(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
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
		a.Length(srv.Jobs(), 1)

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() {
			defer close(done)
			srv.Serve(ctx)
		}()
		synctest.Wait()

		a.Equal(srv.State(), Running)

		tickers2 := make([]time.Time, 0, 20)
		srv.Tick(localeutil.StringPhrase("zero-ticker2"), func(t time.Time) error {
			tickers2 = append(tickers2, t)
			println("zero-ticker2", t.String())
			return nil
		}, time.Second*2, true)

		synctest.Sleep(5 * time.Second)
		cancel()
		<-done

		a.Empty(tickers1)
		a.NotEmpty(tickers2)
		for i := 1; i < len(tickers2); i++ {
			prev := tickers2[i-1].Unix()
			curr := tickers2[i].Unix()
			a.Equal(prev+2, curr, "%v != %v at %d", prev, curr, i)
		}
	})
}

// 运行时间超过一个时间间隔的任务
func TestServer_Serve_delay(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a := assert.New(t, false)
		srv := NewServer(nil, nil, nil)
		a.NotNil(srv)

		tickers1 := make([]time.Time, 0, 20)
		srv.Tick(localeutil.StringPhrase("delay-ticker1"), func(t time.Time) error {
			tickers1 = append(tickers1, t)
			println("delay-ticker1", t.String())
			synctest.Sleep(2 * time.Second)
			return nil
		}, time.Second, true)

		tickers2 := make([]time.Time, 0, 20)
		srv.Tick(localeutil.StringPhrase("delay-ticker2"), func(t time.Time) error {
			tickers2 = append(tickers2, t)
			println("delay-ticker2", t.String())
			synctest.Sleep(2 * time.Second)
			return nil
		}, time.Second, false)

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() {
			// Serve 是个阻塞 goroutine，需要等其完成。
			// 此处因为在 Tick 中添加了 time.Sleep，效果尤为突出。
			defer close(done)
			srv.Serve(ctx)
		}()
		synctest.Wait()
		a.Equal(srv.State(), Running)

		synctest.Sleep(5 * time.Second)
		cancel()
		<-done

		a.NotEmpty(tickers1).
			NotEmpty(tickers2)
		for i := 2; i < len(tickers1); i++ {
			prev := tickers1[i-1].Unix()
			curr := tickers1[i].Unix()
			a.InDelta(prev, curr, 4, "v1=%d, v2=%d", prev, curr)
		}
		for i := 1; i < len(tickers2); i++ {
			prev := tickers2[i-1].Unix()
			curr := tickers2[i].Unix()
			a.InDelta(prev, curr, 4, "v1=%d, v2=%d", prev, curr)
		}
	})
}

func TestServer_Serve_loc(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a := assert.New(t, false)

		// 将 srv 的时区调到 15 小时前，保证 job 还没到时间
		loc := time.FixedZone("UTC-15", -15*60*60)
		srv := NewServer(loc, &defaultLogger{}, nil)
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

		srv.Cron(localeutil.StringPhrase("cron"), job, spec)

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() {
			defer close(done)
			srv.Serve(ctx)
		}()
		synctest.Wait()
		cancel()
		<-done

		a.Equal(0, buf.Len(), buf.String())
	})
}

// imm 由 Job 层实现：首次在服务启动时立即执行，之后的间隔仍然由 scheduler 决定。
func TestServer_Tick_imm(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a := assert.New(t, false)
		srv := NewServer(nil, nil, nil)
		a.NotNil(srv)

		var got []time.Time
		srv.Tick(localeutil.StringPhrase("imm"), func(n time.Time) error {
			got = append(got, n)
			return nil
		}, 10*time.Second, true)

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() {
			defer close(done)
			srv.Serve(ctx)
		}()

		synctest.Sleep(25 * time.Second)
		cancel()
		<-done

		// 0s、10s、20s 各一次
		a.Length(got, 3)
		for i := 1; i < len(got); i++ {
			a.Equal(got[i].Sub(got[i-1]), 10*time.Second)
		}
	})
}
