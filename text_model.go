package flutter

import (
	"sort"
	"unicode"
)

type textModel struct {
	clientID        float64
	word            []rune
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
	ModShiftAlt     int = 5
	ModSuper        int = 8
	ModShiftSuper   int = 9
)

func (state *textModel) isSelected() bool {
	return state.selectionBase != state.selectionExtent
}

func (state *textModel) addChar(char []rune) {
	state.RemoveSelectedText()
	newWord := make([]rune, 0, len(char)+len(state.word))
	newWord = append(newWord, state.word[:state.selectionBase]...)
	newWord = append(newWord, char...)
	newWord = append(newWord, state.word[state.selectionBase:]...)

	state.word = newWord

	state.selectionBase += len(char)
	state.selectionExtent = state.selectionBase
	state.notifyState()
}

func (state *textModel) MoveCursorHome(modsIsModifier bool, modsIsShift bool, modsIsWordModifierShift bool, modsIsWordModifier bool) {
	state.selectionBase = 0
	if !modsIsShift {
		state.selectionExtent = state.selectionBase
	}
	state.notifyState()
}

func (state *textModel) MoveCursorEnd(modsIsModifier bool, modsIsShift bool, modsIsWordModifierShift bool, modsIsWordModifier bool) {
	state.selectionBase = len(state.word)
	if !modsIsShift {
		state.selectionExtent = state.selectionBase
	}
	state.notifyState()
}

func (state *textModel) MoveCursorLeft(modsIsModifier bool, modsIsShift bool, modsIsWordModifierShift bool, modsIsWordModifier bool) {
	if modsIsWordModifierShift {
		if state.isSelected() {
			state.selectionExtent = indexStartLeadingWord(state.word, state.selectionExtent)
		} else {
			state.selectionExtent = indexStartLeadingWord(state.word, state.selectionBase)
		}
	} else if modsIsWordModifier {
		state.selectionBase = indexStartLeadingWord(state.word, state.selectionBase)
		state.selectionExtent = state.selectionBase
	} else if modsIsShift {
		state.selectionExtent--
	} else if !state.isSelected() {
		if state.selectionBase > 0 {
			state.selectionBase--
			state.selectionExtent = state.selectionBase
		}
	} else {
		state.selectionBase = state.selectionExtent
	}
	state.notifyState()
}

func (state *textModel) MoveCursorRight(modsIsModifier bool, modsIsShift bool, modsIsWordModifierShift bool, modsIsWordModifier bool) {
	if modsIsWordModifierShift {
		if state.isSelected() {
			state.selectionExtent = indexEndForwardWord(state.word, state.selectionExtent)
		} else {
			state.selectionExtent = indexEndForwardWord(state.word, state.selectionBase)
		}
	} else if modsIsWordModifier {
		state.selectionBase = indexEndForwardWord(state.word, state.selectionBase)
		state.selectionExtent = state.selectionBase
	} else if modsIsShift {
		state.selectionExtent++
	} else if !state.isSelected() {
		if state.selectionBase < len(state.word) {
			state.selectionBase++
			state.selectionExtent = state.selectionBase
		}
	} else {
		state.selectionBase = state.selectionExtent
	}

	state.notifyState()

}

func (state *textModel) SelectAll() {
	state.selectionBase = 0
	state.selectionExtent = len(state.word)
	state.notifyState()
}

func (state *textModel) Delete(modsIsModifier bool, modsIsShift bool, modsIsWordModifierShift bool, modsIsWordModifier bool) {
	if state.RemoveSelectedText() {
		state.notifyState()
		return
	}

	if state.selectionBase < len(state.word) {
		state.word = append(state.word[:state.selectionBase], state.word[state.selectionBase+1:]...)
		state.notifyState()
	}
}

func (state *textModel) Backspace(modsIsModifier bool, modsIsShift bool, modsIsWordModifierShift bool, modsIsWordModifier bool) {
	if state.RemoveSelectedText() {
		state.notifyState()
		return
	}

	if len(state.word) > 0 && state.selectionBase > 0 {
		if modsIsWordModifier {
			deleteUpTo := indexStartLeadingWord(state.word, state.selectionBase)
			state.word = append(state.word[:deleteUpTo], state.word[state.selectionBase:]...)
			state.selectionBase = deleteUpTo
			state.selectionExtent = deleteUpTo
			state.notifyState()
		} else {
			state.word = append(state.word[:state.selectionBase-1], state.word[state.selectionBase:]...)
			if state.selectionBase > 0 {
				state.selectionBase--
			}
			state.selectionExtent = state.selectionBase
			state.notifyState()
		}
	}

}

// RemoveSelectedText do nothing if no text is selected
// return true if the state has been updated
func (state *textModel) RemoveSelectedText() bool {
	if state.isSelected() {
		selectionIndexStart, selectionIndexEnd, _ := state.GetSelectedText()
		state.word = append(state.word[:selectionIndexStart], state.word[selectionIndexEnd:]...)
		state.selectionBase = selectionIndexStart
		state.selectionExtent = selectionIndexStart
		state.selectionExtent = state.selectionBase
		state.notifyState()
		return true
	}
	return false

}

// GetSelectedText return
// (left index of the selection, right index of the selection,
// the content of the selection)
func (state *textModel) GetSelectedText() (int, int, string) {
	selectionIndex := []int{state.selectionBase, state.selectionExtent}
	sort.Ints(selectionIndex)
	return selectionIndex[0],
		selectionIndex[1],
		string(state.word[selectionIndex[0]:selectionIndex[1]])
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
	lineSize := len(line)
	// Remove whitespace to the right
	for {
		if pos == lineSize || !unicode.IsSpace(line[pos]) {
			break
		}
		pos++
	}
	// Remove non-whitespace to the right
	for {
		if pos == lineSize || unicode.IsSpace(line[pos]) {
			break
		}
		pos++
	}
	return pos
}
