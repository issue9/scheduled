// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package scheduled

import (
	"errors"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/scheduled/schedulers/at"
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
	now := time.Now()

	j := &Job{
		name:      "succ",
		f:         succFunc,
		Scheduler: ticker.New(time.Second, false),
		at:        now,
	}
	j.init(now)
	j.run(nil, nil)
	a.Nil(j.Err()).
		Equal(j.State(), Stopped).
		Equal(j.Next().Unix(), now.Add(1*time.Second).Unix())

	j = &Job{
		name:      "erro",
		f:         erroFunc,
		Scheduler: ticker.New(time.Second, false),
		at:        now,
	}
	j.init(now)
	j.run(errlog, nil)
	a.NotNil(j.Err()).
		Equal(j.State(), Failed).
		Equal(j.Next().Unix(), now.Add(1*time.Second).Unix())

	j = &Job{
		name:      "fail",
		f:         failFunc,
		Scheduler: ticker.New(time.Second, false),
		at:        now,
	}
	j.init(now)
	j.run(nil, nil)
	a.NotNil(j.Err()).
		Equal(j.State(), Failed).
		Equal(j.Next().Unix(), now.Add(1*time.Second).Unix())

	// delay == true
	j = &Job{
		name:      "delay=true",
		f:         delayFunc,
		Scheduler: ticker.New(time.Second, false),
		delay:     true,
		at:        now,
	}
	j.init(now)
	j.run(nil, nil)
	a.Nil(j.Err()).
		Equal(j.State(), Stopped).
		Equal(j.Next().Unix(), now.Add(3*time.Second).Unix()) // delayFunc 延时两秒

	// delay == false
	j = &Job{
		name:      "delay=false",
		f:         delayFunc,
		Scheduler: ticker.New(time.Second, false),
		delay:     false,
		at:        now,
	}
	j.init(now)
	j.run(nil, nil)
	a.Nil(j.Err()).
		Equal(j.State(), Stopped).
		Equal(j.Next().Unix(), now.Add(1*time.Second).Unix())
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
		{
			name:  "6",
			next:  now,
			state: Running, // Running 状态，放在最后
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

	now := time.Now().Format(at.Layout)
	a.NotError(srv.At("j1", succFunc, now, false))
	a.NotError(srv.At("j3", succFunc, now, false))
	a.NotError(srv.At("j2", succFunc, now, false))

	jobs := srv.Jobs()
	a.Equal(len(jobs), len(srv.jobs))
}

func TestServer_NewCron(t *testing.T) {
	a := assert.New(t)

	srv := NewServer(nil)
	a.NotError(srv.Cron("test", nil, "* * * 3-7 * *", false))
	a.Error(srv.Cron("test", nil, "* * * 3-7a * *", false))
}
