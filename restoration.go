package flutter

import (
	"github.com/go-flutter-desktop/go-flutter/plugin"
)

type restorationPlugin struct{}

// all hardcoded because theres not pluggable renderer system.
var defaultRestorationPlugin = &restorationPlugin{}

var _ Plugin = &restorationPlugin{} // compile-time type check

func (p *restorationPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	channel := plugin.NewMethodChannel(messenger, "flutter/restoration", plugin.StandardMethodCodec{})
	// Ignored: desktop doesn't need application "restoration"
	channel.HandleFunc("get", func(_ interface{}) (interface{}, error) { return nil, nil })
	return nil
}
