package gutter

import (
	"log"
	"os"

	"github.com/go-gl/glfw/v3.2/glfw"
)

// Option for gutter
type Option func(*config)

// Check path
func checkPath(p string) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		log.Fatal(err)
	}
}

// ProjectAssetPath specify the flutter asset directory.
func ProjectAssetPath(p string) Option {
	checkPath(p)
	return func(c *config) {
		c.AssetPath = p
	}
}

// ApplicationICUDataPath specify the path to the ICUData.
func ApplicationICUDataPath(p string) Option {
	checkPath(p)
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

// ApplicationWindowDimension specify the startup's dimention of the window.
func ApplicationWindowDimension(x int, y int) Option {
	// Check for initial application display size
	if x < 1 {
		log.Fatal("Wrong initial value for width ")
	}

	// Check for initial application display size
	if y < 1 {
		log.Fatal("Wrong initial value for height ")
	}

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
func OptionAddPluginReceiver(handler PluginReceivers, channelName string) Option {
	return func(c *config) {
		// Check for nil, else initialise the map
		if c.PlatformMessageReceivers == nil {
			c.PlatformMessageReceivers = make(map[string][]PluginReceivers)
		}
		c.PlatformMessageReceivers[channelName] =
			append(c.PlatformMessageReceivers[channelName], handler)
	}
}

// OptionKeyboardLayout allow application to support keyboard that have a different layout
// when the FlutterEngine send a PlatformMessage to the Embedder
func OptionKeyboardLayout(keyboardLayout KeyboardShortcuts) Option {
	return func(c *config) {
		c.KeyboardLayout = &keyboardLayout
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
	PlatformMessageReceivers map[string][]PluginReceivers // The Key is the Channel name.
	KeyboardLayout           *KeyboardShortcuts
}

// KeyboardShortcuts Struct where user can define his own keyboard shortcut.
// This will allow application to support keyboard layout different from US layout
type KeyboardShortcuts struct {
	Cut       glfw.Key
	Copy      glfw.Key
	Paste     glfw.Key
	SelectAll glfw.Key
}

func (t config) merge(options ...Option) config {
	for _, opt := range options {
		opt(&t)
	}

	return t
}
