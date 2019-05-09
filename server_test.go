// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package scheduled

import (
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestServer_Serve(t *testing.T) {
	a := assert.New(t)
	srv := NewServer()
	a.NotNil(srv)
	a.Empty(srv.jobs).
		Equal(srv.Serve(nil), ErrNoJobs)

	srv.NewTicker("tick1", succFunc, 1*time.Second)
	srv.NewTicker("tick2", erroFunc, 2*time.Second)
	go srv.Serve(nil)
	time.Sleep(3 * time.Second)
	srv.Stop()
}

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
