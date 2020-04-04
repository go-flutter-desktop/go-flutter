package glfwkeyconversion

import "github.com/go-gl/glfw/v3.3/glfw"

const (
	modifierControl    = 0x40000
	modifierShift      = 0x20000
	modifierOption     = 0x80000
	modifierCommand    = 0x100000
	modifierCapsLock   = 0x10000
	modifierNumericPad = 0x200000
)

var modifierKeyNames = []string{"controlLeft", "shiftLeft", "altLeft", "metaLeft", "controlRight", "shiftRight", "altRight", "metaRight", "capsLock", "numLock"}
var modifierKeytoMods = map[int]int{
	341: modifierControl,
	340: modifierShift,
	342: modifierOption,
	343: modifierCommand,
	345: modifierControl,
	344: modifierShift,
	346: modifierOption,
	347: modifierCommand,
	280: modifierCapsLock,
	282: modifierNumericPad,
}

// IsModiferKeycode returns true if the keycode is from a ModifierKey
func IsModiferKeycode(keycode glfw.Key) bool {
	// GLFW Key codes for modifier keys.
	if val, ok := glfwToLogicalKey[int(keycode)]; ok {
		for _, modCode := range modifierKeyNames {
			if modCode == val {
				return true
			}
		}
	}
	return false
}

// AsMacOSModifiers translate the keycode to the ModifierKey
func AsMacOSModifiers(keycode glfw.Key) int {
	if val, ok := modifierKeytoMods[int(keycode)]; ok {
		return val
	}
	return 0
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
