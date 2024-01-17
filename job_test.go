// SPDX-License-Identifier: MIT

package scheduled

import (
	"errors"
	"testing"
	"time"

	"github.com/issue9/assert/v3"

	"github.com/issue9/scheduled/schedulers"
	"github.com/issue9/scheduled/schedulers/ticker"
)

var (
	succFunc = func(n time.Time) error {
		println("succ", n.String())
		return nil
	}

	erroFunc = func(n time.Time) error {
		println("erro", n.String())
		return errors.New("erro")
	}

	failFunc = func(n time.Time) error {
		println("fail", n.String())
		panic("fail")
	}

	// 延时两秒执行
	delayFunc = func(n time.Time) error {
		println("delay", n.String())
		time.Sleep(2 * time.Second)
		return nil
	}

	errlog = &defaultLogger{}
)

func TestJob_run(t *testing.T) {
	a := assert.New(t, false)

	newTickerJob := func(duration time.Duration, imm bool) schedulers.Scheduler {
		s := ticker.Tick(duration, imm)
		a.NotNil(s)
		return s
	}

	now := time.Now()
	j := &Job{
		id: "succ",
		f:  succFunc,
		s:  newTickerJob(time.Second, false),
	}
	j.init(now)
	j.run(now, &defaultLogger{}, &defaultLogger{})
	a.Nil(j.Err()).
		Equal(j.State(), Stopped).
		Equal(j.Next().Unix(), now.Add(1*time.Second).Unix())

	now = time.Now()
	j = &Job{
		id: "erro",
		f:  erroFunc,
		s:  newTickerJob(time.Second, false),
	}
	j.init(now)
	j.run(now, errlog, &defaultLogger{})
	a.NotNil(j.Err()).
		Equal(j.State(), Failed).
		Equal(j.Next().Unix(), now.Add(1*time.Second).Unix())

	now = time.Now()
	j = &Job{
		id: "fail",
		f:  failFunc,
		s:  newTickerJob(time.Second, false),
	}
	j.init(now)
	j.run(now, errlog, errlog)
	a.NotNil(j.Err()).
		Equal(j.State(), Failed).
		Equal(j.Next().Unix(), now.Add(1*time.Second).Unix())

	// delay == true
	now = time.Now()
	j = &Job{
		id:    "delay=true",
		f:     delayFunc,
		s:     newTickerJob(time.Second, false),
		delay: true,
	}
	j.init(now)
	j.run(now, &defaultLogger{}, &defaultLogger{})
	a.Nil(j.Err()).
		Equal(j.State(), Stopped).
		True(j.Delay()).
		Equal(j.Next().Unix(), now.Add(3*time.Second).Unix()) // delayFunc 延时两秒

	// delay == false
	now = time.Now()
	j = &Job{
		id:    "delay=false",
		f:     delayFunc,
		s:     newTickerJob(time.Second, false),
		delay: false,
	}
	j.init(now)
	j.run(now, &defaultLogger{}, &defaultLogger{})
	a.Nil(j.Err()).
		Equal(j.State(), Stopped).
		Empty(j.Prev()).
		False(j.Delay()).
		Equal(j.Next().Unix(), now.Add(3*time.Second).Unix())
}

func TestSortJobs(t *testing.T) {
	a := assert.New(t, false)

	now := time.Now()
	jobs := []*Job{
		{
			id:   "1",
			next: now.Add(1111),
		},
		{
			id:   "2",
			next: time.Time{}, // zero 放在最后
		},
		{
			id:   "3",
			next: now,
		},
		{
			id:   "4",
			next: time.Time{}, // zero 放在最后
		},
		{
			id:   "5",
			next: now.Add(222),
		},
	}

	sortJobs(jobs)
	a.Equal(jobs[0].ID(), "3").
		Equal(jobs[1].ID(), "5").
		Equal(jobs[2].ID(), "1")
}

func TestServer_Jobs(t *testing.T) {
	a := assert.New(t, false)
	srv := NewServer(nil, nil, nil)
	a.NotNil(srv)

	now := time.Now()
	srv.At("j1", succFunc, now, false)
	srv.At("j3", succFunc, now, false)
	srv.At("j2", succFunc, now, false)

	jobs := srv.Jobs()
	a.Equal(len(jobs), len(srv.jobs))
}

func TestServer_Cron(t *testing.T) {
	a := assert.New(t, false)

	srv := NewServer(nil, nil, nil)
	srv.Cron("test", nil, "* * * 3-7 * *", false)
	a.Panic(func() {
		srv.Cron("test", nil, "* * * 3-7a * *", false)
	})
}
