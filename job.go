// SPDX-License-Identifier: MIT

package scheduled

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/issue9/scheduled/schedulers/at"
	"github.com/issue9/scheduled/schedulers/cron"
	"github.com/issue9/scheduled/schedulers/ticker"
)

// JobFunc 每一个定时任务实际上执行的函数签名
type JobFunc = func(time.Time) error

// Job 定时任务
type Job struct {
	s     Scheduler
	name  string
	f     JobFunc
	delay bool

	// 以下内容需要上锁

	locker sync.RWMutex
	state  State
	err    error     // 出错时的错误内容
	prev   time.Time // 上次实际上执行的时间
	next   time.Time // 下一次可能执行的时间
}

// Name 任务的名称
func (j *Job) Name() string { return j.name }

// Next 返回下次执行的时间点
//
// 如果返回值的 IsZero() 为 true，则表示该任务不需要再执行。
func (j *Job) Next() time.Time {
	j.locker.RLock()
	defer j.locker.RUnlock()
	return j.next
}

// Prev 当前正在执行或是上次执行的时间点
func (j *Job) Prev() time.Time {
	j.locker.RLock()
	defer j.locker.RUnlock()
	return j.prev
}

// State 获取当前的状态
func (j *Job) State() State {
	j.locker.RLock()
	defer j.locker.RUnlock()
	return j.state
}

// Err 返回当前的错误信息
func (j *Job) Err() error {
	j.locker.RLock()
	defer j.locker.RUnlock()
	return j.err
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
	j.next = j.s.Next(now) // 先计算 next，保证调用者重复调用 run 时能获取正确的 next。
}

// 运行当前的任务
func (j *Job) run(at time.Time, errlog, infolog Logger) {
	infolog.Printf("scheduled: start job %s at %s\n", j.Name(), at.String())

	j.locker.Lock()
	defer j.locker.Unlock()

	defer func() {
		if msg := recover(); msg != nil {
			if err, ok := msg.(error); ok {
				j.err = err
			} else {
				j.err = fmt.Errorf("%v", msg)
			}
			j.state = Failed
			errlog.Print(j.err)
		}
	}()

	if j.err = j.f(at); j.err != nil {
		j.state = Failed
		errlog.Print(j.err)
	} else {
		j.state = Stopped
	}

	j.next = j.s.Next(time.Now()) // j.f 可能会花费大量时间，所以重新计算 next
}

// 初始化当前任务，获取其下次执行时间。
func (j *Job) init(now time.Time) {
	j.locker.Lock()
	defer j.locker.Unlock()
	j.next = j.s.Next(now)
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
//
// 返回的是当前状态下的副本，具有时效性。
func (s *Server) Jobs() []*Job {
	// NOTE: jobs 有顺序要求，如果直接返回给用户，
	// 用户可能会对数据进行排序，造成无法使用，所以返回副本。

	// TODO(go1.21): 可以采用 slices.Clone
	jobs := make([]*Job, 0, len(s.jobs))
	jobs = append(jobs, s.jobs...)

	sort.SliceStable(jobs, func(i, j int) bool {
		return jobs[i].name < jobs[j].name
	})

	return jobs
}

// Tick 添加一个新的定时任务
func (s *Server) Tick(name string, f JobFunc, dur time.Duration, imm, delay bool) {
	s.New(name, f, ticker.Tick(dur, imm), delay)
}

// Cron 使用 cron 表达式新建一个定时任务
//
// 具体文件可以参考 [cron.Parse]
func (s *Server) Cron(name string, f JobFunc, spec string, delay bool) {
	scheduler, err := cron.Parse(spec, s.Location())
	if err != nil {
		panic(err)
	}
	s.New(name, f, scheduler, delay)
}

// At 添加 At 类型的定时器
//
// 具体文件可以参考 [at.At]
func (s *Server) At(name string, f JobFunc, t time.Time, delay bool) {
	s.New(name, f, at.At(t), delay)
}

// New 添加一个新的定时任务
//
// name 作为定时任务的一个简短描述，不作唯一要求；
// delay 是否从任务执行完之后，才开始计算下个执行的时间点。
func (s *Server) New(name string, f JobFunc, scheduler Scheduler, delay bool) {
	job := &Job{
		s:     scheduler,
		name:  name,
		f:     f,
		delay: delay,
	}
	s.jobs = append(s.jobs, job)

	if s.running { // 服务已经运行，则需要触发调度任务。
		job.init(time.Now())
		s.sendNextScheduled()
		s.exitSchedule <- struct{}{}
	}
}
