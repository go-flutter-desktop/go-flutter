package flutter

import "github.com/go-gl/glfw/v3.2/glfw"

func (p *keyboardShortcutsGLFW) isModifier() bool {
	return p.mod&glfw.ModControl != 0
}

func (p *keyboardShortcutsGLFW) isShift() bool {
	return p.mod&glfw.ModShift != 0
}

func (p *keyboardShortcutsGLFW) isWordTravel() bool {
	return p.mod&glfw.ModControl != 0
}

func (p *keyboardShortcutsGLFW) isWordTravelShift() bool {
	return p.mod&glfw.ModControl != 0 && p.mod&glfw.ModShift != 0
}

func (p *textinputPlugin) MoveCursorHome(mods keyboardShortcutsGLFW) {
	if mods.isShift() {
		p.moveCursorHomeSelect()
	} else {
		p.moveCursorHomeSimple()
	}
}

func (p *textinputPlugin) MoveCursorEnd(mods keyboardShortcutsGLFW) {
	if mods.isShift() {
		p.moveCursorEndSelect()
	} else {
		p.moveCursorEndSimple()
	}
}

func (p *textinputPlugin) MoveCursorLeft(mods keyboardShortcutsGLFW) {
	if mods.isWordTravelShift() {
		p.extentSelectionLeftLine()
	} else if mods.isWordTravel() {
		p.extentSelectionLeftWord()
	} else if mods.isShift() {
		p.extentSelectionLeftChar()
	} else {
		p.extentSelectionLeftReset()
	}
}

func (p *textinputPlugin) MoveCursorRight(mods keyboardShortcutsGLFW) {
	if mods.isWordTravelShift() {
		p.extentSelectionRightLine()
	} else if mods.isWordTravel() {
		p.extentSelectionRightWord()
	} else if mods.isShift() {
		p.extentSelectionRightChar()
	} else {
		p.extentSelectionRightReset()
	}
}

func (p *textinputPlugin) Backspace(mods keyboardShortcutsGLFW) {
	if p.removeSelectedText() {
		p.updateEditingState()
		return
	}

	if mods.isWordTravelShift() {
		p.sliceLeftLine()
	} else if mods.isWordTravel() {
		p.sliceLeftWord()
	} else {
		p.sliceLeftChar()
	}
}

func (p *textinputPlugin) Delete(mods keyboardShortcutsGLFW) {
	if p.removeSelectedText() {
		p.updateEditingState()
		return
	}

	if mods.isWordTravelShift() {
		p.sliceRightLine()
	} else if mods.isWordTravel() {
		p.sliceRightWord()
	} else {
		p.sliceRightChar()
	}
}
