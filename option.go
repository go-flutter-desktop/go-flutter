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

	windowIconProvider      func() ([]image.Image, error)
	windowInitialDimensions windowDimensions
	windowInitialLocation   windowLocation
	windowDimensionLimits   windowDimensionLimits
	windowMode              windowMode
	windowAlwaysOnTop       bool

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
	keyboardLayout:    KeyboardQwertyLayout,
	windowMode:        WindowModeDefault,
	windowAlwaysOnTop: false,
}

// Option for Application
type Option func(*config)

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

// WindowIcon sets an icon provider func, which is called during window
// initialization. For tips on the kind of images to provide, see
// https://godoc.org/github.com/go-gl/glfw/v3.2/glfw#Window.SetIcon
func WindowIcon(iconProivder func() ([]image.Image, error)) Option {
	return func(c *config) {
		c.windowIconProvider = iconProivder
	}
}

// ForcePixelRatio forces the the scale factor for the screen. By default,
// go-flutter will calculate the correct pixel ratio for the user, based on
// their monitor DPI. Setting this option is not advised.
func ForcePixelRatio(ratio float64) Option {
	return func(c *config) {
		c.forcePixelRatio = ratio
	}
}

// WindowAlwaysOnTop sets the application window to be always on top of other windows
func WindowAlwaysOnTop(enabled bool) Option {
	return func(c *config) {
		c.windowAlwaysOnTop = enabled
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

// VirtualKeyboardShow sets an func called when the flutter framework want to
// show the keyboard.
// This Option is interesting for people wanting to display the on-screen
// keyboard on TextField focus.
// It's up to the flutter developer to implement (or not) this function with
// the OS related call.
func VirtualKeyboardShow(showCallback func()) Option {
	return func(c *config) {
		// Reference the callback to the platform plugin (singleton) responsible
		// for textinput.
		defaultTextinputPlugin.virtualKeyboardShow = showCallback
	}
}

// VirtualKeyboardHide sets an func called when the flutter framework want to
// hide the keyboard.
func VirtualKeyboardHide(hideCallback func()) Option {
	return func(c *config) {
		// Reference the callback to the platform plugin (singleton) responsible
		// for textinput.
		defaultTextinputPlugin.virtualKeyboardHide = hideCallback
	}
}
