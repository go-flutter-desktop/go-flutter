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
		p.extentSelectionLineLeft()
	} else if mods.isWordTravel() {
		p.extentSelectionWordLeft()
	} else if mods.isShift() {
		p.extentSelectionCharLeft()
	} else {
		p.extentSelectionResetLeft()
	}
}

func (p *textinputPlugin) MoveCursorRight(mods keyboardShortcutsGLFW) {
	if mods.isWordTravelShift() {
		p.extentSelectionLineRight()
	} else if mods.isWordTravel() {
		p.extentSelectionWordRight()
	} else if mods.isShift() {
		p.extentSelectionCharRight()
	} else {
		p.extentSelectionResetRight()
	}
}

func (p *textinputPlugin) Backspace(mods keyboardShortcutsGLFW) {
	if p.RemoveSelectedText() {
		p.updateEditingState()
		return
	}

	if mods.isModifier() {
		p.backspaceLine()
	} else if mods.isWordTravel() {
		p.backspaceWord()
	} else {
		p.backspaceChar()
	}
}

func (p *textinputPlugin) Delete(mods keyboardShortcutsGLFW) {
	if p.RemoveSelectedText() {
		p.updateEditingState()
		return
	}

	if mods.isWordTravelShift() {
		p.deleteLine()
	} else if mods.isWordTravel() {
		p.deleteWord()
	} else {
		p.deleteChar()
	}
}
