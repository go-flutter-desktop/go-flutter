package flutter

import (
	"github.com/go-flutter-desktop/go-flutter/plugin"
)

type restorationPlugin struct{}

// all hardcoded because theres not pluggable renderer system.
var defaultRestorationPlugin = &restorationPlugin{}

var _ Plugin = &restorationPlugin{} // compile-time type check

// InitPlugin implements PluginGLFW
func (p *restorationPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	channel := plugin.NewMethodChannel(messenger, "flutter/restoration", plugin.StandardMethodCodec{})
	// Currently not supported. Used only to avoid warnings.
	channel.HandleFunc("get", func(_ interface{}) (interface{}, error) { return nil, nil })
	return nil
}
