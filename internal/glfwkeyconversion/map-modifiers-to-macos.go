package glfwkeyconversion

import "github.com/go-gl/glfw/v3.3/glfw"

// Modifier key masks pulled from flutter/packages/flutter/lib/src/services/raw_keyboard_macos.dart
// URL: https://github.com/flutter/flutter/blob/3e63411256cc88afc48044aa5ea06c5c9c6a6846/packages/flutter/lib/src/services/raw_keyboard_macos.dart#L241
const (
	modifierControl    = 0x40000
	modifierShift      = 0x20000
	modifierOption     = 0x80000
	modifierCommand    = 0x100000
	modifierCapsLock   = 0x10000
	modifierNumericPad = 0x200000
)

var modifierKeytoMods = map[glfw.Key]int{
	glfw.KeyLeftControl:  modifierControl,
	glfw.KeyLeftShift:    modifierShift,
	glfw.KeyLeftAlt:      modifierOption,
	glfw.KeyLeftSuper:    modifierCommand,
	glfw.KeyRightControl: modifierControl,
	glfw.KeyRightShift:   modifierShift,
	glfw.KeyRightAlt:     modifierOption,
	glfw.KeyRightSuper:   modifierCommand,
	glfw.KeyCapsLock:     modifierCapsLock,
	glfw.KeyNumLock:      modifierNumericPad,
}

// AsMacOSModifiers translate the keycode to the ModifierKey
func AsMacOSModifiers(keycode glfw.Key) (int, bool) {
	val, ok := modifierKeytoMods[keycode]
	return val, ok
}

// ToMacOSModifiers takes a glfw ModifierKey and return his MacOS equivalent
// as defined in https://github.com/flutter/flutter/blob/3e63411256cc88afc48044aa5ea06c5c9c6a6846/packages/flutter/lib/src/services/raw_keyboard_macos.dart#L241
func ToMacOSModifiers(mods glfw.ModifierKey) int {
	macOSmods := 0
	if mods&glfw.ModControl != 0 {
		macOSmods |= modifierControl
	}
	if mods&glfw.ModShift != 0 {
		macOSmods |= modifierShift
	}
	if mods&glfw.ModAlt != 0 {
		macOSmods |= modifierOption
	}
	if mods&glfw.ModSuper != 0 {
		macOSmods |= modifierCommand
	}
	return macOSmods
}
