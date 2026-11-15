// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package scheduled

import (
	"context"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/issue9/localeutil"

	"github.com/issue9/scheduled/schedulers"
	"github.com/issue9/scheduled/schedulers/every"
)

// JobFunc 每一个定时任务实际上执行的函数签名
type JobFunc = func(time.Time) error

// Job 定时任务的快照信息
type Job struct {
	First    time.Time           // 第一次运行时间
	Count    int                 // 运行的次数
	ErrCount int                 // 运行出错的次数
	Title    localeutil.Stringer // 简要描述信息
	Next     time.Time           // 下次执行的时间点
	Prev     time.Time           // 当前正在执行或是上次执行的时间点
	State    State               // 当前的状态
	Err      error               // 当前的错误信息
}

type job struct {
	srv       *Server
	scheduler Scheduler
	title     localeutil.Stringer
	f         JobFunc
	first     time.Time
	imm       bool // 首次调度时是否立即执行，只在 [job.init] 中使用

	// 以下内容需要上锁

	locker   sync.RWMutex
	state    State
	err      error // 出错时的错误内容
	prev     time.Time
	next     time.Time
	count    int
	errCount int
}

// 运行当前的任务
//
// at 调用此函数的大致时间
//
// NOTE: 此方法由 WaitGroup.Go 调用，不能 panic。
func (j *job) run(at time.Time) {
	j.srv.wait.Go(func() {
		j.srv.info.LocaleString(localeutil.Phrase("scheduled: start job %s at %s", j.title, at.String()))
	})

	// 所有可能的错误信息都通过此变量中转给 j.finish。
	//
	// j.f 如果返回 err 那么就不会在 recover 中捕获 panic，
	// 所以不存在 panic 会覆盖 j.f 返回值的问题。
	var err error

	defer func() { // NOTE: 此 defer 应该始终执行到最后，那里有触发任务调度的机制。
		if msg := recover(); msg != nil {
			err = any2Error(msg)
		}
		j.finish(err)

		// 一旦有一个 job 的 next 重算，需要重新调度任务。
		select {
		case j.srv.schedule <- struct{}{}:
		default:
		}
	}()

	err = j.f(time.Now().In(j.srv.Location()))
}

// finish 统一处理任务结束后的状态；err 为 nil 表示成功。
func (j *job) finish(err error) {
	if err != nil {
		j.srv.wait.Go(func() { j.srv.erro.Error(err) })
	}
	now := time.Now().In(j.srv.Location())

	j.locker.Lock()
	defer func() {
		// 虽然约定了 [Scheduler.Next] 不能 panic，但是万一呢。
		if msg := recover(); msg != nil {
			// err!=nil 的错误计数已经在之前统计过。只 msg != nil 且 err == nil 的
			// 才是由 j.scheduler.Next 中 panic 的新信息，需要计数。
			if err == nil {
				j.errCount++
			}
			j.fromRecover(msg)
		}
		j.locker.Unlock()
	}()

	j.count++
	if err != nil {
		j.errCount++
		j.err = err
	}
	if j.state == Cancelled { // 期间被取消，保持 Cancelled 且不再排期
		j.next = time.Time{}
		return
	}
	if err != nil {
		j.state = Failed
	} else {
		j.state, j.err = Stopped, nil
	}

	// NOTE: 如果 j.scheduler.Next 占用大量时间，那么会占用大量锁的时间。
	// 如果放在锁外，对于 Cancelled 状态又会浪费一次 Next 的调用。
	// 把锁拆分为两段，那么就无法保证状态是统一的。
	j.prev, j.next = j.next, j.scheduler.Next(now)

	// 时间上没有推进
	if !j.next.IsZero() && !j.next.After(now) {
		j.state = Failed
		j.err = localeutil.Error("can not get next schedule time")
		j.next = time.Time{}
		if err == nil { // 如果不是进入函数的时候就是带 err，那么还是要加上 errCount 的
			j.errCount++
		}
	}
}

// 初始化当前任务，获取其下次执行时间。
func (j *job) init(now time.Time) {
	j.locker.Lock()

	defer func() {
		// 虽然约定了 [Scheduler.Next] 不能 panic，但是万一呢。
		if msg := recover(); msg != nil {
			j.fromRecover(msg)
		}
		j.locker.Unlock()
	}()

	if j.imm { // 首次立即执行，不需要也不应该经过 scheduler
		j.next = now
	} else {
		// fromRecover 未设置 j.first，所以此语句要放前面。
		j.next = j.scheduler.Next(now)
	}
	j.first = j.next
}

// 从 recover 的返回值对 job 做状态修改
func (j *job) fromRecover(msg any) {
	j.err = any2Error(msg)
	j.next = time.Time{}
	j.state = Failed
}

func sortJobs(jobs []*job) {
	slices.SortFunc(jobs, func(i, j *job) int {
		i.locker.RLock()
		ni := i.next
		i.locker.RUnlock()

		j.locker.RLock()
		nj := j.next
		j.locker.RUnlock()

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

// Jobs 返回当前所有任务的快照
func (s *Server) Jobs() []*Job {
	s.locker.RLock()
	jobs := make([]*Job, 0, len(s.jobs))
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
	s.locker.RUnlock()
	return jobs
}

// Tick 添加一个新的定时任务
//
// imm 表示是否先执行一次任务！
func (s *Server) Tick(title localeutil.Stringer, f JobFunc, dur time.Duration, imm bool) context.CancelFunc {
	return s.New(title, f, schedulers.Tick(dur), imm)
}

// Cron 使用 cron 表达式新建一个定时任务
//
// 具体文档可以参考 [every.ParseCron]
// 相对于 [every.ParseCron]，当前方法添加了对 '@reboot' 指令的支持。
func (s *Server) Cron(title localeutil.Stringer, f JobFunc, spec string) context.CancelFunc {
	spec = strings.TrimSpace(spec)
	if spec == "@reboot" {
		return s.New(title, f, schedulers.Nop(), true)
	}
	return s.New(title, f, every.FromCron(spec), false)
}

// At 添加 [at.At] 类型的定时器
func (s *Server) At(title localeutil.Stringer, f JobFunc, t time.Time) context.CancelFunc {
	return s.New(title, f, schedulers.At(t), false)
}

// Earliest 添加 [earliest.Earliest] 类型的调度器
func (s *Server) Earliest(title localeutil.Stringer, f JobFunc, scheduler ...schedulers.Scheduler) context.CancelFunc {
	return s.New(title, f, schedulers.Earliest(scheduler...), false)
}

// New 添加一个新的定时任务
//
// title 任务的简要描述；
// f 定时任务实际执行的函数；
// scheduler 定时任务的调度器；
// imm 是否将首次执行时间设为当前时间（即启动后立即执行一次）；
// 返回一个取消当前任务的方法，同时会从列表中删除。
//
// 如果 [Server.State] 为 [Cancelled] 那么将会直接触发 panic。
func (s *Server) New(title localeutil.Stringer, f JobFunc, scheduler Scheduler, imm bool) context.CancelFunc {
	if scheduler == nil {
		panic("参数 scheduler 不能为空")
	}
	if f == nil {
		panic("参数 f 不能为空")
	}

	j := &job{
		srv:       s,
		scheduler: scheduler,
		title:     title,
		f:         f,
		imm:       imm,
	}

	s.locker.Lock()
	if s.state == Cancelled {
		s.locker.Unlock()
		panic(errServerState(Cancelled))
	}

	s.jobs = append(s.jobs, j)
	isRunning := s.state == Running
	s.locker.Unlock()

	if isRunning { // 服务已经运行，则需要触发调度任务。
		j.init(time.Now().In(s.Location()))
		select {
		case s.schedule <- struct{}{}:
		default:
		}
	}

	return func() {
		j.locker.Lock()
		if j.state == Cancelled { // 保证 cancel 函数是幂等的。
			j.locker.Unlock()
			return
		}
		j.state = Cancelled
		j.next = time.Time{}
		j.locker.Unlock()

		s.locker.Lock()
		s.jobs = slices.DeleteFunc(s.jobs, func(e *job) bool { return e == j })
		s.locker.Unlock()
		select {
		case s.schedule <- struct{}{}:
		default:
		}
	}
}
