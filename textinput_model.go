package flutter

import (
	"sort"
	"unicode"
)

type argsEditingState struct {
	Text                   string `json:"text"`
	SelectionBase          int    `json:"selectionBase"`
	SelectionExtent        int    `json:"selectionExtent"`
	SelectionAffinity      string `json:"selectionAffinity"`
	SelectionIsDirectional bool   `json:"selectionIsDirectional"`
	ComposingBase          int    `json:"composingBase"`
	ComposingExtent        int    `json:"composingExtent"`
}

func (p *textinputPlugin) isSelected() bool {
	return p.selectionBase != p.selectionExtent
}

func (p *textinputPlugin) addChar(char []rune) {
	p.RemoveSelectedText()
	newWord := make([]rune, 0, len(char)+len(p.word))
	newWord = append(newWord, p.word[:p.selectionBase]...)
	newWord = append(newWord, char...)
	newWord = append(newWord, p.word[p.selectionBase:]...)

	p.word = newWord

	p.selectionBase += len(char)
	p.selectionExtent = p.selectionBase
	p.updateEditingState()
}

func (p *textinputPlugin) MoveCursorHome(modsIsModifier bool, modsIsShift bool, modsIsWordModifierShift bool, modsIsWordModifier bool) {
	p.selectionBase = 0
	if !modsIsShift {
		p.selectionExtent = p.selectionBase
	}
	p.updateEditingState()
}

func (p *textinputPlugin) MoveCursorEnd(modsIsModifier bool, modsIsShift bool, modsIsWordModifierShift bool, modsIsWordModifier bool) {
	p.selectionBase = len(p.word)
	if !modsIsShift {
		p.selectionExtent = p.selectionBase
	}
	p.updateEditingState()
}

func (p *textinputPlugin) MoveCursorLeft(modsIsModifier bool, modsIsShift bool, modsIsWordModifierShift bool, modsIsWordModifier bool) {
	if modsIsWordModifierShift {
		if p.isSelected() {
			p.selectionExtent = indexStartLeadingWord(p.word, p.selectionExtent)
		} else {
			p.selectionExtent = indexStartLeadingWord(p.word, p.selectionBase)
		}
	} else if modsIsWordModifier {
		p.selectionBase = indexStartLeadingWord(p.word, p.selectionBase)
		p.selectionExtent = p.selectionBase
	} else if modsIsShift {
		p.selectionExtent--
	} else if !p.isSelected() {
		if p.selectionBase > 0 {
			p.selectionBase--
			p.selectionExtent = p.selectionBase
		}
	} else {
		p.selectionBase = p.selectionExtent
	}
	p.updateEditingState()
}

func (p *textinputPlugin) MoveCursorRight(modsIsModifier bool, modsIsShift bool, modsIsWordModifierShift bool, modsIsWordModifier bool) {
	if modsIsWordModifierShift {
		if p.isSelected() {
			p.selectionExtent = indexEndForwardWord(p.word, p.selectionExtent)
		} else {
			p.selectionExtent = indexEndForwardWord(p.word, p.selectionBase)
		}
	} else if modsIsWordModifier {
		p.selectionBase = indexEndForwardWord(p.word, p.selectionBase)
		p.selectionExtent = p.selectionBase
	} else if modsIsShift {
		p.selectionExtent++
	} else if !p.isSelected() {
		if p.selectionBase < len(p.word) {
			p.selectionBase++
			p.selectionExtent = p.selectionBase
		}
	} else {
		p.selectionBase = p.selectionExtent
	}

	p.updateEditingState()

}

func (p *textinputPlugin) SelectAll() {
	p.selectionBase = 0
	p.selectionExtent = len(p.word)
	p.updateEditingState()
}

func (p *textinputPlugin) Delete(modsIsModifier bool, modsIsShift bool, modsIsWordModifierShift bool, modsIsWordModifier bool) {
	if p.RemoveSelectedText() {
		p.updateEditingState()
		return
	}

	if p.selectionBase < len(p.word) {
		p.word = append(p.word[:p.selectionBase], p.word[p.selectionBase+1:]...)
		p.updateEditingState()
	}
}

func (p *textinputPlugin) Backspace(modsIsModifier bool, modsIsShift bool, modsIsWordModifierShift bool, modsIsWordModifier bool) {
	if p.RemoveSelectedText() {
		p.updateEditingState()
		return
	}

	if len(p.word) > 0 && p.selectionBase > 0 {
		if modsIsWordModifier {
			deleteUpTo := indexStartLeadingWord(p.word, p.selectionBase)
			p.word = append(p.word[:deleteUpTo], p.word[p.selectionBase:]...)
			p.selectionBase = deleteUpTo
			p.selectionExtent = deleteUpTo
			p.updateEditingState()
		} else {
			p.word = append(p.word[:p.selectionBase-1], p.word[p.selectionBase:]...)
			if p.selectionBase > 0 {
				p.selectionBase--
			}
			p.selectionExtent = p.selectionBase
			p.updateEditingState()
		}
	}

}

// RemoveSelectedText do nothing if no text is selected
// return true if the state has been updated
func (p *textinputPlugin) RemoveSelectedText() bool {
	if p.isSelected() {
		selectionIndexStart, selectionIndexEnd, _ := p.GetSelectedText()
		p.word = append(p.word[:selectionIndexStart], p.word[selectionIndexEnd:]...)
		p.selectionBase = selectionIndexStart
		p.selectionExtent = selectionIndexStart
		p.selectionExtent = p.selectionBase
		p.updateEditingState()
		return true
	}
	return false

}

// GetSelectedText return
// (left index of the selection, right index of the selection,
// the content of the selection)
func (p *textinputPlugin) GetSelectedText() (int, int, string) {
	selectionIndex := []int{p.selectionBase, p.selectionExtent}
	sort.Ints(selectionIndex)
	return selectionIndex[0],
		selectionIndex[1],
		string(p.word[selectionIndex[0]:selectionIndex[1]])
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

// UpupdateEditingState updates the TextInput with the current state by invoking
// TextInputClient.updateEditingState in the Flutter Framework.
func (p *textinputPlugin) updateEditingState() {
	editingState := argsEditingState{
		Text:                   string(p.word),
		SelectionAffinity:      "TextAffinity.downstream",
		SelectionBase:          p.selectionBase,
		SelectionExtent:        p.selectionExtent,
		SelectionIsDirectional: false,
	}
	arguments := []interface{}{
		p.clientID,
		editingState,
	}
	p.channel.InvokeMethod("TextInputClient.updateEditingState", arguments)
}

// performAction invokes the TextInputClient performAction method in the Flutter
// Framework.
func (p *textinputPlugin) performAction(action string) {
	p.channel.InvokeMethod("TextInputClient.performAction", []interface{}{
		p.clientID,
		"TextInputAction." + action,
	})
}
