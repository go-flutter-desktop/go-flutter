package gutter

import (
	"github.com/Drakirus/go-flutter-desktop-embedder/flutter"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// Option for gutter
type Option func(*config)

// OptionAssetPath specify the flutter asset directory.
func OptionAssetPath(p string) Option {
	return func(c *config) {
		c.AssetPath = p
	}
}

// OptionICUDataPath specify the path to the ICUData.
func OptionICUDataPath(p string) Option {
	return func(c *config) {
		c.ICUDataPath = p
	}
}

// OptionVMArguments specify the arguments to the Dart VM.
func OptionVMArguments(a []string) Option {
	return func(c *config) {
		// First should be argument is argv[0]
		c.VMArguments = append([]string{""}, a...)
	}
}

// OptionWindowDimension specify the startup's dimention of the window.
func OptionWindowDimension(x int, y int) Option {
	return func(c *config) {
		c.WindowDimension.x = x
		c.WindowDimension.y = y
	}
}

// OptionWindowInitializer allow initializing the window.
func OptionWindowInitializer(ini func(*glfw.Window) error) Option {
	return func(c *config) {
		c.WindowInitializer = ini
	}
}

// OptionPixelRatio specify the scale factor for the physical screen.
func OptionPixelRatio(ratio float64) Option {
	return func(c *config) {
		c.PixelRatio = ratio
	}
}

// OptionAddPluginReceiver add a new function that will be trigger
// when the FlutterEngine send a PlatformMessage to the Embedder
func OptionAddPluginReceiver(handler PluginReceivers) Option {
	return func(c *config) {
		c.PlatformMessageReceivers = append(c.PlatformMessageReceivers, handler)
	}
}

type config struct {
	WindowDimension struct {
		x int
		y int
	}
	AssetPath                string
	ICUDataPath              string
	WindowInitializer        func(*glfw.Window) error
	PixelRatio               float64
	VMArguments              []string
	PlatformMessageReceivers []PluginReceivers
}

// PluginReceivers do stuff when receiving Message from the Engine,
// send result with `flutterEngine.SendPlatformMessageResponse`
type PluginReceivers func(
	message flutter.PlatformMessage,
	flutterEngine *flutter.EngineOpenGL,
	window *glfw.Window,
) bool

func (t config) merge(options ...Option) config {
	for _, opt := range options {
		opt(&t)
	}

	return t
}
