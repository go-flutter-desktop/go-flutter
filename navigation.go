package flutter

import (
	"github.com/go-flutter-desktop/go-flutter/plugin"
)

const navigationChannelName = "flutter/navigation"

// navigationPlugin implements flutter.Plugin and handles method calls to the
// flutter/navigation channel.
type navigationPlugin struct {
	channel *plugin.MethodChannel
}

// all hardcoded because theres not pluggable renderer system.
var defaultNavigationPlugin = &navigationPlugin{}

var _ Plugin = &navigationPlugin{} // compile-time type check

func (p *navigationPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	p.channel = plugin.NewMethodChannel(messenger, navigationChannelName, plugin.JSONMethodCodec{})
	return nil
}
