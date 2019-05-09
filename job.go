// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package scheduled

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/issue9/scheduled/schedulers"
	"github.com/issue9/scheduled/schedulers/at"
	"github.com/issue9/scheduled/schedulers/cron"
	"github.com/issue9/scheduled/schedulers/ticker"
)

// 表示任务状态
const (
	Stoped State = iota
	Running
	Failed
)

// State 状态值类型
type State int8

// JobFunc 每一个定时任务实际上执行的函数签名
type JobFunc func() error

// Job 一个定时任务的基本接口
type Job struct {
	name      string
	f         JobFunc
	scheduler schedulers.Scheduler
	state     State
	err       error // 出错时的错误内容

	prev, next time.Time
}

// New 添加一个新的定时任务
//
// name 作为定时任务的一个简短描述，不作唯一要求
func (s *Server) New(name string, f JobFunc, scheduler schedulers.Scheduler) error {
	if s.running {
		return ErrRunning
	}

	s.jobs = append(s.jobs, &Job{
		name:      name,
		f:         f,
		scheduler: scheduler,
	})
	return nil
}

// Name 任务的名称
func (j *Job) Name() string { return j.name }

// Next 该任务关联的 Nexter 接口
func (j *Job) Next() schedulers.Scheduler { return j.scheduler }

// State 获取当前的状态
func (j *Job) State() State { return j.state }

// Err 返回当前的错误信息
func (j *Job) Err() error { return j.err }

// 运行当前的任务
//
// errlog 在出错时，日志的输出通道，可以为空，表示不输出。
func (j *Job) run(now time.Time, errlog *log.Logger) {
	defer func() {
		if msg := recover(); msg != nil {
			if err, ok := msg.(error); ok {
				j.err = err
			} else {
				j.err = fmt.Errorf("job %s error: %v", j.name, msg)
			}

			j.state = Failed
		}

		if errlog != nil && j.err != nil {
			errlog.Println(j.err)
		}
	}()

	j.state = Running
	j.err = j.f()

	if j.err != nil {
		j.state = Failed
	} else {
		j.state = Stoped
		j.err = nil
	}

	j.prev = j.next
	j.next = j.scheduler.Next(j.next)
}

// 初始化当前任务，获取其下次执行时间。
func (j *Job) init(now time.Time) {
	j.next = j.scheduler.Next(now)
}

func sortJobs(jobs []*Job) {
	sort.SliceStable(jobs, func(i, j int) bool {
		if jobs[i].next.IsZero() {
			return false
		}
		if jobs[j].next.IsZero() {
			return true
		}
		return jobs[i].next.Before(jobs[j].next)
	})
}

// Jobs 返回所有注册的任务
func (s *Server) Jobs() []*Job {
	jobs := make([]*Job, 0, len(s.jobs))
	jobs = append(jobs, s.jobs...)

	sort.SliceStable(jobs, func(i, j int) bool {
		return jobs[i].name < jobs[j].name
	})

	return jobs
}

// NewTicker 添加一个新的定时任务
func (s *Server) NewTicker(name string, f JobFunc, dur time.Duration) error {
	return s.New(name, f, ticker.New(dur))
}

// NewCron 使用 cron 表达式新建一个定时任务
//
// 具体文件可以参考 schedulers/cron.Parse
func (s *Server) NewCron(name string, f JobFunc, spec string) error {
	scheduler, err := cron.Parse(spec)
	if err != nil {
		return err
	}

	return s.New(name, f, scheduler)
}

// NewAt 添加 At 类型的定时器
//
// 具体文件可以参考 schedulers/at.At
func (s *Server) NewAt(name string, f JobFunc, t time.Time) error {
	return s.New(name, f, at.At(t))
}
