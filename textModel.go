package main

import (
	"sort"
	"unicode"
)

type textModel struct {
	clientID        float64
	word            string
	selectionBase   int
	selectionExtent int
	notifyState     func()
}

// Modifier keys from glfw
const (
	ModNone         int = 0
	ModShift        int = 1
	ModControl      int = 2
	ModShiftControl int = 3
	ModAlt          int = 4
	ModSuper        int = 8
)

func (state *textModel) isSelected() bool {
	return state.selectionBase != state.selectionExtent
}

func (state *textModel) addChar(char rune) {
	state.removeSelectedText()
	state.word = state.word[:state.selectionBase] + string(char) + state.word[state.selectionBase:]
	state.selectionBase++
	state.selectionExtent = state.selectionBase

	state.notifyState()
}

func (state *textModel) MoveCursorHome(mods int) {
	state.selectionBase = 0
	if mods != ModShift {
		state.selectionExtent = state.selectionBase
	}
	state.notifyState()
}

func (state *textModel) MoveCursorEnd(mods int) {
	state.selectionBase = len(state.word)
	if mods != ModShift {
		state.selectionExtent = state.selectionBase
	}
	state.notifyState()
}

func (state *textModel) MoveCursorLeft(mods int) {
	if mods == ModShiftControl || mods == ModControl {
		state.selectionBase = indexStartLeadingWord([]rune(state.word), state.selectionBase)
	} else if state.selectionBase > 0 {
		state.selectionBase--
	}

	if mods == ModNone || mods == ModControl {
		state.selectionExtent = state.selectionBase
	}

	state.notifyState()
}

func (state *textModel) MoveCursorRight(mods int) {
	if mods == ModShiftControl || mods == ModControl {
		state.selectionBase = indexEndForwardWord([]rune(state.word), state.selectionBase)
	} else if state.selectionBase < len(state.word) {
		state.selectionBase++
	}
	if mods == ModNone || mods == ModControl {
		state.selectionExtent = state.selectionBase
	}

	state.notifyState()
}

func (state *textModel) Delete(mods int) {
	if state.removeSelectedText() {
		state.notifyState()
		return
	}

	if state.selectionBase < len(state.word) {
		state.word = state.word[:state.selectionBase] + state.word[state.selectionBase+1:]
		state.notifyState()
	}
}

func (state *textModel) Backspace(mods int) {
	if state.removeSelectedText() {
		state.notifyState()
		return
	}

	if state.word != "" {
		if mods == ModControl {
			deleteUpTo := indexStartLeadingWord([]rune(state.word), state.selectionBase)
			state.word = state.word[:deleteUpTo] + state.word[state.selectionBase:]
			state.selectionBase = deleteUpTo
			state.selectionExtent = deleteUpTo
			state.notifyState()
		} else {
			state.word = state.word[:state.selectionBase-1] + state.word[state.selectionBase:]
			if state.selectionBase > 0 {
				state.selectionBase--
			}
			state.selectionExtent = state.selectionBase
			state.notifyState()
		}
	}

}

// removeSelectedText do nothing if no text is selected
// return true if the state has been updated
func (state *textModel) removeSelectedText() bool {
	if state.isSelected() {
		selection := []int{state.selectionBase, state.selectionExtent}
		sort.Ints(selection)
		state.word = state.word[:selection[0]] + state.word[selection[1]:]
		state.selectionBase = selection[0]
		state.selectionExtent = selection[0]
		state.selectionExtent = state.selectionBase
		state.notifyState()
		return true
	}
	return false

}

// Helpers
func indexStartLeadingWord(line []rune, start int) int {
	pos := start
	// Remove whitespace to the left
	for {
		if pos == 0 || !unicode.IsSpace(line[pos-1]) {
			break
		}
		pos--
	}
	// Remove non-whitespace to the left
	for {
		if pos == 0 || unicode.IsSpace(line[pos-1]) {
			break
		}
		pos--
	}
	return pos
}

func indexEndForwardWord(line []rune, start int) int {
	pos := start
	lineSize := len(line) - 1
	// Remove whitespace to the left
	for {
		if pos == lineSize || !unicode.IsSpace(line[pos]) {
			break
		}
		pos++
	}
	// Remove non-whitespace to the left
	for {
		if pos == lineSize || unicode.IsSpace(line[pos]) {
			break
		}
		pos++
	}
	return pos
}
