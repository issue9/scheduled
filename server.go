// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package scheduled

import (
	"context"
	"sync"
	"time"
)

// Server 管理所有的定时任务
type Server struct {
	jobs                []*Job
	nextScheduled       chan struct{} // 需要指行下一次调度任务
	exitSchedule        chan struct{} // 没有立即了执行的任务，则退出调度任务
	scheduleLocker      sync.Mutex    // 保证 schedule 方法调用的唯一性
	nextScheduledLocker sync.Mutex    // 保证 sendNextScheduled 方法调用的唯一性

	loc        *time.Location
	running    bool
	erro, info Logger
}

// NewServer 声明 [Server]
//
// loc 指定当前所采用的时区，若为 nil，则会采用 [time.Local] 的值；
// erro 计划任务发生的错误，向此输出，可以为空，表示不输出；
// info 计划任务的执行信息，向此输出，可以为空，表示不输出；
func NewServer(loc *time.Location, erro, info Logger) *Server {
	if loc == nil {
		loc = time.Local
	}

	if erro == nil {
		erro = &defaultLogger{}
	}

	if info == nil {
		info = &defaultLogger{}
	}

	return &Server{
		jobs:          make([]*Job, 0, 100),
		nextScheduled: make(chan struct{}, 1),
		exitSchedule:  make(chan struct{}, 1),

		loc:  loc,
		erro: erro,
		info: info,
	}
}

// Location 返回当前任务相关联的时区信息
func (s *Server) Location() *time.Location { return s.loc }

// Serve 运行服务
func (s *Server) Serve(ctx context.Context) error {
	s.running = true
	defer func() {
		s.running = false
	}()

	// 初始化任务
	now := time.Now()
	for _, job := range s.jobs {
		job.init(now)
	}

	s.sendNextScheduled()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.nextScheduled:
			if err := s.schedule(ctx); err != nil {
				return err
			}
			s.sendNextScheduled()
		}
	}
}

// sendNextScheduled 立即触发一次任务调度
func (s *Server) sendNextScheduled() {
	s.nextScheduledLocker.Lock()
	if len(s.nextScheduled) == 0 {
		s.nextScheduled <- struct{}{} // 只在 sendNextScheduled 中写入，所以可以保证正常写入。
	}
	s.nextScheduledLocker.Unlock()
}

// clearAndSendNextScheduled 清除现在有任务并立即触发一次任务调度
//
// 如果有正在运行的 schedule，则会让其退出。
func (s *Server) clearAndSendNextScheduled() {
	s.nextScheduledLocker.Lock()

	if len(s.exitSchedule) == 0 { // 让之前正在运行的 schedule 退出
		s.exitSchedule <- struct{}{} // 只在 sendNextScheduled 中写入，所以可以保证正常写入。
	}

	if len(s.nextScheduled) == 0 {
		s.nextScheduled <- struct{}{} // 只在 sendNextScheduled 中写入，所以可以保证正常写入。
	}

	s.nextScheduledLocker.Unlock()
}

// 调度计划任务
//
// 每完成一个计划任务时，都会调用此函数重新计算调度时间，并重新生成一个最近时间的定时器。
func (s *Server) schedule(ctx context.Context) error {
	s.scheduleLocker.Lock()
	defer s.scheduleLocker.Unlock()

	sortJobs(s.jobs) // 按执行时间进行排序

	now := time.Now()

	dur := time.Minute
	if len(s.jobs) > 0 && !s.jobs[0].Next().IsZero() {
		dur = s.jobs[0].Next().Sub(now)
	}

	if dur > 0 { // dur > 0 表示没有需要立即执行的，根据最早的一条任务做一个计时器。
		timer := time.NewTimer(dur)
		for {
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C: // 计时结束，表示 jobs 没有变化，直接跳至 LOOP 部分执行
				goto LOOP
			case <-s.exitSchedule:
				timer.Stop()
				return nil
			}
		}
	}

LOOP:
	for _, j := range s.jobs {
		if next := j.Next(); next.After(now) || next.IsZero() {
			break
		}

		if j.State() == Running && j.Delay() {
			continue
		}

		// j.run 启动需要时间，可能存在 j.run 未初始化完成，第二次调用已经开始，
		// 所以此处先初始化相关的状态信息，使第二次调用处理非法状态。
		j.calcState(now)
		go j.run(now, s.erro, s.info)
	}

	if len(s.exitSchedule) > 0 { // 在退出函数和执行 sendNextScheduled 之前清空 exitSchedule
		<-s.exitSchedule
	}

	return nil
}
