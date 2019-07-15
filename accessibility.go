package flutter

import "github.com/go-flutter-desktop/go-flutter/plugin"

type accessibilityPlugin struct{}

var _ Plugin = &accessibilityPlugin{} // compile-time type check

// hardcoded because there is no swappable renderer interface.
var defaultAccessibilityPlugin = &accessibilityPlugin{}

func (p *accessibilityPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	channel := plugin.NewBasicMessageChannel(messenger, "flutter/accessibility", plugin.StandardMessageCodec{})
	// Ignored: go-flutter dosn't support accessibility events
	channel.HandleFunc(func(_ interface{}) (interface{}, error) { return nil, nil })
	return nil
}
