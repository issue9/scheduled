// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package scheduled

import (
	"context"
	"iter"
	"slices"
	"sync"
	"time"

	"github.com/issue9/localeutil"

	"github.com/issue9/scheduled/schedulers/at"
	"github.com/issue9/scheduled/schedulers/every"
	"github.com/issue9/scheduled/schedulers/ticker"
)

// JobFunc 每一个定时任务实际上执行的函数签名
type JobFunc = func(time.Time) error

// Job 定时任务
type Job struct {
	srv       *Server
	scheduler Scheduler
	title     localeutil.Stringer
	f         JobFunc
	delay     bool

	// 以下内容需要上锁

	locker sync.RWMutex
	state  State
	err    error     // 出错时的错误内容
	prev   time.Time // 上次实际上执行的时间
	next   time.Time // 下一次可能执行的时间
}

// Title 简要描述信息
func (j *Job) Title() localeutil.Stringer { return j.title }

// Next 返回下次执行的时间点
//
// 如果返回值的 IsZero() 为 true，则表示该任务不需要再执行。
func (j *Job) Next() (t time.Time) {
	j.locker.RLock()
	t = j.next
	j.locker.RUnlock()
	return t
}

// Prev 当前正在执行或是上次执行的时间点
func (j *Job) Prev() (t time.Time) {
	j.locker.RLock()
	t = j.prev
	j.locker.RUnlock()
	return t
}

// State 获取当前的状态
func (j *Job) State() (s State) {
	j.locker.RLock()
	s = j.state
	j.locker.RUnlock()
	return
}

// Err 返回当前的错误信息
func (j *Job) Err() (err error) {
	j.locker.RLock()
	err = j.err
	j.locker.RUnlock()
	return err
}

// Delay 是否在延迟执行
//
// 即从任务执行完成的时间点计算下一次执行时间。
func (j *Job) Delay() bool { return j.delay }

func (j *Job) calcState(now time.Time) {
	j.locker.Lock()
	defer j.locker.Unlock()

	j.state = Running
	j.prev = j.next
	j.next = j.scheduler.Next(now) // 先计算 next，保证调用者重复调用 run 时能获取正确的 next。
}

// 运行当前的任务
func (j *Job) run(at time.Time) {
	go j.srv.info.LocaleString(localeutil.Phrase("scheduled: start job %s at %s", j.Title(), at.String()))

	j.locker.Lock()
	defer j.locker.Unlock()

	defer func() {
		if msg := recover(); msg != nil {
			if err, ok := msg.(error); ok {
				j.err = err
			} else {
				j.err = localeutil.Error("recover msg %v", msg)
			}
			j.state = Failed
			j.srv.erro.Error(j.err) // 不像 info 一样采用异步，可能会在执行过程中 j.err 被改值。
		}
	}()

	if j.err = j.f(at); j.err != nil {
		j.state = Failed
		j.srv.erro.Error(j.err)
	} else {
		j.state = Stopped
	}

	j.next = j.scheduler.Next(time.Now().In(j.srv.Location())) // j.f 可能会花费大量时间，所以重新计算 next
}

// 初始化当前任务，获取其下次执行时间。
func (j *Job) init(now time.Time) {
	j.locker.Lock()
	defer j.locker.Unlock()
	j.next = j.scheduler.Next(now)
}

func sortJobs(jobs []*Job) {
	slices.SortFunc(jobs, func(i, j *Job) int {
		zi, zj := i.next.IsZero(), j.next.IsZero()
		switch {
		case zi && zj:
			return 0
		case zi:
			return 1 // zero 放最后
		case zj:
			return -1
		}
		return i.next.Compare(j.next)
	})
}

// Jobs 返回所有注册的任务
func (s *Server) Jobs() iter.Seq[*Job] {
	return func(yield func(*Job) bool) {
		for _, job := range s.jobs {
			if !yield(job) {
				break
			}
		}
	}
}

// Tick 添加一个新的定时任务
//
// imm 表示是否先执行一次任务，如果为 true，将会排在任务队列的前列，而不是立即执行！
func (s *Server) Tick(title localeutil.Stringer, f JobFunc, dur time.Duration, imm, delay bool) context.CancelFunc {
	return s.New(title, f, ticker.Tick(dur, imm), delay)
}

// Cron 使用 cron 表达式新建一个定时任务
//
// 具体文件可以参考 [every.ParseCron]
func (s *Server) Cron(title localeutil.Stringer, f JobFunc, spec string, delay bool) context.CancelFunc {
	return s.New(title, f, every.FromCron(spec), delay)
}

// At 添加 At 类型的定时器
//
// 具体文件可以参考 [at.At]
func (s *Server) At(title localeutil.Stringer, f JobFunc, t time.Time, delay bool) context.CancelFunc {
	return s.New(title, f, at.At(t), delay)
}

// New 添加一个新的定时任务
//
// title 任务的简要描述；
// delay 是否从任务执行完之后，才开始计算下个执行的时间点。
// 返回一个取消当前任务的方法，同时会从列表中删除。
func (s *Server) New(title localeutil.Stringer, f JobFunc, scheduler Scheduler, delay bool) context.CancelFunc {
	job := &Job{
		srv:       s,
		scheduler: scheduler,
		title:     title,
		f:         f,
		delay:     delay,
	}
	s.jobs = append(s.jobs, job)

	if s.running { // 服务已经运行，则需要触发调度任务。
		job.init(time.Now().In(s.Location()))
		s.schedule <- struct{}{} // 执行一次调度任务
	}

	return func() {
		s.jobs = slices.DeleteFunc(s.jobs, func(e *Job) bool { return e == job })
	}
}
