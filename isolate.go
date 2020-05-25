package flutter

import "github.com/go-flutter-desktop/go-flutter/plugin"

type isolatePlugin struct{}

// hardcoded because there is no swappable renderer interface.
var defaultIsolatePlugin = &isolatePlugin{}

func (p *isolatePlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	channel := plugin.NewBasicMessageChannel(messenger, "flutter/isolate", plugin.StringCodec{})
	// Ignored: go-flutter dosn't support isolate events
	channel.HandleFunc(func(_ interface{}) (interface{}, error) { return nil, nil })
	return nil
}
