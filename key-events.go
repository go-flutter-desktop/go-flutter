package flutter

import (
	"encoding/json"
	"fmt"

	"github.com/go-flutter-desktop/go-flutter/plugin"
	"github.com/go-gl/glfw/v3.2/glfw"
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
	Toolkit   string `json:"toolkit"`
	KeyCode   int    `json:"keyCode"`
	Type      string `json:"type"`
	ScanCode  int    `json:"scanCode"`
	Modifiers int    `json:"modifiers"`
	Keymap    string `json:"keymap"`
}

// Flutter only support keydown & keyup events type.
// 2019-05-19, LOC of the flutter keyevent handler:
// https://github.com/flutter/flutter/blob/7a4c33425ddd78c54aba07d86f3f9a4a0051769b/packages/flutter/lib/src/services/raw_keyboard.dart#L291-L298
// On GLFW we are receiving glfw.Repeat events, those events are,
// unknown to the Flutter Framework.
// To stay consistent with other embedders, glfw.Repeat is "seen" as a
// keydown event.
// Comment about RawKeyEvent on others embedders:
// https://github.com/go-flutter-desktop/go-flutter#issuecomment-494998771
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

	event := keyEventMessage{
		KeyCode:   int(key),
		Keymap:    "linux",
		Toolkit:   "glfw",
		Type:      typeKey,
		ScanCode:  scancode,
		Modifiers: int(mods),
	}
	err := p.keyEventChannel.Send(event)
	if err != nil {
		fmt.Printf("go-flutter: Failed to send raw_keyboard event %v: %v\n", event, err)
	}

}
