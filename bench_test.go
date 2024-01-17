// SPDX-License-Identifier: MIT

package scheduled

import (
	"testing"
	"time"

	"github.com/issue9/localeutil"
)

func BenchmarkSortJobs(b *testing.B) {
	now := time.Now()
	jobs := []*Job{
		{
			title: localeutil.StringPhrase("1"),
			next:  now.Add(1111),
		},
		{
			title: localeutil.StringPhrase("2"),
			next:  time.Time{}, // zero 放在最后
		},
		{
			title: localeutil.StringPhrase("3"),
			next:  now,
		},
		{
			title: localeutil.StringPhrase("4"),
			next:  time.Time{}, // zero 放在最后
		},
		{
			title: localeutil.StringPhrase("5"),
			next:  now.Add(222),
		},
	}

	for i := 0; i < b.N; i++ {
		sortJobs(jobs)
		jobs[0], jobs[2] = jobs[2], jobs[0]
	}
}
