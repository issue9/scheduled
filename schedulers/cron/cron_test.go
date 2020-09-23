// SPDX-License-Identifier: MIT

package cron

import "github.com/issue9/scheduled/schedulers"

var _ schedulers.Scheduler = &cron{}
