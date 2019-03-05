package flutter

import (
	"fmt"
	"image"
	"os"

	"github.com/go-gl/glfw/v3.2/glfw"
)

type config struct {
	WindowDimension struct {
		x int
		y int
	}
	AssetsPath                  string
	ICUDataPath                 string
	WindowInitializerDeprecated func(*glfw.Window) error
	WindowIconProvider          func() ([]image.Image, error)
	ForcePixelRatio             float64
	VMArguments                 []string
	PlatformMessageReceivers    map[string][]PluginReceivers // The Key is the Channel name.
	KeyboardLayout              *KeyboardShortcuts
}

func (c config) merge(options ...Option) config {
	for _, option := range options {
		option(&c)
	}

	return c
}

// Option for gutter
type Option func(*config)

// ProjectAssetPath specify the flutter assets directory.
func ProjectAssetPath(p string) Option {
	// deprecated on 2019-03-05
	fmt.Println("ProjectAssetPath is deprecated, use ProjectAssetsPath (with an s).")
	return ProjectAssetsPath(p)
}

// ProjectAssetsPath specify the flutter assets directory.
func ProjectAssetsPath(p string) Option {
	_, err := os.Stat(p)
	if err != nil {
		fmt.Printf("Failed to stat asset path: %v\n", err)
		os.Exit(1)
	}
	return func(c *config) {
		c.AssetsPath = p
	}
}

// ApplicationICUDataPath specify the path to the ICUData.
func ApplicationICUDataPath(p string) Option {
	_, err := os.Stat(p)
	if err != nil {
		fmt.Printf("Failed to stat icu data path: %v\n", err)
		os.Exit(1)
	}
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
		fmt.Println("Wrong initial value for width ")
		os.Exit(1)
	}

	// Check for initial application display size
	if y < 1 {
		fmt.Println("Wrong initial value for height ")
		os.Exit(1)
	}

	return func(c *config) {
		c.WindowDimension.x = x
		c.WindowDimension.y = y
	}
}

// OptionWindowInitializer allow initializing the window.
func OptionWindowInitializer(ini func(*glfw.Window) error) Option {
	// deprecated on 2019-03-05
	fmt.Println("OptionWindowInitializer is deprecated. Please read https://is.gd/gflut_window_init_deprecated")
	return func(c *config) {
		c.WindowInitializerDeprecated = ini
	}
}

// WindowIcon sets an icon provider func, which is called during window initialization.
// For tips on the kind of images to provide, see https://godoc.org/github.com/go-gl/glfw/v3.2/glfw#Window.SetIcon
func WindowIcon(iconProivder func() ([]image.Image, error)) Option {
	return func(c *config) {
		c.WindowIconProvider = iconProivder
	}
}

// OptionPixelRatio forces the the scale factor for the screen.
// By default, go-flutter will calculate the correct pixel ratio for the user, based
// on their monitor DPI. Setting this option is not advised.
func OptionPixelRatio(ratio float64) Option {
	return func(c *config) {
		c.ForcePixelRatio = ratio
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

// KeyboardShortcuts Struct where user can define his own keyboard shortcut.
// This will allow application to support keyboard layout different from US layout
type KeyboardShortcuts struct {
	Cut       glfw.Key
	Copy      glfw.Key
	Paste     glfw.Key
	SelectAll glfw.Key
}
