package flutter

import (
	"encoding/json"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/pkg/errors"

	"github.com/go-flutter-desktop/go-flutter/internal/tasker"
	"github.com/go-flutter-desktop/go-flutter/plugin"
)

// platformPlugin implements flutter.Plugin and handles method calls to the
// flutter/platform channel.
type platformPlugin struct {
	popBehavior popBehavior

	messenger  plugin.BinaryMessenger
	glfwTasker *tasker.Tasker
	window     *glfw.Window
	channel    *plugin.MethodChannel

	// flutterInitialized is used as callbacks to know when the flutter framework
	// is running and ready to process upstream plugin calls.
	// (It usually takes ~10 rendering frame).
	// flutterInitialized is trigger when the plugin "flutter/platform" received
	// a message from "SystemChrome.setApplicationSwitcherDescription".
	flutterInitialized []func()
}

// hardcoded because there is no swappable renderer interface.
var defaultPlatformPlugin = &platformPlugin{
	popBehavior: PopBehaviorNone,
}

var _ Plugin = &platformPlugin{}     // compile-time type check
var _ PluginGLFW = &platformPlugin{} // compile-time type check

func (p *platformPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	p.messenger = messenger
	p.glfwTasker = tasker.New()
	return nil
}

func (p *platformPlugin) InitPluginGLFW(window *glfw.Window) (err error) {
	p.window = window
	p.channel = plugin.NewMethodChannel(p.messenger, "flutter/platform", plugin.JSONMethodCodec{})

	p.channel.HandleFunc("Clipboard.setData", p.handleClipboardSetData)
	p.channel.HandleFunc("Clipboard.getData", p.handleClipboardGetData)
	p.channel.HandleFunc("SystemChrome.setApplicationSwitcherDescription", p.handleWindowSetTitle)
	p.channel.HandleFunc("SystemNavigator.pop", p.handleSystemNavigatorPop)
	// Ignored: Desktop's don't have system overlays
	p.channel.HandleFuncSync("SystemChrome.setSystemUIOverlayStyle", func(_ interface{}) (interface{}, error) { return nil, nil })
	// Ignored: Desktop's don't have haptic feedback
	p.channel.HandleFuncSync("HapticFeedback.vibrate", func(_ interface{}) (interface{}, error) { return nil, nil })
	// Ignored: Desktop's don't play sound on every clicks
	p.channel.HandleFuncSync("SystemSound.play", func(_ interface{}) (interface{}, error) { return nil, nil })
	return nil
}

func (p *platformPlugin) handleClipboardSetData(arguments interface{}) (reply interface{}, err error) {
	newClipboard := struct {
		Text string `json:"text"`
	}{}
	err = json.Unmarshal(arguments.(json.RawMessage), &newClipboard)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode json arguments for handleClipboardSetData")
	}
	p.glfwTasker.Do(func() {
		p.window.SetClipboardString(newClipboard.Text)
	})
	return nil, nil
}

func (p *platformPlugin) handleClipboardGetData(arguments interface{}) (reply interface{}, err error) {
	requestedMime := ""
	err = json.Unmarshal(arguments.(json.RawMessage), &requestedMime)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode json arguments for handleClipboardGetData")
	}
	if requestedMime != "text/plain" {
		return nil, errors.New("obtaining mime type " + requestedMime + " from clipboard is not yet supported in go-flutter")
	}

	var clipText string
	p.glfwTasker.Do(func() {
		clipText = p.window.GetClipboardString()
	})

	reply = struct {
		Text string `json:"text"`
	}{
		Text: clipText,
	}
	return reply, nil
}

func (p *platformPlugin) handleWindowSetTitle(arguments interface{}) (reply interface{}, err error) {
	// triggers flutter framework initialized callbacks
	for _, f := range p.flutterInitialized {
		f()
	}

	return nil, nil
}

// addFrameworkReadyCallback adds a callback which if trigger when the flutter
// framework is ready.
func (p *platformPlugin) addFrameworkReadyCallback(f func()) {
	p.flutterInitialized = append(p.flutterInitialized, f)
}
