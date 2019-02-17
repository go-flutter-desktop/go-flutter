package flutter

import "github.com/go-gl/glfw/v3.2/glfw"

// KeyboardQwertyLayout  is the default key for shortcuts (US-layout)
var KeyboardQwertyLayout = KeyboardShortcuts{
	Cut:       glfw.KeyX,
	Copy:      glfw.KeyC,
	Paste:     glfw.KeyV,
	SelectAll: glfw.KeyA,
}

// KeyboardAzertyLayout gives an Azerty layout (french)
var KeyboardAzertyLayout = KeyboardShortcuts{
	Cut:       glfw.KeyX,
	Copy:      glfw.KeyC,
	Paste:     glfw.KeyV,
	SelectAll: glfw.KeyQ,
}
