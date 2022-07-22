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

func (p *navigationPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	p.channel = plugin.NewMethodChannel(messenger, navigationChannelName, plugin.JSONMethodCodec{})

	// Ignored: This information isn't properly formated to set the window.SetTitle
	p.channel.HandleFuncSync("routeUpdated", func(_ interface{}) (interface{}, error) { return nil, nil })

	// Currently ignored on platforms other than web
	p.channel.HandleFuncSync("selectSingleEntryHistory", func(_ interface{}) (interface{}, error) { return nil, nil })
	p.channel.HandleFuncSync("routeInformationUpdated", func(_ interface{}) (interface{}, error) { return nil, nil })

	return nil
}
