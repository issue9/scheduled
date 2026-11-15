// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package schedulers

import "time"

// Nop 创建一个永远不会执行的调度器
//
// 在某些情况下也会发挥作用，比如和 Server.New 的 imm 参数一起使用，
// 可以起到 @reboot 的功能。
func Nop() Scheduler { return SchedulerFunc(func(time.Time) time.Time { return time.Time{} }) }
