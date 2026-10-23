// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package every

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"

	"github.com/issue9/scheduled/schedulers"
	"github.com/issue9/scheduled/schedulers/at"
)

// 常用的便捷指令
var direct = map[string]string{
	"@yearly":   "0 0 0 1 1 *",
	"@annually": "0 0 0 1 1 *",
	"@monthly":  "0 0 0 1 * *",
	"@weekly":   "0 0 0 * * 0",
	"@daily":    "0 0 0 * * *",
	"@midnight": "0 0 0 * * *",
	"@hourly":   "0 0 * * * *",
}

// FromCron 从 [crontab] 表达式创建调度器
//
// spec 语法出错，会 panic。如果想要自行处理错误可以调用 [ParseCron]。
//
// [crontab]: https://zh.wikipedia.org/wiki/FromCron
func FromCron(spec string) schedulers.Scheduler {
	s, err := ParseCron(spec)
	if err != nil {
		panic(err)
	}
	return s
}

// ParseCron 将 [crontab] 语法转换为 [schedulers.Scheduler]
//
// 区分大小写，支持秒，其格式如下：
//
//	! * * * * * *
//	  | | | | | |
//	  | | | | | --- 星期
//	  | | | | ----- 月
//	  | | | ------- 日
//	  | | --------- 小时
//	  | ----------- 分
//	  ------------- 秒
//
// 星期与日若同时存在，则以或的形式组合。！用于使 go fmt 不会自动格式化内容，无实际意义。
//
// 支持以下符号：
//   - - 表示范围
//   - , 表示和
//
// 同时支持以下便捷指令：
//
//	@reboot:   启动时执行一次
//	@yearly:   0 0 0 1 1 *
//	@annually: 0 0 0 1 1 *
//	@monthly:  0 0 0 1 * *
//	@weekly:   0 0 0 * * 0
//	@daily:    0 0 0 * * *
//	@midnight: 0 0 0 * * *
//	@hourly:   0 0 * * * *
//
// [crontab]: https://zh.wikipedia.org/wiki/FromCron
func ParseCron(spec string) (schedulers.Scheduler, error) {
	switch {
	case spec == "":
		return nil, syntaxError(localeutil.Phrase("can not be empty"))
	case spec == "@reboot":
		return at.At(time.Now()), nil
	case spec[0] == '@':
		d, found := direct[spec]
		if !found {
			return nil, syntaxError(localeutil.Phrase("invalid direct %s", spec))
		}
		spec = d
	}

	fs := strings.Fields(spec)
	if len(fs) != 6 {
		return nil, syntaxError(localeutil.Phrase("incorrect length"))
	}

	var (
		months                               []time.Month
		weekdays                             []time.Weekday
		years, days, hours, minutes, seconds []int
		err                                  error
	)
	for i, field := range fs {
		switch i {
		case 0:
			seconds, err = parseField(field, secondBound)
		case 1:
			minutes, err = parseField(field, minuteBound)
		case 2:
			hours, err = parseField(field, hourBound)
		case 3:
			days, err = parseField(field, dayBound)
		case 4:
			months, err = parseField(field, monthBound)
		case 5:
			weekdays, err = parseField(field, weekdayBound)
		}

		if err != nil {
			return nil, err
		}
	}

	return schedulers.SchedulerFunc(func(t time.Time) time.Time {
		return next(t, years, months, weekdays, days, hours, minutes, seconds)
	}), nil
}

// 分析单个数字域内容
//
// field 可以是以下格式：
//
//	*
//	n1-n2
//	n1,n2
//	n1-n2,n3-n4,n5
func parseField[T integer](field string, b bound[T]) ([]T, error) {
	if field == "*" {
		return nil, nil
	}

	fs := strings.FieldsFunc(field, func(r rune) bool { return r == ',' })
	list := make([]T, 0, len(fs))

	for _, v := range fs {
		if len(v) <= 2 { // 少于 3 个字符，说明不可能有特殊字符。
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}

			if !b.valid(n) {
				return nil, syntaxError(localeutil.Phrase("the value %d out of range [%d,%d]", n, b.min, b.max))
			}

			list = append(list, T(n))
			continue
		}

		var n1, n2 int
		hasRange := false

		if before, after, ok := strings.Cut(v, "-"); ok {
			var err error

			if n1, err = strconv.Atoi(before); err != nil {
				return nil, syntaxError(localeutil.Phrase("invalid format %s", err))
			}

			if n2, err = strconv.Atoi(after); err != nil {
				return nil, syntaxError(localeutil.Phrase("invalid format %s", err))
			}

			hasRange = true
		} else {
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, syntaxError(localeutil.Phrase("invalid format %s", err))
			}
			n1, n2 = n, n
		}

		if !b.valid(n1) {
			return nil, syntaxError(localeutil.Phrase("the value %d out of range [%d,%d]", n1, b.min, b.max))
		}
		if !b.valid(n2) {
			return nil, syntaxError(localeutil.Phrase("the value %d out of range [%d,%d]", n2, b.min, b.max))
		}
		if n1 > n2 {
			return nil, syntaxError(localeutil.Phrase("invalid range %s", v))
		}

		if hasRange {
			for i := n1; i <= n2; i++ {
				list = append(list, T(i))
			}
		} else {
			list = append(list, T(n1))
		}
	}

	if indexes := sliceutil.Dup(list, func(i, j T) bool { return i == j }); len(indexes) > 0 {
		return nil, syntaxError(localeutil.Phrase("duplicate value %d", list[indexes[0]]))
	}

	slices.Sort(list)
	return list, nil
}

func syntaxError(s localeutil.Stringer) error {
	return localeutil.Error("cron syntax error %s", s)
}
