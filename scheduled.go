// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

//go:generate web locale -l=und -func=github.com/issue9/localeutil.Phrase,github.com/issue9/localeutil.Error -m -f=yaml ./
//go:generate web update-locale -src=./locales/und.yaml -dest=./locales/zh-CN.yaml

// Package scheduled 个计划任务管理工具
//
// 通过 scheduled 可以实现管理类似 linux 中 crontab 功能的计划任务功能。
// 当然功能并不止于此，用户可以实现自己的调度算法，定制任务的启动机制。
//
// 目前 scheduled 内置了以下三种算法：
// - cron 实现了 crontab 中的大部分语法功能；
// - at 在固定的时间点执行一次任务；
// - ticker 以固定的时间段执行任务，与 [time.Ticker] 相同。
package scheduled

import (
	"fmt"

	"github.com/issue9/localeutil"

	"github.com/issue9/scheduled/schedulers"
)

// 任务的几种状态
const (
	Stopped State = iota
	Running
	Failed
)

type (
	Scheduler     = schedulers.Scheduler
	SchedulerFunc = schedulers.SchedulerFunc

	// Logger 日志接口
	//
	// NOTE: 同时实现了 [github.com/issue9/logs.Logger] 对象
	Logger interface {
		// Error 输出错误对象到日志
		Error(error)

		// LocaleString 输出本地化的内容到日志
		LocaleString(localeutil.Stringer)
	}

	State int8

	defaultLogger struct{}

	termLogger struct {
		p *localeutil.Printer
	}
)

var (
	stateStringMap = map[State]string{
		Stopped: "stopped",
		Running: "running",
		Failed:  "failed",
	}

	stringStateMap = map[string]State{
		"stopped": Stopped,
		"running": Running,
		"failed":  Failed,
	}
)

// NewTermLogger 声明输出到终的日志
func NewTermLogger(p *localeutil.Printer) Logger { return &termLogger{p: p} }

func (l *defaultLogger) Error(error) {}

func (l *defaultLogger) LocaleString(localeutil.Stringer) {}

func (l *termLogger) Error(err error) { fmt.Println(err) }

func (l *termLogger) LocaleString(s localeutil.Stringer) { fmt.Println(s.LocaleString(l.p)) }

func (s State) String() string {
	v, found := stateStringMap[s]
	if !found {
		v = "<unknown>"
	}
	return v
}

func (s State) MarshalText() ([]byte, error) {
	if v, found := stateStringMap[s]; found {
		return []byte(v), nil
	}
	return nil, localeutil.Error("invalid state %d", s)
}

func (s *State) UnmarshalText(data []byte) error {
	if v, found := stringStateMap[string(data)]; found {
		*s = v
		return nil
	}
	return localeutil.Error("invalid state text %s", string(data))
}
