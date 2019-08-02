package flutter

import (
	"fmt"

	"github.com/go-flutter-desktop/go-flutter/plugin"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const lifecycleChannelName = "flutter/lifecycle"

// lifecyclePlugin implements flutter.Plugin and handles method calls to the
// flutter/lifecycle channel.
type lifecyclePlugin struct {
	channel *plugin.BasicMessageChannel
}

// all hardcoded because theres not pluggable renderer system.
var defaultLifecyclePlugin = &lifecyclePlugin{}

var _ Plugin = &lifecyclePlugin{} // compile-time type check

func (p *lifecyclePlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	p.channel = plugin.NewBasicMessageChannel(messenger, lifecycleChannelName, plugin.StringCodec{})
	return nil
}

func (p *lifecyclePlugin) glfwIconifyCallback(w *glfw.Window, iconified bool) {
	var state string
	switch iconified {
	case true:
		state = "AppLifecycleState.paused"
	case false:
		state = "AppLifecycleState.resumed"
	}
	err := p.channel.Send(state)
	if err != nil {
		fmt.Printf("go-flutter: Failed to send lifecycle event %s: %v\n", state, err)
	}
}
