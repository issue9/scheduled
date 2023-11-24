// SPDX-License-Identifier: MIT

//go:generate web locale -l=und -func=github.com/issue9/scheduled.Logger.Printf -m -f=yaml ./
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
	Logger interface {
		Error(error) // 输出 error 对象到日志
		Print(...interface{})
		Printf(format string, v ...interface{})
	}

	State int8

	defaultLogger struct{}
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

func (l *defaultLogger) Error(error) {}

func (l *defaultLogger) Print(v ...interface{}) {}

func (l *defaultLogger) Printf(format string, v ...interface{}) {}

func (s State) String() string {
	v, found := stateStringMap[s]
	if !found {
		v = "<unknown>"
	}
	return v
}

func (s State) MarshalText() ([]byte, error) {
	v, found := stateStringMap[s]
	if found {
		return []byte(v), nil
	}
	return nil, fmt.Errorf("无效的值 %v", s)
}

func (s *State) UnmarshalText(data []byte) error {
	v, found := stringStateMap[string(data)]
	if !found {
		return fmt.Errorf("无效的值 %v", string(data))
	}
	*s = v
	return nil
}
