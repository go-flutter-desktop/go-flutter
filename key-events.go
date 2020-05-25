package flutter

import (
	"encoding/json"
	"fmt"

	"github.com/go-flutter-desktop/go-flutter/internal/keyboard"
	"github.com/go-flutter-desktop/go-flutter/plugin"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const keyEventChannelName = "flutter/keyevent"

// keyeventPlugin implements flutter.Plugin and handles method calls to
// flutter/keyevent channel.
// The sent keyevents are RawKeyEventDataLinux on linux and windows
// The sent keyevents are RawKeyEventDataMacOs on darwin (needs a conversion layer)
type keyeventPlugin struct {
	channel *plugin.BasicMessageChannel
}

var defaultKeyeventsPlugin = &keyeventPlugin{}

func (p *keyeventPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	p.channel = plugin.NewBasicMessageChannel(messenger, keyEventChannelName, keyEventJSONMessageCodec{})
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

func (p *keyeventPlugin) sendKeyEvent(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	event, err := keyboard.Normalize(key, scancode, mods, action)
	if err != nil {
		fmt.Printf("go-flutter: failed to Normalize key event: %v", err)
		return
	}

	err = p.channel.Send(event)
	if err != nil {
		fmt.Printf("go-flutter: Failed to send raw_keyboard event %v: %v\n", event, err)
	}
}
