package flutter

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"unicode"

	"github.com/go-flutter-desktop/go-flutter/plugin"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/pkg/errors"
)

const textinputChannelName = "flutter/textinput"

// textinputPlugin implements flutter.Plugin and handles method calls to the
// flutter/textinput channel.
type textinputPlugin struct {
	messenger plugin.BinaryMessenger
	window    *glfw.Window
	channel   *plugin.MethodChannel

	keyboardLayout KeyboardShortcuts

	clientID        float64
	clientConf      argSetClientConf
	word            []rune
	selectionBase   int
	selectionExtent int

	virtualKeyboardShow func()
	virtualKeyboardHide func()
}

// argsEditingState is used to share text content between the framework and
// embedder
type argsEditingState struct {
	Text                   string `json:"text"`
	SelectionBase          int    `json:"selectionBase"`
	SelectionExtent        int    `json:"selectionExtent"`
	SelectionAffinity      string `json:"selectionAffinity"`
	SelectionIsDirectional bool   `json:"selectionIsDirectional"`
	ComposingBase          int    `json:"composingBase"`
	ComposingExtent        int    `json:"composingExtent"`
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

// all hardcoded because theres not pluggable renderer system.
var defaultTextinputPlugin = &textinputPlugin{}

var _ Plugin = &textinputPlugin{}     // compile-time type check
var _ PluginGLFW = &textinputPlugin{} // compile-time type check

func (p *textinputPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	p.messenger = messenger

	return nil
}

func (p *textinputPlugin) InitPluginGLFW(window *glfw.Window) error {
	p.window = window
	p.channel = plugin.NewMethodChannel(p.messenger, textinputChannelName, plugin.JSONMethodCodec{})
	p.channel.HandleFuncSync("TextInput.setClient", p.handleSetClient)
	p.channel.HandleFuncSync("TextInput.clearClient", p.handleClearClient)
	p.channel.HandleFuncSync("TextInput.setEditingState", p.handleSetEditingState)
	p.channel.HandleFunc("TextInput.show", func(_ interface{}) (interface{}, error) {
		if p.virtualKeyboardShow != nil {
			p.virtualKeyboardShow()
		}
		return nil, nil
	})
	p.channel.HandleFunc("TextInput.hide", func(_ interface{}) (interface{}, error) {
		if p.virtualKeyboardHide != nil {
			p.virtualKeyboardHide()
		}
		return nil, nil
	})
	// Ignored: This information is used by the Flutter Web Engine
	p.channel.HandleFuncSync("TextInput.setStyle", func(_ interface{}) (interface{}, error) { return nil, nil })
	// Ignored: This information is used by the Flutter Web Engine
	p.channel.HandleFuncSync("TextInput.setEditableSizeAndTransform", func(_ interface{}) (interface{}, error) { return nil, nil })
	return nil
}

func (p *textinputPlugin) handleSetClient(arguments interface{}) (reply interface{}, err error) {
	args := []json.RawMessage{}
	err = json.Unmarshal(arguments.(json.RawMessage), &args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode json arguments for handleSetClient")
	}

	err = json.Unmarshal(args[0], &p.clientID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode clientID for handleSetClient")
	}

	err = json.Unmarshal(args[1], &p.clientConf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode clientConf for handleSetClient")
	}

	return nil, nil
}

func (p *textinputPlugin) handleClearClient(arguments interface{}) (reply interface{}, err error) {
	p.clientID = 0
	return nil, nil
}

func (p *textinputPlugin) hasClient() bool {
	return p.clientID != 0
}

func (p *textinputPlugin) handleSetEditingState(arguments interface{}) (reply interface{}, err error) {
	if !p.hasClient() {
		return nil, errors.New("cannot set editing state when no client is selected")
	}

	editingState := argsEditingState{}
	err = json.Unmarshal(arguments.(json.RawMessage), &editingState)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode json arguments for handleSetEditingState")
	}

	p.word = []rune(editingState.Text)
	wordLen := len(p.word)

	// Dart/flutter string nightmare
	var errorMsg string
	if editingState.SelectionBase > wordLen || editingState.SelectionExtent > wordLen {
		errorMsg = fmt.Sprintf("invalid text selection, bounds out of range: selectionBase:%v, selectionExtent:%v, len(text): %v. Refer to go-flutter-desktop/go-flutter#221\n",
			editingState.SelectionBase, editingState.SelectionExtent, len(p.word))
	}
	if editingState.SelectionBase < 0 || editingState.SelectionExtent < 0 {
		errorMsg = fmt.Sprintf("invalid text selection: selectionBase:%v, selectionExtent:%v. Refer to go-flutter-desktop/go-flutter#332\n",
			editingState.SelectionBase, editingState.SelectionExtent)
	}
	if errorMsg != "" {
		p.selectionBase = wordLen
		p.selectionExtent = wordLen
		p.updateEditingState()
		fmt.Printf("go-flutter: recover from wrong editingState: %s", errorMsg)
		return nil, nil
	}

	p.selectionBase = editingState.SelectionBase
	p.selectionExtent = editingState.SelectionExtent
	return nil, nil
}

func (p *textinputPlugin) glfwCharCallback(w *glfw.Window, char rune) {
	if !p.hasClient() {
		return
	}
	// Opinionated: If a flutter dev uses TextCapitalization.characters
	//              in a TextField, that means he wants only to receive
	//              uppercase characters.
	// TODO(Drakirus): Handle language-specific case mappings such as Turkish.
	if p.clientConf.TextCapitalization == "TextCapitalization.characters" {
		char = unicode.ToUpper(char)
	}
	p.addChar([]rune{char})
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

func (p *textinputPlugin) addChar(char []rune) {
	newWord := make([]rune, 0, len(char)+len(p.word))
	newWord = append(newWord, p.word[:p.selectionBase]...)
	newWord = append(newWord, char...)
	newWord = append(newWord, p.word[p.selectionBase:]...)

	p.word = newWord

	p.selectionBase += len(char)
	p.selectionExtent = p.selectionBase
	p.updateEditingState()
}

func (p *textinputPlugin) glfwKeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		err := defaultNavigationPlugin.channel.InvokeMethod("popRoute", nil)
		if err != nil {
			fmt.Printf("go-flutter: failed to pop route after escape key press: %v\n", err)
		}
		return
	}

	if (action == glfw.Repeat || action == glfw.Press) && p.hasClient() {

		// Enter
		if key == glfw.KeyEnter {
			if (runtime.GOOS == "darwin" && mods == glfw.ModSuper) || (runtime.GOOS != "darwin" && mods == glfw.ModControl) {
				// Indicates that they are done typing in the TextInput
				p.performAction("TextInputAction.done")
				return
			} else if p.clientConf.InputType.Name == "TextInputType.multiline" {
				p.addChar([]rune{'\n'})
			}
			// this action is described by argSetClientConf.
			p.performAction(p.clientConf.InputAction)
		}

		// Backspace
		if key == glfw.KeyBackspace {
			if p.removeSelectedText() {
				p.updateEditingState()
				return
			}

			if (runtime.GOOS == "darwin" && mods == glfw.ModAlt) || (runtime.GOOS != "darwin" && mods == glfw.ModControl) {
				p.sliceLeftWord()
			} else {
				p.sliceLeftChar()
			}
			p.updateEditingState()

		}
	}
}

// performAction invokes the TextInputClient performAction method in the Flutter
// Framework.
func (p *textinputPlugin) performAction(action string) {
	p.channel.InvokeMethod("TextInputClient.performAction", []interface{}{
		p.clientID,
		action,
	})
}

// TEXT Utils
// For Backspace only

// TODO: watch flutter/flutter/blob/beta/packages/flutter/lib/src/rendering/editable.dart
// Backspace might get ported into the framework

// removeSelectedText do nothing if no text is selected return true if the
// state needs to updated
func (p *textinputPlugin) removeSelectedText() bool {
	selectionIndexStart, selectionIndexEnd, _ := p.getSelectedText()
	if selectionIndexStart != selectionIndexEnd {
		p.word = append(p.word[:selectionIndexStart], p.word[selectionIndexEnd:]...)
		p.selectionBase = selectionIndexStart
		p.selectionExtent = selectionIndexStart
		p.selectionExtent = p.selectionBase
		return true
	}
	return false

}

// getSelectedText return (left index of the selection, right index of the
// selection, the content of the selection)
func (p *textinputPlugin) getSelectedText() (int, int, string) {
	selectionIndex := []int{p.selectionBase, p.selectionExtent}
	sort.Ints(selectionIndex)
	return selectionIndex[0],
		selectionIndex[1],
		string(p.word[selectionIndex[0]:selectionIndex[1]])
}

func (p *textinputPlugin) sliceLeftWord() {
	if len(p.word) > 0 && p.selectionBase > 0 {
		deleteUpTo := indexStartLeadingWord(p.word, p.selectionBase)
		p.word = append(p.word[:deleteUpTo], p.word[p.selectionBase:]...)
		p.selectionBase = deleteUpTo
		p.selectionExtent = deleteUpTo

	}
}

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

func (p *textinputPlugin) sliceLeftChar() {
	if len(p.word) > 0 && p.selectionBase > 0 {
		p.word = append(p.word[:p.selectionBase-1], p.word[p.selectionBase:]...)
		p.selectionBase--
		p.selectionExtent = p.selectionBase

	}
}
