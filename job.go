// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package scheduled

import (
	"context"
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

// Job 定时任务的快照信息
type Job struct {
	First    time.Time           // 第一次运行时间
	Count    int                 // 运行的次数
	ErrCount int                 // 运行出错的次数
	Title    localeutil.Stringer // 简要描述信息
	Next     time.Time           // 返回下次执行的时间点
	Prev     time.Time           // 当前正在执行或是上次执行的时间点
	State    State               // 获取当前的状态
	Err      error               // 返回当前的错误信息
}

type job struct {
	srv       *Server
	scheduler Scheduler
	title     localeutil.Stringer
	f         JobFunc
	first     time.Time

	// 以下内容需要上锁

	locker   sync.RWMutex
	state    State
	err      error // 出错时的错误内容
	prev     time.Time
	next     time.Time
	count    int
	errCount int
}

func (j *job) setRunning() {
	j.locker.Lock()
	j.state = Running
	j.locker.Unlock()
}

// 运行当前的任务
//
// at 调用此函数的大致时间
func (j *job) run(at time.Time) {
	go j.srv.info.LocaleString(localeutil.Phrase("scheduled: start job %s at %s", j.title, at.String()))

	// NOTE: 此 defer 应该始终执行到最后，那里有触发任务调度的机制。
	defer func() {
		if msg := recover(); msg != nil {
			err, ok := msg.(error)
			if !ok {
				err = localeutil.Error("recover msg %v", msg)
			}
			j.calcFailedState(err)
		}

		// 一旦有一个 job 的 next 重算，需要重新调度任务。
		select {
		case j.srv.schedule <- struct{}{}:
		default:
		}
	}()

	if err := j.f(at); err != nil {
		j.calcFailedState(err)
		return
	}

	j.calcStoppedState()
}

func (j *job) calcStoppedState() {
	j.locker.Lock()
	if j.state == Cancelled {
		j.next = time.Time{}
		j.locker.Unlock()
		return
	}
	j.locker.Unlock()

	now := time.Now().In(j.srv.Location())
	next := j.scheduler.Next(now) // 由文档约定不能 panic

	j.locker.Lock()
	j.prev = j.next
	j.next = next
	j.state = Stopped
	j.count++
	j.locker.Unlock()
}

func (j *job) calcFailedState(err error) {
	go j.srv.erro.Error(err)

	j.locker.Lock()
	if j.state == Cancelled {
		j.err = err
		j.next = time.Time{}
		j.locker.Unlock()
		return
	}
	j.locker.Unlock()

	now := time.Now().In(j.srv.Location())
	next := j.scheduler.Next(now) // 由文档约定不能 panic

	j.locker.Lock()
	j.err = err
	j.errCount++
	j.prev = j.next
	j.next = next
	j.state = Failed
	j.count++
	j.locker.Unlock()
}

// 初始化当前任务，获取其下次执行时间。
func (j *job) init(now time.Time) {
	j.locker.Lock()
	j.next = j.scheduler.Next(now)
	j.first = j.next
	j.locker.Unlock()
}

func sortJobs(jobs []*job) {
	slices.SortFunc(jobs, func(i, j *job) int {
		i.locker.RLock()
		ni := i.next
		i.locker.RUnlock()

		i.locker.RLock()
		nj := j.next
		i.locker.RUnlock()

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

// Jobs 还回当前所有任务的快照
func (s *Server) Jobs() []*Job {
	jobs := make([]*Job, 0, len(s.jobs))
	s.jobsMu.RLock()
	for _, j := range s.jobs {
		j.locker.RLock()
		jobs = append(jobs, &Job{
			First:    j.first,
			Count:    j.count,
			ErrCount: j.errCount,
			Title:    j.title,
			Next:     j.next,
			Prev:     j.prev,
			State:    j.state,
			Err:      j.err,
		})
		j.locker.RUnlock()
	}
	s.jobsMu.RUnlock()
	return jobs
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

	j := &job{
		srv:       s,
		scheduler: scheduler,
		title:     title,
		f:         f,
	}
	s.jobsMu.Lock()
	s.jobs = append(s.jobs, j)
	s.jobsMu.Unlock()

	if s.running { // 服务已经运行，则需要触发调度任务。
		j.init(time.Now().In(s.Location()))
		select {
		case s.schedule <- struct{}{}:
		default:
		}
	}

	return func() {
		j.locker.Lock()
		j.state = Cancelled
		j.next = time.Time{}
		j.locker.Unlock()

		s.jobsMu.Lock()
		s.jobs = slices.DeleteFunc(s.jobs, func(e *job) bool { return e == j })
		s.jobsMu.Unlock()
	}
}
