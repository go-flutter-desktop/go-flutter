package plugin

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestMessageTypeError(t *testing.T) {
	m := MessageTypeError{
		hint: "unexpected type uintptr",
	}
	Equal(t, "invalid type provided to message codec: unexpected type uintptr", m.Error())
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
	Equal(t, "Error error in Flutter: This is totally wrong ([foo 42 bar])", f.Error())
}
