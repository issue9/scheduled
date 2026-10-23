// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package every

import (
	"testing"
	"time"

	"github.com/issue9/assert/v6"
)

func TestParseCron(t *testing.T) {
	const last = "2024-01-01T10:30:30Z"

	data := []struct {
		name string
		spec string
		want string // 空表示期望返回零值
	}{
		{
			name: "每秒",
			spec: "* * * * * *",
			want: "2024-01-01T10:30:31Z",
		},
		{
			name: "每分钟",
			spec: "0 * * * * *",
			want: "2024-01-01T10:31:00Z",
		},
		{
			name: "每小时",
			spec: "0 0 * * * *",
			want: "2024-01-01T11:00:00Z",
		},
		{
			name: "每天",
			spec: "0 0 0 * * *",
			want: "2024-01-02T00:00:00Z",
		},
		{
			name: "每周日",
			spec: "0 0 0 * * 0",
			want: "2024-01-07T00:00:00Z",
		},
		{
			name: "每周六",
			spec: "0 0 0 * * 6",
			want: "2024-01-06T00:00:00Z",
		},
		{
			name: "每周一至周五",
			spec: "0 0 0 * * 1-5",
			want: "2024-01-02T00:00:00Z",
		},
		{
			name: "每年",
			spec: "0 0 0 1 1 *",
			want: "2025-01-01T00:00:00Z",
		},
		{
			name: "每月",
			spec: "0 0 0 1 * *",
			want: "2024-02-01T00:00:00Z",
		},
		{
			name: "日期范围",
			spec: "0 0 0 1-3 * *",
			want: "2024-01-02T00:00:00Z",
		},
		{
			name: "月份范围",
			spec: "0 0 0 * 1-2 *",
			want: "2024-01-02T00:00:00Z",
		},
		{
			name: "日与星期取更早的-星期",
			spec: "0 0 0 5 * 3",
			want: "2024-01-03T00:00:00Z",
		},
		{
			name: "日与星期取更早的-日期",
			spec: "0 0 0 2 * 5",
			want: "2024-01-02T00:00:00Z",
		},
		{
			name: "闰年 2-29",
			spec: "0 0 0 29 2 *",
			want: "2024-02-29T00:00:00Z",
		},
		{
			name: "没有匹配的日期",
			spec: "0 0 0 31 2 *",
			want: "",
		},
		{
			name: "分钟列表乱序",
			spec: "0 30,0 * * * *",
			want: "2024-01-01T11:00:00Z",
		},
		{
			name: "分钟列表有序",
			spec: "0 0,30 * * * *",
			want: "2024-01-01T11:00:00Z",
		},
		{
			name: "@hourly",
			spec: "@hourly",
			want: "2024-01-01T11:00:00Z",
		},
		{
			name: "@daily",
			spec: "@daily",
			want: "2024-01-02T00:00:00Z",
		},
		{
			name: "@midnight",
			spec: "@midnight",
			want: "2024-01-02T00:00:00Z",
		},
		{
			name: "@weekly",
			spec: "@weekly",
			want: "2024-01-07T00:00:00Z",
		},
		{
			name: "@monthly",
			spec: "@monthly",
			want: "2024-02-01T00:00:00Z",
		},
		{
			name: "@yearly",
			spec: "@yearly",
			want: "2025-01-01T00:00:00Z",
		},
		{
			name: "@annually",
			spec: "@annually",
			want: "2025-01-01T00:00:00Z",
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			a := assert.New(t, false)
			s, err := ParseCron(item.spec)
			a.NotError(err).NotNil(s)
			checkNext(t, s, last, item.want)
		})
	}
}

func TestParseCronError(t *testing.T) {
	data := []struct {
		name string
		spec string
	}{
		{
			name: "空字符串",
			spec: "",
		},
		{
			name: "字段数不足",
			spec: "0 0 0 * *",
		},
		{
			name: "字段数过多",
			spec: "0 0 0 * * * *",
		},
		{
			name: "未知指令",
			spec: "@unknown",
		},
		{
			name: "指令区分大小写",
			spec: "@DAILY",
		},
		{
			name: "秒字段非数字",
			spec: "xx 0 0 * * *",
		},
		{
			name: "秒字段越界",
			spec: "60 0 0 * * *",
		},
		{
			name: "分字段越界",
			spec: "0 60 0 * * *",
		},
		{
			name: "时字段越界",
			spec: "0 0 24 * * *",
		},
		{
			name: "日字段为 0",
			spec: "0 0 0 0 * *",
		},
		{
			name: "日字段越界",
			spec: "0 0 0 32 * *",
		},
		{
			name: "月字段为 0",
			spec: "0 0 0 * 0 *",
		},
		{
			name: "月字段越界",
			spec: "0 0 0 * 13 *",
		},
		{
			name: "星期字段越界",
			spec: "0 0 0 * * 8",
		},
		{
			name: "反向范围",
			spec: "0 10-5 * * * *",
		},
		{
			name: "超长数字",
			spec: "100 0 0 * * *",
		},
		{
			name: "超长非数字",
			spec: "abc 0 0 * * *",
		},
		{
			name: "重复值",
			spec: "0 0,0 0 * * *",
		},
		{
			name: "范围与值重复",
			spec: "0 0 0 * 1-2,1 *",
		},
		{
			name: "星期重复值",
			spec: "0 0 0 * * 1,1",
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			a := assert.New(t, false)
			s, err := ParseCron(item.spec)
			a.Error(err).Nil(s)
		})
	}
}

func TestParseCronReboot(t *testing.T) {
	a := assert.New(t, false)

	s, err := ParseCron("@reboot")
	a.NotError(err).NotNil(s)

	first := s.Next(time.Time{})
	a.False(first.IsZero())
	a.True(s.Next(first).IsZero())
}

func TestParseField(t *testing.T) {
	data := []struct {
		name   string
		field  string
		b      bound[int]
		want   []int
		hasErr bool
	}{
		{
			name:  "任意值",
			field: "*",
			b:     minuteBound,
			want:  nil,
		},
		{
			name:  "单个值",
			field: "5",
			b:     minuteBound,
			want:  []int{5},
		},
		{
			name:  "逗号有序",
			field: "0,30",
			b:     minuteBound,
			want:  []int{0, 30},
		},
		{
			// 输出必须升序，nextValue 依赖该顺序。
			name:  "逗号乱序",
			field: "30,0",
			b:     minuteBound,
			want:  []int{0, 30},
		},
		{
			name:  "范围",
			field: "10-12",
			b:     minuteBound,
			want:  []int{10, 11, 12},
		},
		{
			name:  "范围与值",
			field: "10-12,5",
			b:     minuteBound,
			want:  []int{5, 10, 11, 12},
		},
		{
			name:   "反向范围",
			field:  "10-5",
			b:      minuteBound,
			hasErr: true,
		},
		{
			name:   "超长数字",
			field:  "100",
			b:      minuteBound,
			hasErr: true,
		},
		{
			name:   "非数字",
			field:  "abc",
			b:      minuteBound,
			hasErr: true,
		},
		{
			name:   "越界",
			field:  "60",
			b:      minuteBound,
			hasErr: true,
		},
		{
			name:   "负值",
			field:  "-1",
			b:      minuteBound,
			hasErr: true,
		},
		{
			name:   "重复值",
			field:  "1,1",
			b:      minuteBound,
			hasErr: true,
		},
		{
			name:   "范围与值重复",
			field:  "1-2,1",
			b:      minuteBound,
			hasErr: true,
		},
	}

	for _, item := range data {
		t.Run(item.name, func(t *testing.T) {
			a := assert.New(t, false)
			got, err := parseField(item.field, item.b)
			if item.hasErr {
				a.Error(err)
				return
			}
			a.NotError(err).Equal(item.want, got)
		})
	}

	t.Run("星期列表排序", func(t *testing.T) {
		a := assert.New(t, false)
		got, err := parseField("5,1", weekdayBound)
		a.NotError(err).Equal([]time.Weekday{time.Monday, time.Friday}, got)
	})
}

func TestFromCron(t *testing.T) {
	a := assert.New(t, false)

	a.NotPanic(func() { FromCron("@daily") })
	a.Panic(func() { FromCron("invalid") })
}
