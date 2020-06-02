package keyboard

import (
	"fmt"
	"runtime/debug"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/pkg/errors"
)

// Event corresponds to a Flutter (dart) compatible RawKeyEventData keyevent data.
// Multi-platform keycode translation is handled within this package.
//
// As input, go-flutter gets GLFW-keyevent who are only compatible with
// RawKeyEventDataLinux. To fully support keyboard shortcut (like Command+C to
// copy on darwin), the flutter framework expect the sent keyevent data to be
// in the form of a RawKeyEventDataMacOs keyevent data.
// This package maps the GLFW key-codes to the MacOS ones.
//
// On the flutter framework side:
//    RawKeyEventDataMacOs data is received for darwin
//    RawKeyEventDataLinux data is received for linux and windows
type Event struct {
	// Common
	Keymap    string `json:"keymap"` // Linux/MacOS switch
	Character string `json:"character"`
	KeyCode   int    `json:"keyCode"`
	Modifiers int    `json:"modifiers"`
	Type      string `json:"type"`

	// Linux
	Toolkit             string `json:"toolkit,omitempty"`
	ScanCode            int    `json:"scanCode,omitempty"`
	UnicodeScalarValues uint32 `json:"unicodeScalarValues,omitempty"`

	// MacOS
	CharactersIgnoringModifiers string `json:"charactersIgnoringModifiers,omitempty"`
	Characters                  string `json:"characters,omitempty"`
}

// Normalize takes a GLFW-based key and normalizes it by converting
// the input to a keyboard.Event struct compatible with Flutter for the current
// OS.
//
//    RawKeyEventDataMacOs data for darwin
//    RawKeyEventDataLinux data for linux and windows
func Normalize(key glfw.Key, scancode int, mods glfw.ModifierKey, action glfw.Action) (event Event, err error) {
	var typeKey string
	if action == glfw.Release {
		typeKey = "keyup"
	} else if action == glfw.Press {
		typeKey = "keydown"
	} else if action == glfw.Repeat {
		typeKey = "keydown"
	} else {
		return event, errors.Errorf("unknown key event type: %v\n", action)
	}

	defer func() {
		p := recover()
		if p != nil {
			fmt.Printf("go-flutter: recovered from panic while handling %s event: %v\n", typeKey, p)
			debug.PrintStack()
		}
	}()

	// This function call can fail with panic()
	utf8 := glfw.GetKeyName(key, scancode)

	event = Event{
		Type:                typeKey,
		Character:           utf8,
		UnicodeScalarValues: codepointFromGLFWKey([]rune(utf8)),
	}

	event.platfromNormalize(key, scancode, mods)

	return event, nil
}

// codepointFromGLFWKey Queries for the printable key name given a key.
// The Flutter framework accepts only one code point, therefore, only the first
// code point will be used. There is unlikely to be more than one, but there
// is no guarantee that it won't happen.
func codepointFromGLFWKey(utf8 []rune) uint32 {
	if len(utf8) == 0 {
		return 0
	}
	return uint32(utf8[0])
}
