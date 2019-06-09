package flutter

import "github.com/go-gl/glfw/v3.2/glfw"

func (p *keyboardShortcutsGLFW) isModifier() bool {
	return p.mod&glfw.ModSuper != 0
}

func (p *keyboardShortcutsGLFW) isShift() bool {
	return p.mod&glfw.ModShift != 0
}

func (p *keyboardShortcutsGLFW) isWordTravel() bool {
	return p.mod&glfw.ModAlt != 0
}

func (p *keyboardShortcutsGLFW) isWordTravelShift() bool {
	return p.mod&glfw.ModAlt != 0 && p.mod&glfw.ModShift != 0
}

func (p *textinputPlugin) MoveCursorHome(mods keyboardShortcutsGLFW) {

	if mods.isShift() {
		p.moveCursorHomeSelect()
	} else {
		p.moveCursorHomeNoSelect()
	}
}

func (p *textinputPlugin) MoveCursorEnd(mods keyboardShortcutsGLFW) {
	if mods.isShift() {
		p.moveCursorEndSelect()
	} else {
		p.moveCursorEndNoSelect()
	}
}

func (p *textinputPlugin) MoveCursorLeft(mods keyboardShortcutsGLFW) {
	if mods.isWordTravelShift() {
		p.extendSelectionLeftLine()
	} else if mods.isWordTravel() {
		p.extendSelectionLeftWord()
	} else if mods.isShift() {
		p.extendSelectionLeftChar()
	} else {
		p.extendSelectionLeftReset()
	}
}

func (p *textinputPlugin) MoveCursorRight(mods keyboardShortcutsGLFW) {
	if mods.isWordTravelShift() {
		p.extendSelectionRightLine()
	} else if mods.isWordTravel() {
		p.extendSelectionRightWord()
	} else if mods.isShift() {
		p.extendSelectionRightChar()
	} else {
		p.extendSelectionRightReset()
	}
}

func (p *textinputPlugin) Backspace(mods keyboardShortcutsGLFW) {
	if p.removeSelectedText() {
		p.updateEditingState()
		return
	}

	if mods.isModifier() {
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
