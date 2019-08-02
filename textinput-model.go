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
	p.removeSelectedText()
	newWord := make([]rune, 0, len(char)+len(p.word))
	newWord = append(newWord, p.word[:p.selectionBase]...)
	newWord = append(newWord, char...)
	newWord = append(newWord, p.word[p.selectionBase:]...)

	p.word = newWord

	p.selectionBase += len(char)
	p.selectionExtent = p.selectionBase
	p.updateEditingState()
}

func (p *textinputPlugin) moveCursorHomeNoSelect() {
	p.selectionBase = 0
	p.selectionExtent = p.selectionBase
}

func (p *textinputPlugin) moveCursorHomeSelect() {
	p.selectionBase = 0
}

func (p *textinputPlugin) moveCursorEndNoSelect() {
	p.selectionBase = len(p.word)
	p.selectionExtent = p.selectionBase
}

func (p *textinputPlugin) moveCursorEndSelect() {
	p.selectionBase = len(p.word)
}

func (p *textinputPlugin) extendSelectionLeftChar() {
	if p.selectionExtent > 0 {
		p.selectionExtent--
	}
}

func (p *textinputPlugin) extendSelectionLeftWord() {
	p.selectionBase = indexStartLeadingWord(p.word, p.selectionBase)
	p.selectionExtent = p.selectionBase

}

func (p *textinputPlugin) extendSelectionLeftLine() {
	if p.isSelected() {
		p.selectionExtent = indexStartLeadingWord(p.word, p.selectionExtent)
	} else {
		p.selectionExtent = indexStartLeadingWord(p.word, p.selectionBase)
	}

}

func (p *textinputPlugin) extendSelectionLeftReset() {
	if !p.isSelected() {
		if p.selectionBase > 0 {
			p.selectionBase--
			p.selectionExtent = p.selectionBase
		}
	} else {
		p.selectionBase = p.selectionExtent
	}
}

func (p *textinputPlugin) extendSelectionRightChar() {
	if p.selectionExtent < len(p.word) {
		p.selectionExtent++
	}
}

func (p *textinputPlugin) extendSelectionRightWord() {
	p.selectionBase = indexEndForwardWord(p.word, p.selectionBase)
	p.selectionExtent = p.selectionBase

}

func (p *textinputPlugin) extendSelectionRightLine() {
	if p.isSelected() {
		p.selectionExtent = indexEndForwardWord(p.word, p.selectionExtent)
	} else {
		p.selectionExtent = indexEndForwardWord(p.word, p.selectionBase)
	}

}

func (p *textinputPlugin) extendSelectionRightReset() {
	if !p.isSelected() {
		if p.selectionBase < len(p.word) {
			p.selectionBase++
			p.selectionExtent = p.selectionBase
		}
	} else {
		p.selectionBase = p.selectionExtent
	}
}

func (p *textinputPlugin) selectAll() {
	p.selectionBase = 0
	p.selectionExtent = len(p.word)
}

func (p *textinputPlugin) sliceRightChar() {
	if p.selectionBase < len(p.word) {
		p.word = append(p.word[:p.selectionBase], p.word[p.selectionBase+1:]...)

	}
}

func (p *textinputPlugin) sliceRightWord() {
	UpTo := indexEndForwardWord(p.word, p.selectionBase)
	p.word = append(p.word[:p.selectionBase], p.word[UpTo:]...)
}

func (p *textinputPlugin) sliceRightLine() {
	p.word = p.word[:p.selectionBase]
}

func (p *textinputPlugin) sliceLeftChar() {
	if len(p.word) > 0 && p.selectionBase > 0 {
		p.word = append(p.word[:p.selectionBase-1], p.word[p.selectionBase:]...)
		p.selectionBase--
		p.selectionExtent = p.selectionBase

	}
}

func (p *textinputPlugin) sliceLeftWord() {
	if len(p.word) > 0 && p.selectionBase > 0 {
		deleteUpTo := indexStartLeadingWord(p.word, p.selectionBase)
		p.word = append(p.word[:deleteUpTo], p.word[p.selectionBase:]...)
		p.selectionBase = deleteUpTo
		p.selectionExtent = deleteUpTo

	}
}

func (p *textinputPlugin) sliceLeftLine() {
	p.word = p.word[p.selectionBase:]
	p.selectionBase = 0
	p.selectionExtent = 0
}

// removeSelectedText do nothing if no text is selected
// return true if the state has been updated
func (p *textinputPlugin) removeSelectedText() bool {
	if p.isSelected() {
		selectionIndexStart, selectionIndexEnd, _ := p.getSelectedText()
		p.word = append(p.word[:selectionIndexStart], p.word[selectionIndexEnd:]...)
		p.selectionBase = selectionIndexStart
		p.selectionExtent = selectionIndexStart
		p.selectionExtent = p.selectionBase

		return true
	}
	return false

}

// getSelectedText return
// (left index of the selection, right index of the selection,
// the content of the selection)
func (p *textinputPlugin) getSelectedText() (int, int, string) {
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
		action,
	})
}

// performClientAction invokes the TextInputClient performAction of the
// TextInputAction. The action is described by argSetClientConf.
func (p *textinputPlugin) performTextInputAction() {
	p.channel.InvokeMethod("TextInputClient.performAction", []interface{}{
		p.clientID,
		p.clientConf.InputAction,
	})
}

// argSetClientConf is used to define the config of the TextInput. Options used:
//   The type of information for which to optimize the text input control.
//   https://docs.flutter.io/flutter/services/TextInputType-class.html

//   An action the user has requested the text input control to perform.
//   https://docs.flutter.io/flutter/services/TextInputAction-class.html

//   Configures how the platform keyboard will select an uppercase or lowercase keyboard.
//   https://api.flutter.dev/flutter/services/TextCapitalization-class.html
type argSetClientConf struct {
	InputType struct {
		Name string `json:"name"`
	} `json:"inputType"`
	InputAction        string `json:"inputAction"`
	TextCapitalization string `json:"textCapitalization"`
}
