// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

//go:generate web locale -l=und -func=github.com/issue9/localeutil.Phrase,github.com/issue9/localeutil.Error -m -f=yaml ./
//go:generate web update-locale -src=./locales/und.yaml -dest=./locales/zh.yaml

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

var (
	errServerRunning   = localeutil.Error("server is running, cannot start again")
	errServerCancelled = localeutil.Error("server is cancelled, cannot start again")
)

// ErrServerRunning 表示 [Server] 已经处于运行状态
//
// 当多次调用 [Server.Serve] 时，可能返回此错误信息。
func ErrServerRunning() error { return errServerRunning }

// ErrServerCancelled 表示 [Server] 已经被取消
//
// 当 [Server.Serve] 的 ctx 已经 Done 时，再次调用 [Server.Serve] 可能返回此错误信息。
func ErrServerCancelled() error { return errServerCancelled }

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
