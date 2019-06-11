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
type JobFunc func(time.Time) error

// Job 一个定时任务的基本接口
type Job struct {
	schedulers.Scheduler

	name  string
	f     JobFunc
	state State
	err   error // 出错时的错误内容
	delay bool

	// prev 上次实际上执行的时间
	// next 下一次可能执行的时间
	// at 是由调度器在实际调用时的时间。
	prev, next, at time.Time
}

func (s State) String() string {
	switch s {
	case Stoped:
		return "stoped"
	case Running:
		return "running"
	case Failed:
		return "failed"
	default:
		return "<unknown>"
	}
}

// Name 任务的名称
func (j *Job) Name() string { return j.name }

// Next 返回下次执行的时间点
//
// 如果返回值的 IsZero() 为 true，则表示该任务不需要再执行，
// 一般为 At 之类的一次任务。
func (j *Job) Next() time.Time { return j.next }

// Prev 当前正在执行或是上次执行的时间点
func (j *Job) Prev() time.Time { return j.prev }

// State 获取当前的状态
func (j *Job) State() State { return j.state }

// Err 返回当前的错误信息
func (j *Job) Err() error { return j.err }

// Delay 是否在延迟执行。
// 即从任务执行完成的时间点计算下一次执行时间。
func (j *Job) Delay() bool { return j.delay }

// 运行当前的任务
//
// errlog 在出错时，日志的输出通道，可以为空，表示不输出。
func (j *Job) run(errlog *log.Logger) {
	now := j.at

	defer func() {
		if msg := recover(); msg != nil {
			if err, ok := msg.(error); ok {
				j.err = err
			} else {
				j.err = fmt.Errorf("job %s error: %v", j.name, msg)
			}

			j.state = Failed

			if errlog != nil && j.err != nil {
				errlog.Println(j.err)
			}
		}
	}()

	j.err = j.f(now)
	if j.err != nil {
		j.state = Failed
	} else {
		j.state = Stoped
	}

	if j.Delay() {
		now = time.Now().In(now.Location())
	}

	j.prev = j.next
	j.next = j.Scheduler.Next(now)
}

// 初始化当前任务，获取其下次执行时间。
func (j *Job) init(now time.Time) {
	j.next = j.Scheduler.Next(now)
}

func sortJobs(jobs []*Job) {
	sort.SliceStable(jobs, func(i, j int) bool {
		if jobs[i].next.IsZero() || jobs[i].State() == Running {
			return false
		}
		if jobs[j].next.IsZero() || jobs[j].State() == Running {
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
func (s *Server) NewTicker(name string, f JobFunc, dur time.Duration, delay bool) error {
	return s.New(name, f, ticker.New(dur), delay)
}

// NewCron 使用 cron 表达式新建一个定时任务
//
// 具体文件可以参考 schedulers/cron.Parse
func (s *Server) NewCron(name string, f JobFunc, spec string, delay bool) error {
	scheduler, err := cron.Parse(spec)
	if err != nil {
		return fmt.Errorf("解析参数 spec 出错：%s" + err.Error())
	}

	return s.New(name, f, scheduler, delay)
}

// NewAt 添加 At 类型的定时器
//
// 具体文件可以参考 schedulers/at.At
func (s *Server) NewAt(name string, f JobFunc, t string, delay bool) error {
	scheduler, err := at.At(t)
	if err != nil {
		return fmt.Errorf("解析参数 t 出错：%s" + err.Error())
	}
	return s.New(name, f, scheduler, delay)
}

// New 添加一个新的定时任务
//
// name 作为定时任务的一个简短描述，不作唯一要求；
// delay 是否从任务执行完之后，才开始计算下个执行的时间点。
func (s *Server) New(name string, f JobFunc, scheduler schedulers.Scheduler, delay bool) error {
	if s.running {
		return ErrRunning
	}

	s.jobs = append(s.jobs, &Job{
		Scheduler: scheduler,
		name:      name,
		f:         f,
		delay:     delay,
	})
	return nil
}
