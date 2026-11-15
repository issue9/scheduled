// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package schedulers

import "time"

// At 返回只在指定时间执行一次的调度器
func At(t time.Time) Scheduler {
	return SchedulerFunc(func(last time.Time) time.Time {
		if !t.After(last) {
			return time.Time{}
		}
		return t
	})
}
