// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

//go:generate gomsgs extract -l=und -f=yaml ./
//go:generate gomsgs update -src=./locales/und.yaml -dest=./locales/zh.yaml

// Package scheduled 计划任务管理工具
//
// 目前 scheduled 内置了以下三种算法：
//   - every 实现了 crontab 中的大部分语法功能以及自定义的 Every 方法；
//   - at 在固定的时间点执行一次任务；
//   - ticker 以固定的时间段执行任务，与 [time.Ticker] 相同。
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
	Cancelled // 取消，一般由用户主动发起，比如：Server 关闭，或是 Server.New 返回的函数主动关闭。
)

func errServerState(s State) error {
	return localeutil.Error("server state is %v, this operation is not allowed", s)
}

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
		Stopped:   "stopped",
		Running:   "running",
		Failed:    "failed",
		Cancelled: "cancelled",
	}

	stringStateMap = map[string]State{
		"stopped":   Stopped,
		"running":   Running,
		"failed":    Failed,
		"cancelled": Cancelled,
	}
)

func any2Error(msg any) error {
	if e, ok := msg.(error); ok {
		return e
	} else {
		return localeutil.Error("recover msg %v", msg)
	}
}

// NewTermLogger 声明输出到终端的日志
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
