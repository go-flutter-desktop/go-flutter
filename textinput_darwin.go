package flutter

import (
	"github.com/go-gl/glfw/v3.2/glfw"
)

func (p *keyboardShortcutsGLFW) isModifier() bool {
	return p.mod & glfw.ModSuper != 0
}

func (p *keyboardShortcutsGLFW) isShift() bool {
	return p.mod & glfw.ModShift != 0
}

func (p *keyboardShortcutsGLFW) isWordTravel() bool {
	return p.mod & glfw.ModAlt != 0
}

func (p *keyboardShortcutsGLFW) isWordTravelShift() bool {
	return p.mod & glfw.ModAlt != 0 && p.mod & glfw.ModShift != 0
}

func (p *textinputPlugin) MoveCursorHome(mods keyboardShortcutsGLFW) {

	if mods.isShift(){
		p.MoveCursorHomeSelect()
	} else {
		p.MoveCursorHomeSimple()
	}
}

func (p *textinputPlugin) MoveCursorEnd(mods keyboardShortcutsGLFW) {
	if mods.isShift() {
		p.MoveCursorEndSelect()
	} else {
		p.MoveCursorEndSimple()
	}
}

func (p *textinputPlugin) MoveCursorLeft(mods keyboardShortcutsGLFW) {
	if mods.isWordTravelShift() {
		p.MoveCursorLeftLine()
	} else if mods.isWordTravel() {
		p.MoveCursorLeftWord()		
	} else if mods.isShift() {
		p.MoveCursorLeftSimple()
	} else {
		p.MoveCursorLeftReset()
	}
}

func (p *textinputPlugin) MoveCursorRight(mods keyboardShortcutsGLFW) {
	if mods.isWordTravelShift() {
		p.MoveCursorRightLine()
	} else if mods.isWordTravel() {
		p.MoveCursorRightWord()		
	} else if mods.isShift() {
		p.MoveCursorRightSimple()
	} else {
		p.MoveCursorRightReset()
	}
}

func (p *textinputPlugin) Backspace(mods keyboardShortcutsGLFW) {
	if p.RemoveSelectedText() {
		p.updateEditingState()
		return
	}

	if mods.isModifier() {
		p.BackspaceLine()
	} else if mods.isWordTravel() {
		p.BackspaceWord()
	} else {
		p.BackspaceSimple()
	}
}

func (p *textinputPlugin) Delete(mods keyboardShortcutsGLFW) {
	if p.RemoveSelectedText() {
		p.updateEditingState()
		return
	}

	if mods.isWordTravelShift() {
		p.DeleteLine()
	} else if mods.isWordTravel() {
		p.DeleteWord()
	} else {
		p.DeleteSimple()
	}
}