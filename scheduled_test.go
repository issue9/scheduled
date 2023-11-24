// SPDX-License-Identifier: MIT

package scheduled

import (
	"encoding"
	"testing"

	"github.com/issue9/assert/v3"
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
	a.Nil(text).ErrorString(err, "无效的值")

	err = s.UnmarshalText([]byte("not-exists"))
	a.ErrorString(err, "无效的值")
}
