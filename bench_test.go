// SPDX-License-Identifier: MIT

package scheduled

import (
	"testing"
	"time"
)

func BenchmarkSortJobs(b *testing.B) {
	now := time.Now()
	jobs := []*Job{
		{
			id:   "1",
			next: now.Add(1111),
		},
		{
			id:   "2",
			next: time.Time{}, // zero 放在最后
		},
		{
			id:   "3",
			next: now,
		},
		{
			id:   "4",
			next: time.Time{}, // zero 放在最后
		},
		{
			id:   "5",
			next: now.Add(222),
		},
	}

	for i := 0; i < b.N; i++ {
		sortJobs(jobs)
		jobs[0], jobs[2] = jobs[2], jobs[0]
	}
}
