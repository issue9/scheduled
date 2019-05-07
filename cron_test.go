// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import (
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestSortJobs(t *testing.T) {
	a := assert.New(t)

	now := time.Now()
	jobs := []*Job{
		&Job{
			name: "1",
			next: now.Add(1111),
		},
		&Job{
			name: "2",
			next: time.Time{},
		},
		&Job{
			name: "3",
			next: now,
		},
		&Job{
			name: "4",
			next: time.Time{},
		},
		&Job{
			name: "5",
			next: now.Add(222),
		},
	}

	sortJobs(jobs)
	a.Equal(jobs[0].name, "3").
		Equal(jobs[1].name, "5").
		Equal(jobs[2].name, "1")
}
