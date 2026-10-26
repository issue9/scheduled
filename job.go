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

	// 以下内容需要上锁

	locker sync.RWMutex
	state  State
	err    error // 出错时的错误内容
	prev   time.Time
	next   time.Time
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

func (j *Job) calcState(now time.Time) {
	j.locker.Lock()
	j.state = Running
	j.prev = j.next
	// 由文档约定不能 panic，先计算 next，保证调用者重复调用 run 时能获取正确的 next。
	j.next = j.scheduler.Next(now)
	j.locker.Unlock()
}

// 运行当前的任务
//
// at 调用此函数的大致时间
func (j *Job) run(at time.Time) {
	// 可能会存在通过 Server.New 返回的函数主动取消单个 Job 的方式，所以还要考虑 j.State 的值。
	if j.State() == Cancelled {
		return
	}

	go j.srv.info.LocaleString(localeutil.Phrase("scheduled: start job %s at %s", j.Title(), at.String()))

	defer func() {
		var err error
		if msg := recover(); msg != nil {
			var ok bool
			if err, ok = msg.(error); !ok {
				err = localeutil.Error("recover msg %v", msg)
			}
			go j.srv.erro.Error(err)
		}

		j.locker.Lock()
		if j.state != Cancelled {
			if err != nil {
				j.state = Failed
				j.err = err
			}
			j.next = j.scheduler.Next(time.Now().In(j.srv.Location())) // 由文档约定不能 panic
		}
		j.locker.Unlock()
	}()

	if err := j.f(at); err != nil {
		j.srv.erro.Error(err)

		if j.State() == Cancelled { // Cancelled 不需要改状态
			j.locker.Lock()
			j.next = time.Time{}
			j.err = err
			j.locker.Unlock()
			return
		}

		j.locker.Lock()
		j.state = Failed
		j.err = err
		j.locker.Unlock()
		return
	}

	// Cancelled 不用改状态为 Stopped，也不用生成 j.next 数据。
	if j.State() == Cancelled {
		return
	}

	j.locker.Lock()
	j.state = Stopped
	j.locker.Unlock()
}

// 初始化当前任务，获取其下次执行时间。
func (j *Job) init(now time.Time) {
	j.locker.Lock()
	j.next = j.scheduler.Next(now)
	j.locker.Unlock()
}

func sortJobs(jobs []*Job) {
	slices.SortFunc(jobs, func(i, j *Job) int {
		ni, nj := i.Next(), j.Next()

		zi, zj := ni.IsZero(), nj.IsZero()
		switch {
		case zi && zj:
			return 0
		case zi:
			return 1 // zero 放最后
		case zj:
			return -1
		}

		return ni.Compare(nj)
	})
}

// Jobs 返回所有注册的任务
func (s *Server) Jobs() iter.Seq[*Job] {
	s.jobsMu.RLock()
	jobs := slices.Clone(s.jobs)
	s.jobsMu.RUnlock()

	return func(yield func(*Job) bool) {
		for _, job := range jobs {
			if !yield(job) {
				break
			}
		}
	}
}

// Tick 添加一个新的定时任务
//
// imm 表示是否先执行一次任务，如果为 true，将会排在任务队列的前列，而不是立即执行！
func (s *Server) Tick(title localeutil.Stringer, f JobFunc, dur time.Duration, imm bool) context.CancelFunc {
	return s.New(title, f, ticker.Tick(dur, imm))
}

// Cron 使用 cron 表达式新建一个定时任务
//
// 具体文件可以参考 [every.ParseCron]
func (s *Server) Cron(title localeutil.Stringer, f JobFunc, spec string) context.CancelFunc {
	return s.New(title, f, every.FromCron(spec))
}

// At 添加 At 类型的定时器
//
// 具体文件可以参考 [at.At]
func (s *Server) At(title localeutil.Stringer, f JobFunc, t time.Time) context.CancelFunc {
	return s.New(title, f, at.At(t))
}

// New 添加一个新的定时任务
//
// title 任务的简要描述；
// f 定时任务实际执行的函数；
// scheduler 定时任务的调度器；
// 返回一个取消当前任务的方法，同时会从列表中删除。
func (s *Server) New(title localeutil.Stringer, f JobFunc, scheduler Scheduler) context.CancelFunc {
	if scheduler == nil {
		panic("参数 scheduler 不能为空")
	}
	if f == nil {
		panic("参数 f 不能为空")
	}
	if s.cancelled {
		panic(ErrServerCancelled())
	}

	job := &Job{
		srv:       s,
		scheduler: scheduler,
		title:     title,
		f:         f,
	}
	s.jobsMu.Lock()
	s.jobs = append(s.jobs, job)
	s.jobsMu.Unlock()

	if s.running { // 服务已经运行，则需要触发调度任务。
		job.init(time.Now().In(s.Location()))
		select {
		case s.schedule <- struct{}{}:
		default:
		}
	}

	return func() {
		job.locker.Lock()
		job.state = Cancelled
		job.next = time.Time{}
		job.locker.Unlock()

		s.jobsMu.Lock()
		s.jobs = slices.DeleteFunc(s.jobs, func(e *Job) bool { return e == job })
		s.jobsMu.Unlock()
	}
}
