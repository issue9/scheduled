// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package scheduled

import (
	"encoding"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/localeutil"
)

var (
	_ encoding.TextMarshaler   = Stopped
	s State                    = 1
	_ encoding.TextUnmarshaler = &s
	_ Logger                   = &defaultLogger{}
)

func TestMarshal(t *testing.T) {
	a := assert.New(t, false)

	for k, v := range stateStringMap {
		text, err := k.MarshalText()
		a.NotError(err).
			Equal(string(text), v).
			Equal(k.String(), v)

		var s State = -1
		a.NotError(s.UnmarshalText(text))
		a.Equal(s, k)
	}

	// 无效的值

	var s State = -1
	a.Equal(s.String(), "<unknown>")

	text, err := s.MarshalText()
	a.Nil(text).Equal(err, localeutil.Error("invalid state %d", s))

	err = s.UnmarshalText([]byte("not-exists"))
	a.Equal(err, localeutil.Error("invalid state text %s", "not-exists"))
}
