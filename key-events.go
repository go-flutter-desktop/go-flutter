package flutter

import (
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/go-flutter-desktop/go-flutter/internal/glfwkeyconversion"
	"github.com/go-flutter-desktop/go-flutter/plugin"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const keyEventChannelName = "flutter/keyevent"

// keyeventPlugin implements flutter.Plugin and handles method calls to
// flutter/keyevent channel.
// The sent keyevents are RawKeyEventDataLinux for every platform.
// RawKeyEventDataLinux is the only flutter implementation that support the
// GLFW Toolkit, the events data is sent as it is, no translation is made in
// go-flutter, everything is handled by the flutter framework (dart).
type keyeventPlugin struct {
	keyEventChannel *plugin.BasicMessageChannel
}

var defaultKeyeventsPlugin = &keyeventPlugin{}

var _ Plugin = &keyeventPlugin{} // compile-time type check

func (p *keyeventPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	p.keyEventChannel = plugin.NewBasicMessageChannel(messenger, keyEventChannelName, keyEventJSONMessageCodec{})
	return nil
}

type keyEventJSONMessageCodec struct{}

// EncodeMessage encodes a keyEventMessage to a slice of bytes.
func (j keyEventJSONMessageCodec) EncodeMessage(message interface{}) (binaryMessage []byte, err error) {
	return json.Marshal(message)
}

// send-only channel
func (j keyEventJSONMessageCodec) DecodeMessage(binaryMessage []byte) (message interface{}, err error) {
	return message, err
}

type keyEventMessage struct {
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

// Flutter only support keydown & keyup events type.
// On GLFW we are receiving glfw.Repeat events, those events are,
// unknown to the Flutter Framework.
// To stay consistent with other embedders, glfw.Repeat is treated as a
// keydown event.
func (p *keyeventPlugin) sendKeyEvent(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

	var typeKey string
	if action == glfw.Release {
		typeKey = "keyup"
	} else if action == glfw.Press {
		typeKey = "keydown"
	} else if action == glfw.Repeat {
		typeKey = "keydown"
	} else {
		fmt.Printf("go-flutter: Unknown key event type: %v\n", action)
		return
	}

	defer func() {
		p := recover()
		if p != nil {
			fmt.Printf("go-flutter: recovered from panic while handling %s event: %v\n", typeKey, p)
			debug.PrintStack()
		}
	}()

	utf8 := glfw.GetKeyName(key, scancode)
	var event keyEventMessage

	if runtime.GOOS == "darwin" {
		macosMods := glfwkeyconversion.ToMacOSModifiers(mods)
		if val, ok := glfwkeyconversion.AsMacOSModifiers(key); ok {
			// On GLFW, the "modifiers" keycode is the state as it is BEFORE this event
			// happened, not AFTER, like every other platform.
			macosMods = val | int(macosMods)
		}

		event = keyEventMessage{
			KeyCode:                     glfwkeyconversion.ToMacOSKeyCode(key, scancode),
			Keymap:                      "macos",
			Type:                        typeKey,
			Character:                   utf8,
			CharactersIgnoringModifiers: utf8,
			Characters:                  utf8,
			Modifiers:                   macosMods,
		}
	} else {
		event = keyEventMessage{
			KeyCode:   int(key),
			Keymap:    "linux",
			Toolkit:   "glfw",
			Type:      typeKey,
			ScanCode:  scancode,
			Modifiers: int(mods),
			Character: utf8,
		}
	}

	if len(utf8) > 0 {
		event.UnicodeScalarValues = CodepointFromGLFWKey([]rune(utf8))
	}

	if err := p.keyEventChannel.Send(event); err != nil {
		fmt.Printf("go-flutter: Failed to send raw_keyboard event %v: %v\n", event, err)
	}

}

// CodepointFromGLFWKey Queries GLFW for the printable key name given a key.
// The Flutter framework accepts only one code point, therefore, only the first
// code point will be used. There is unlikely to be more than one, but there
// is no guarantee that it won't happen.
func CodepointFromGLFWKey(utf8 []rune) uint32 {
	return uint32(utf8[0])
}
