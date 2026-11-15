// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package schedulers

import "time"

// Earliest 创建一个从众多调度器中选择最近时间点的调度器
//
// 所有小于 last 的时间点将被忽略
func Earliest(s ...Scheduler) Scheduler {
	if len(s) == 0 {
		panic("参数 s 不能为空")
	}

	return SchedulerFunc(func(last time.Time) time.Time {
		var early time.Time
		for _, ss := range s {
			v := ss.Next(last)
			if v.IsZero() || v.Before(last) { // 终结或已过期的项直接忽略
				continue
			}
			if early.IsZero() || v.Before(early) {
				early = v
			}
		}
		return early
	})
}
