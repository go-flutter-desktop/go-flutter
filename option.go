package flutter

import (
	"fmt"
	"image"
	"os"

	"github.com/go-gl/glfw/v3.2/glfw"
)

type config struct {
	assetsPath  string
	icuDataPath string
	vmArguments []string

	windowInitializerDeprecated func(*glfw.Window) error
	windowIconProvider          func() ([]image.Image, error)
	windowInitialDimensions     windowDimensions

	forcePixelRatio float64
	keyboardLayout  KeyboardShortcuts

	// PlatformMessageReceivers    map[string][]PluginReceivers // The Key is the Channel name.
	plugins []Plugin
}

type windowDimensions struct {
	x int
	y int
}

// defaultApplicationConfig define the default configuration values for a new
// Application. These values may be changed at any time.
var defaultApplicationConfig = config{
	windowInitialDimensions: windowDimensions{
		x: 800,
		y: 600,
	},
	keyboardLayout: KeyboardQwertyLayout,
}

// Option for Application
type Option func(*config)

// ProjectAssetPath specify the flutter assets directory.
//
// Deprecated, please use ProjectAssetsPath(path).
func ProjectAssetPath(p string) Option {
	// deprecated on 2019-03-05
	fmt.Println("go-flutter: ProjectAssetPath is deprecated, use ProjectAssetsPath (with an s).")
	return ProjectAssetsPath(p)
}

// ProjectAssetsPath specify the flutter assets directory.
func ProjectAssetsPath(p string) Option {
	_, err := os.Stat(p)
	if err != nil {
		fmt.Printf("go-flutter: failed to stat asset path: %v\n", err)
		os.Exit(1)
	}
	return func(c *config) {
		c.assetsPath = p
	}
}

// ApplicationICUDataPath specify the path to the ICUData.
func ApplicationICUDataPath(p string) Option {
	_, err := os.Stat(p)
	if err != nil {
		fmt.Printf("go-flutter: failed to stat icu data path: %v\n", err)
		os.Exit(1)
	}
	return func(c *config) {
		c.icuDataPath = p
	}
}

// OptionVMArguments specify the arguments to the Dart VM.
func OptionVMArguments(a []string) Option {
	return func(c *config) {
		// First should be argument is argv[0]
		c.vmArguments = append([]string{""}, a...)
	}
}

// ApplicationWindowDimension specify the startup's dimentions of the window.
//
// Deprecated, please use WindowInitialDimensions(x, y).
func ApplicationWindowDimension(x, y int) Option {
	// deprecated on 2019-03-10
	fmt.Println("go-flutter: ApplicationWindowDimension is deprecated, use WindowInitialDimensions(x, y).")
	return WindowInitialDimensions(x, y)
}

// WindowInitialDimensions specify the startup's dimention of the window.
func WindowInitialDimensions(x, y int) Option {
	if x < 1 {
		fmt.Println("go-flutter: invalid initial value for width, must be 1 or greater.")
		os.Exit(1)
	}
	if y < 1 {
		fmt.Println("go-flutter: invalid initial value for height, must be 1 or greater.")
		os.Exit(1)
	}

	return func(c *config) {
		c.windowInitialDimensions.x = x
		c.windowInitialDimensions.y = y
	}
}

// OptionWindowInitializer allow initializing the window.
//
// Deprecated, please use WindowIcon if you'd like to set the window icon.
func OptionWindowInitializer(ini func(*glfw.Window) error) Option {
	// deprecated on 2019-03-05
	fmt.Println("go-flutter: OptionWindowInitializer is deprecated. Please read https://is.gd/gflut_window_init_deprecated")
	return func(c *config) {
		c.windowInitializerDeprecated = ini
	}
}

// WindowIcon sets an icon provider func, which is called during window
// initialization. For tips on the kind of images to provide, see
// https://godoc.org/github.com/go-gl/glfw/v3.2/glfw#Window.SetIcon
func WindowIcon(iconProivder func() ([]image.Image, error)) Option {
	return func(c *config) {
		c.windowIconProvider = iconProivder
	}
}

// OptionPixelRatio forces the the scale factor for the screen. By default,
// go-flutter will calculate the correct pixel ratio for the user, based on
// their monitor DPI. Setting this option is not advised.
//
// Deprecated, please use ForcePixelRatio(ratio).
func OptionPixelRatio(ratio float64) Option {
	// deprecated on 2019-03-10
	fmt.Println("go-flutter: OptionPixelRatio is deprecated. Please use ForcePixelRatio(ratio)")
	return ForcePixelRatio(ratio)
}

// ForcePixelRatio forces the the scale factor for the screen. By default,
// go-flutter will calculate the correct pixel ratio for the user, based on
// their monitor DPI. Setting this option is not advised.
func ForcePixelRatio(ratio float64) Option {
	return func(c *config) {
		c.forcePixelRatio = ratio
	}
}

// AddPlugin adds a plugin to the flutter application.
func AddPlugin(p Plugin) Option {
	return func(c *config) {
		c.plugins = append(c.plugins, p)
	}
}

// OptionKeyboardLayout allow application to support keyboard that have a
// different layout and therefore different keyboard shortcuts.
func OptionKeyboardLayout(keyboardLayout KeyboardShortcuts) Option {
	return func(c *config) {
		c.keyboardLayout = keyboardLayout
	}
}

// KeyboardShortcuts contains the configuration for keyboard shortcut keys. This
// allows an application to support keyboard layout different from US layout.
type KeyboardShortcuts struct {
	Cut       glfw.Key
	Copy      glfw.Key
	Paste     glfw.Key
	SelectAll glfw.Key
}
