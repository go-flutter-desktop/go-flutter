// +build !darwin

package keyboard

import "github.com/go-gl/glfw/v3.3/glfw"

// DetectWordMod returns true if the modifiers pressed
// correspond to the word movement modifier
func DetectWordMod(mods glfw.ModifierKey) bool {
	return mods&glfw.ModControl != 0
}

// DetectTextInputDoneMod returns true if the modifiers pressed
// indicate the typed text can be committed
func DetectTextInputDoneMod(mods glfw.ModifierKey) bool {
	return mods&glfw.ModControl != 0
}

// platfromNormalize normalizes for linux and windows
func (e *Event) platfromNormalize(key glfw.Key, scancode int, mods glfw.ModifierKey) {
	e.Keymap = "linux"
	e.Toolkit = "glfw"
	e.Modifiers = int(mods)
	e.KeyCode = int(key)
	e.ScanCode = scancode
}
