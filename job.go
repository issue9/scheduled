// SPDX-License-Identifier: MIT

package scheduled

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/issue9/scheduled/schedulers/at"
	"github.com/issue9/scheduled/schedulers/cron"
	"github.com/issue9/scheduled/schedulers/ticker"
)

// JobFunc 每一个定时任务实际上执行的函数签名
type JobFunc func(time.Time) error

// Job 一个定时任务的基本接口
type Job struct {
	s Scheduler

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

// Delay 是否在延迟执行
//
// 即从任务执行完成的时间点计算下一次执行时间。
func (j *Job) Delay() bool { return j.delay }

// 运行当前的任务
//
// errlog 在出错时，日志的输出通道，可以为空，表示不输出。
func (j *Job) run(errlog, infolog *log.Logger) {
	defer func() {
		if msg := recover(); msg != nil {
			if err, ok := msg.(error); ok {
				j.err = err
			} else {
				j.err = fmt.Errorf("%v", msg)
			}

			j.state = Failed

			if errlog != nil && j.err != nil {
				errlog.Println(j.err)
			}
		}
	}()

	// 第一条执行语句，保证最快地初始化状态为 Running
	j.state = Running

	if infolog != nil {
		infolog.Printf("scheduled: start job %s at %s\n", j.Name(), j.at.String())
	}

	j.err = j.f(j.at)
	if j.err != nil {
		j.state = Failed
	} else {
		j.state = Stopped
	}

	j.prev = j.next
	j.next = j.s.Next(time.Now())
}

// 初始化当前任务，获取其下次执行时间。
func (j *Job) init(now time.Time) { j.next = j.s.Next(now) }

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

// Tick 添加一个新的定时任务
func (s *Server) Tick(name string, f JobFunc, dur time.Duration, imm, delay bool) {
	s.New(name, f, ticker.New(dur, imm), delay)
}

// Cron 使用 cron 表达式新建一个定时任务
//
// 具体文件可以参考 schedulers/cron.Parse
func (s *Server) Cron(name string, f JobFunc, spec string, delay bool) {
	scheduler, err := cron.Parse(spec, s.Location())
	if err != nil {
		panic(err)
	}
	s.New(name, f, scheduler, delay)
}

// At 添加 At 类型的定时器
//
// 具体文件可以参考 schedulers/at.At
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

	// 服务已经运行，则需要触发调度任务。
	if s.running {
		job.init(time.Now())
		if len(s.nextScheduled) == 0 {
			s.nextScheduled <- struct{}{}
		}
	}
}
