package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageTypeError(t *testing.T) {
	m := MessageTypeError{
		hint: "unexpected type uintptr",
	}
	assert.Equal(t, "invalid type provided to message codec: unexpected type uintptr", m.Error())
}

func TestFlutterError(t *testing.T) {
	f := FlutterError{
		Code:    "error",
		Message: "This is totally wrong",
		Details: []interface{}{
			"foo",
			42,
			"bar",
		},
	}
	assert.Equal(t, "Error error in Flutter: This is totally wrong ([foo 42 bar])", f.Error())
}
