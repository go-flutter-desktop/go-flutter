// +build !darwin

package keyboard

import "github.com/go-gl/glfw/v3.3/glfw"

// platfromNormalize normalizes for linux and windows
func (e *Event) platfromNormalize(key glfw.Key, scancode int, mods glfw.ModifierKey) {
	e.Keymap = "linux"
	e.Toolkit = "glfw"
	e.Modifiers = int(mods)
	e.KeyCode = int(key)
	e.ScanCode = scancode
}
