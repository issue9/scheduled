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
//   - */n n 为步长，* 表示当前域的最小和最大值，* 也可以自己指定，比如 5/10 表示从 5 开始，步长为 10，5-20/10 表示从 5 到 20，步长为 10
//
// 同时支持以下便捷指令：
//
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
		return nil, syntaxError(localeutil.Error("can not be empty"))
	case spec[0] == '@':
		d, found := direct[spec]
		if !found {
			return nil, syntaxError(localeutil.Error("invalid direct %s", spec))
		}
		spec = d
	}

	fs := strings.Fields(spec)
	if len(fs) != 6 {
		return nil, syntaxError(localeutil.Error("incorrect length"))
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
			weekdays, err = parseField(field, cronWeekdayBound)
		}

		if err != nil {
			return nil, err
		}
	}

	// 将星期中的 7 转换为 0
	if len(weekdays) > 0 {
		for i, w := range weekdays {
			if w == 7 {
				weekdays[i] = time.Sunday
			}
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

	fs := strings.Split(field, ",")
	list := make([]T, 0, len(fs))

	for _, v := range fs {
		v = strings.TrimSpace(v)

		if v == "" {
			return nil, syntaxError(localeutil.Error("invalid format %s", field))
		}

		if strings.Contains(v, "/") {
			base, stepText, ok := strings.Cut(v, "/")
			if !ok || stepText == "" || strings.Contains(stepText, "/") {
				return nil, syntaxError(localeutil.Error("invalid format %s", v))
			}

			// 从 base 中获取最大和最小值

			minV, maxV := b.min, b.max
			if base != "*" {
				minVal, maxVal, ok, err := parseRange(base, b)
				if err != nil {
					return nil, syntaxError(err)
				}
				if !ok {
					minV = T(minVal)
				} else {
					minV, maxV = T(minVal), T(maxVal)
				}
			}

			// 解析步长
			step, err := strconv.Atoi(stepText)
			if err != nil || step <= 0 || maxV == -1 {
				return nil, syntaxError(localeutil.Error("invalid format %s", v))
			}

			// 根据步长和范围将合法的值添加到 list
			for n, max := int(minV), int(maxV); n <= max; {
				list = append(list, T(n))
				if step > max-n {
					break
				}
				n += step
			}

			// 步长没有给定值
			if len(list) == 0 {
				return nil, syntaxError(localeutil.Error("invalid format %s", v))
			}

			continue
		}

		if len(v) <= 2 { // 少于 3 个字符，说明不可能有特殊字符。
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, syntaxError(err)
			}

			if !b.valid(n) {
				return nil, syntaxError(localeutil.Error("the value %d out of range [%d,%d]", n, b.min, b.max))
			}

			list = append(list, T(n))
			continue
		}

		n1, n2, hasRange, err := parseRange(v, b)
		if err != nil {
			return nil, syntaxError(err)
		}

		if hasRange {
			for i := n1; i <= n2; i++ {
				list = append(list, T(i))
			}
		} else {
			list = append(list, T(n1))
		}
	}

	list = sliceutil.Unique(list, func(i, j T) bool { return i == j })
	slices.Sort(list)
	return list, nil
}

func parseRange[T integer](field string, b bound[T]) (n1, n2 int, ok bool, err error) {
	if before, after, oo := strings.Cut(field, "-"); oo {

		if n1, err = strconv.Atoi(before); err != nil {
			return 0, 0, false, syntaxError(localeutil.Error("invalid format %s", err))
		}

		if n2, err = strconv.Atoi(after); err != nil {
			return 0, 0, false, syntaxError(localeutil.Error("invalid format %s", err))
		}

		ok = true
	} else {
		n, err := strconv.Atoi(field)
		if err != nil {
			return 0, 0, false, syntaxError(localeutil.Error("invalid format %s", err))
		}
		n1, n2 = n, n
	}

	if n1 > n2 {
		return 0, 0, false, syntaxError(localeutil.Error("invalid range %s", field))
	}
	if !b.valid(n1) {
		return 0, 0, false, syntaxError(localeutil.Error("the value %d out of range [%d,%d]", n1, b.min, b.max))
	}
	if !b.valid(n2) {
		return 0, 0, false, syntaxError(localeutil.Error("the value %d out of range [%d,%d]", n2, b.min, b.max))
	}

	return
}

func syntaxError(s error) error { return localeutil.Error("cron syntax error %s", s) }
