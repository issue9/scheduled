// SPDX-License-Identifier: MIT

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

// 表示任务状态
const (
	Stopped State = iota
	Running
	Failed
)

type Scheduler = schedulers.Scheduler

type Logger interface {
	Print(...interface{})
	Println(...interface{})
	Printf(format string, v ...interface{})
}

// State 状态值类型
type State int8

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
