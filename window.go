package flutter

// windowMode determines the kind of window mode to use for new windows.
type windowMode int

const (
	// WindowModeDefault is the default window mode. Windows are created with
	// borders and close/minimize buttons.
	WindowModeDefault windowMode = iota
	// WindowModeBorderless removes decorations such as borders and
	// close/minimize buttons from the window.
	WindowModeBorderless
	// WindowModeBorderlessFullscreen starts the application in borderless
	// fullscreen mode. Currently, only fullscreen on the primary monitor is
	// supported. This option overrides WindowInitialDimensions. Note that on
	// some systems a fullscreen window is very hard to close. Make sure your
	// Flutter application has a close button and use PopBehaviorIconify to
	// minimize or PopBehaviorClose to close the application.
	WindowModeBorderlessFullscreen
)

// WindowMode sets the window mode on the application.
func WindowMode(w windowMode) Option {
	return func(c *config) {
		c.windowMode = w
	}
}
