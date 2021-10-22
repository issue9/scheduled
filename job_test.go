// SPDX-License-Identifier: MIT

package scheduled

import (
	"errors"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/issue9/assert"

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

	errlog = log.New(ioutil.Discard, "ERRO", 0)
)

func TestJob_run(t *testing.T) {
	a := assert.New(t)

	newTickerJob := func(duration time.Duration, imm bool) schedulers.Scheduler {
		s := ticker.New(duration, imm)
		a.NotNil(s)
		return s
	}

	now := time.Now()
	j := &Job{
		name: "succ",
		f:    succFunc,
		s:    newTickerJob(time.Second, false),
	}
	j.init(now)
	j.run(now, nil, nil)
	a.Nil(j.Err()).
		Equal(j.State(), Stopped).
		Equal(j.Next().Unix(), now.Add(1*time.Second).Unix())

	now = time.Now()
	j = &Job{
		name: "erro",
		f:    erroFunc,
		s:    newTickerJob(time.Second, false),
	}
	j.init(now)
	j.run(now, errlog, nil)
	a.NotNil(j.Err()).
		Equal(j.State(), Failed).
		Equal(j.Next().Unix(), now.Add(1*time.Second).Unix())

	now = time.Now()
	j = &Job{
		name: "fail",
		f:    failFunc,
		s:    newTickerJob(time.Second, false),
	}
	j.init(now)
	j.run(now, nil, nil)
	a.NotNil(j.Err()).
		Equal(j.State(), Failed).
		Equal(j.Next().Unix(), now.Add(1*time.Second).Unix())

	// delay == true
	now = time.Now()
	j = &Job{
		name:  "delay=true",
		f:     delayFunc,
		s:     newTickerJob(time.Second, false),
		delay: true,
	}
	j.init(now)
	j.run(now, nil, nil)
	a.Nil(j.Err()).
		Equal(j.State(), Stopped).
		Equal(j.Next().Unix(), now.Add(3*time.Second).Unix()) // delayFunc 延时两秒

	// delay == false
	now = time.Now()
	j = &Job{
		name:  "delay=false",
		f:     delayFunc,
		s:     newTickerJob(time.Second, false),
		delay: false,
	}
	j.init(now)
	j.run(now, nil, nil)
	a.Nil(j.Err()).
		Equal(j.State(), Stopped).
		Equal(j.Next().Unix(), now.Add(3*time.Second).Unix())
}

func TestSortJobs(t *testing.T) {
	a := assert.New(t)

	now := time.Now()
	jobs := []*Job{
		{
			name: "1",
			next: now.Add(1111),
		},
		{
			name: "2",
			next: time.Time{}, // zero 放在最后
		},
		{
			name: "3",
			next: now,
		},
		{
			name: "4",
			next: time.Time{}, // zero 放在最后
		},
		{
			name: "5",
			next: now.Add(222),
		},
	}

	sortJobs(jobs)
	a.Equal(jobs[0].name, "3").
		Equal(jobs[1].name, "5").
		Equal(jobs[2].name, "1")
}

func TestServer_Jobs(t *testing.T) {
	a := assert.New(t)
	srv := NewServer(nil)
	a.NotNil(srv)

	now := time.Now()
	srv.At("j1", succFunc, now, false)
	srv.At("j3", succFunc, now, false)
	srv.At("j2", succFunc, now, false)

	jobs := srv.Jobs()
	a.Equal(len(jobs), len(srv.jobs))
}

func TestServer_NewCron(t *testing.T) {
	a := assert.New(t)

	srv := NewServer(nil)
	srv.Cron("test", nil, "* * * 3-7 * *", false)
	a.Panic(func() {
		srv.Cron("test", nil, "* * * 3-7a * *", false)
	})
}
