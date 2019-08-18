package flutter

import (
	"fmt"
	"image"
	"os"

	"github.com/go-gl/glfw/v3.2/glfw"
)

type config struct {
	flutterAssetsPath string
	icuDataPath       string
	vmArguments       []string

	windowInitializerDeprecated func(*glfw.Window) error
	windowIconProvider          func() ([]image.Image, error)
	windowInitialDimensions     windowDimensions
	windowInitialLocation       windowLocation
	windowDimensionLimits       windowDimensionLimits
	windowMode                  windowMode

	forcePixelRatio float64
	keyboardLayout  KeyboardShortcuts

	plugins []Plugin
}

type windowDimensions struct {
	width  int
	height int
}

type windowLocation struct {
	xpos int
	ypos int
}

type windowDimensionLimits struct {
	minWidth  int
	minHeight int
	maxWidth  int
	maxHeight int
}

// defaultApplicationConfig define the default configuration values for a new
// Application. These values may be changed at any time.
var defaultApplicationConfig = config{
	windowInitialDimensions: windowDimensions{
		width:  800,
		height: 600,
	},
	keyboardLayout: KeyboardQwertyLayout,
	windowMode:     WindowModeDefault,
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
		fmt.Printf("go-flutter: failed to stat flutter assets path: %v\n", err)
		os.Exit(1)
	}
	return func(c *config) {
		c.flutterAssetsPath = p
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

// ApplicationWindowDimension specify the startup's dimensions of the window.
//
// Deprecated, please use WindowInitialDimensions(x, y).
func ApplicationWindowDimension(x, y int) Option {
	// deprecated on 2019-03-10, to be removed 2020-01-01
	fmt.Println("go-flutter: ApplicationWindowDimension (singular) is deprecated, use WindowInitialDimensions (plural).")
	return WindowInitialDimensions(x, y)
}

// WindowInitialDimensions specify the startup's dimension of the window.
func WindowInitialDimensions(width, height int) Option {
	if width < 1 {
		fmt.Println("go-flutter: invalid initial value for width, must be 1 or greater.")
		os.Exit(1)
	}
	if height < 1 {
		fmt.Println("go-flutter: invalid initial value for height, must be 1 or greater.")
		os.Exit(1)
	}

	return func(c *config) {
		c.windowInitialDimensions.width = width
		c.windowInitialDimensions.height = height
	}
}

// WindowInitialLocations specify the startup's position of the window.
// Location, in screen coordinates, of the upper-left corner of the client area
// of the window.
//
// Deprecated, please use WindowInitialLocation(xpos, ypos).
func WindowInitialLocations(xpos, ypos int) Option {
	// deprecated on 2019-08-18, to be removed 2020-06-01
	fmt.Println("go-flutter: WindowInitialLocations (plural) is deprecated, use WindowInitialLocation (singular).")
	return WindowInitialLocation(xpos, ypos)
}

// WindowInitialLocation specify the startup's position of the window.
// Location, in screen coordinates, of the upper-left corner of the client area
// of the window.
func WindowInitialLocation(xpos, ypos int) Option {
	if xpos < 1 {
		fmt.Println("go-flutter: invalid initial value for xpos location, must be 1 or greater.")
		os.Exit(1)
	}
	if ypos < 1 {
		fmt.Println("go-flutter: invalid initial value for ypos location, must be 1 or greater.")
		os.Exit(1)
	}

	return func(c *config) {
		c.windowInitialLocation.xpos = xpos
		c.windowInitialLocation.ypos = ypos
	}
}

// WindowDimensionLimits specify the dimension limits of the window.
// Does not work when the window is fullscreen or not resizable.
func WindowDimensionLimits(minWidth, minHeight, maxWidth, maxHeight int) Option {
	if minWidth < 1 {
		fmt.Println("go-flutter: invalid initial value for minWidth, must be 1 or greater.")
		os.Exit(1)
	}
	if minHeight < 1 {
		fmt.Println("go-flutter: invalid initial value for minHeight, must be 1 or greater.")
		os.Exit(1)
	}
	if maxWidth < minWidth {
		fmt.Println("go-flutter: invalid initial value for maxWidth, must be greater or equal to minWidth.")
		os.Exit(1)
	}
	if maxHeight < minHeight {
		fmt.Println("go-flutter: invalid initial value for maxHeight, must be greater or equal to minHeight.")
		os.Exit(1)
	}

	return func(c *config) {
		c.windowDimensionLimits.minWidth = minWidth
		c.windowDimensionLimits.minHeight = minHeight
		c.windowDimensionLimits.maxWidth = maxWidth
		c.windowDimensionLimits.maxHeight = maxHeight
	}
}

// OptionWindowInitializer allow initializing the window.
//
// Deprecated, please use WindowIcon if you'd like to set the window icon.
func OptionWindowInitializer(ini func(*glfw.Window) error) Option {
	// deprecated on 2019-03-05, to be removed 2020-01-01
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
