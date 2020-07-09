package flutter

import (
	"encoding/json"
	"fmt"
	"sort"
	"unicode"
	"unicode/utf16"

	"github.com/go-flutter-desktop/go-flutter/internal/keyboard"
	"github.com/go-flutter-desktop/go-flutter/plugin"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/pkg/errors"
)

const textinputChannelName = "flutter/textinput"

// textinputPlugin implements flutter.Plugin and handles method calls to the
// flutter/textinput channel.
type textinputPlugin struct {
	channel *plugin.MethodChannel

	clientID   float64
	clientConf argSetClientConf
	ed         argsEditingState

	virtualKeyboardShow func()
	virtualKeyboardHide func()
}

// argSetClientConf is used to define the config of the TextInput. Options used:
//   The type of information for which to optimize the text input control.
//   An action the user has requested the text input control to perform.
//   Configures how the platform keyboard will select an uppercase or lowercase keyboard.
type argSetClientConf struct {
	InputType struct {
		Name string `json:"name"`
	} `json:"inputType"`
	InputAction        string `json:"inputAction"`
	TextCapitalization string `json:"textCapitalization"`
}

// argsEditingState is used to hold the current TextInput state.
type argsEditingState struct {
	Text              string `json:"text"`
	utf16Text         []uint16
	SelectionBase     int    `json:"selectionBase"`
	SelectionExtent   int    `json:"selectionExtent"`
	SelectionAffinity string `json:"selectionAffinity"`
}

// all hardcoded because theres not pluggable renderer system.
var defaultTextinputPlugin = &textinputPlugin{}

func (p *textinputPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	p.channel = plugin.NewMethodChannel(messenger, textinputChannelName, plugin.JSONMethodCodec{})
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
	// Ignored: This information is used by the flutter Web Engine
	p.channel.HandleFuncSync("TextInput.setStyle", func(_ interface{}) (interface{}, error) { return nil, nil })
	// Ignored: This information is used by the flutter Web Engine
	p.channel.HandleFuncSync("TextInput.setEditableSizeAndTransform", func(_ interface{}) (interface{}, error) { return nil, nil })
	// Ignored: This information is used by flutter on Android, iOS and web
	p.channel.HandleFuncSync("TextInput.requestAutofill", func(_ interface{}) (interface{}, error) { return nil, nil })

	return nil
}

func (p *textinputPlugin) handleSetClient(arguments interface{}) (reply interface{}, err error) {
	args := []json.RawMessage{}
	err = json.Unmarshal(arguments.(json.RawMessage), &args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode json arguments for handleSetClient")
	}

	if len(args) < 2 {
		return nil, errors.New("failed to read client args for handleSetClient")
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

func (p *textinputPlugin) handleSetEditingState(arguments interface{}) (reply interface{}, err error) {
	if p.clientID == 0 {
		return nil, errors.New("cannot set editing state when no client is selected")
	}

	err = json.Unmarshal(arguments.(json.RawMessage), &p.ed)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode json arguments for handleSetEditingState")
	}

	p.ed.utf16Text = utf16.Encode([]rune(p.ed.Text))
	utf16TextLen := len(p.ed.utf16Text)

	// sometimes flutter sends invalid cursor position
	if p.ed.SelectionBase < 0 ||
		p.ed.SelectionExtent < 0 ||
		p.ed.SelectionBase > utf16TextLen ||
		p.ed.SelectionExtent > utf16TextLen {
		// set sane default
		p.ed.SelectionBase = 0
		p.ed.SelectionExtent = 0
		// request a new EditingState if text is present in the TextInput
		if p.ed.Text != "" {
			err := p.channel.InvokeMethod("TextInputClient.requestExistingInputState", nil)
			return nil, err
		}
		return nil, nil
	}

	return nil, nil
}

func (p *textinputPlugin) glfwCharCallback(w *glfw.Window, char rune) {
	if p.clientID == 0 {
		return
	}
	// Opinionated: If a flutter dev uses TextCapitalization.characters
	//              in a TextField, that means she/he wants to receive
	//              uppercase characters.
	// TODO(Drakirus): Handle language-specific case mappings such as Turkish.
	if p.clientConf.TextCapitalization == "TextCapitalization.characters" {
		char = unicode.ToUpper(char)
	}
	p.addText(char)
}

func (p *textinputPlugin) glfwKeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

	if key == glfw.KeyEscape && action == glfw.Press {
		err := defaultNavigationPlugin.channel.InvokeMethod("popRoute", nil)
		if err != nil {
			fmt.Printf("go-flutter: failed to pop route after escape key press: %v\n", err)
		}
		return
	}

	if (action == glfw.Repeat || action == glfw.Press) && p.clientID != 0 {

		// Enter
		if key == glfw.KeyEnter || key == glfw.KeyKPEnter {
			if keyboard.DetectTextInputDoneMod(mods) {
				// Indicates that they are done typing in the TextInput
				p.performAction("TextInputAction.done")
				return
			} else if p.clientConf.InputType.Name == "TextInputType.multiline" {
				p.addText('\n')
			}
			// this action is described by argSetClientConf.
			p.performAction(p.clientConf.InputAction)
		}
		// Backspace
		if key == glfw.KeyBackspace {
			// Selection Backspace
			if p.removeSelectedText() {
				p.updateEditingState()
				return
			}
			// Word Backspace
			if keyboard.DetectWordMod(mods) {
				// Remove whitespace to the left
				for p.ed.SelectionBase != 0 && unicode.IsSpace(utf16.Decode([]uint16{p.ed.utf16Text[p.ed.SelectionBase-1]})[0]) {
					p.sliceLeftChar()
				}
				// Remove non-whitespace to the left
				for {
					if p.ed.SelectionBase == 0 || unicode.IsSpace(utf16.Decode([]uint16{p.ed.utf16Text[p.ed.SelectionBase-1]})[0]) {
						break
					}
					p.sliceLeftChar()
				}
				p.updateEditingState()
				return
			}
			// single char Backspace
			p.sliceLeftChar()
			p.updateEditingState()
		}
		// Mapping to some text navigation shortcut that are already implemented in
		// the flutter framework.
		// Home
		if key == glfw.KeyHome {
			defaultKeyeventsPlugin.sendKeyEvent(window, glfw.KeyLeft, glfw.GetKeyScancode(glfw.KeyLeft), glfw.Press, mods|glfw.ModAlt)
			defaultKeyeventsPlugin.sendKeyEvent(window, glfw.KeyLeft, glfw.GetKeyScancode(glfw.KeyLeft), glfw.Release, mods|glfw.ModAlt)
		}
		// End
		if key == glfw.KeyEnd {
			defaultKeyeventsPlugin.sendKeyEvent(window, glfw.KeyRight, glfw.GetKeyScancode(glfw.KeyRight), glfw.Press, mods|glfw.ModAlt)
			defaultKeyeventsPlugin.sendKeyEvent(window, glfw.KeyRight, glfw.GetKeyScancode(glfw.KeyRight), glfw.Release, mods|glfw.ModAlt)
		}

	}
}

func (p *textinputPlugin) addText(text rune) {
	p.removeSelectedText()
	utf16text := utf16.Encode([]rune{text})
	utf16TextLen := len(p.ed.utf16Text) + len(utf16text)
	newText := make([]uint16, 0, utf16TextLen)
	newText = append(newText, p.ed.utf16Text[:p.ed.SelectionBase]...)
	newText = append(newText, utf16text...)
	newText = append(newText, p.ed.utf16Text[p.ed.SelectionBase:]...)
	p.ed.utf16Text = newText

	p.ed.SelectionBase++
	p.ed.SelectionExtent = p.ed.SelectionBase
	p.updateEditingState()
}

// UpupdateEditingState updates the TextInput with the current state by invoking
// TextInputClient.updateEditingState in the flutter framework
func (p *textinputPlugin) updateEditingState() {
	p.ed.Text = string(utf16.Decode(p.ed.utf16Text))
	arguments := []interface{}{
		p.clientID,
		p.ed,
	}
	p.channel.InvokeMethod("TextInputClient.updateEditingState", arguments)
}

// performAction invokes the TextInputClient performAction method in the flutter
// framework
func (p *textinputPlugin) performAction(action string) {
	p.channel.InvokeMethod("TextInputClient.performAction", []interface{}{
		p.clientID,
		action,
	})
}

// performClientAction invokes the TextInputClient performAction of the
// TextInputAction. The action is described by argSetClientConf.
func (p *textinputPlugin) performTextInputAction() {
	p.performAction(p.clientConf.InputAction)
}

// removeSelectedText do nothing if no text is selected return true if the
// state needs to updated
func (p *textinputPlugin) removeSelectedText() bool {
	selectionIndexStart, selectionIndexEnd := p.getSelectedText()
	if selectionIndexStart != selectionIndexEnd {
		p.ed.utf16Text = append(p.ed.utf16Text[:selectionIndexStart], p.ed.utf16Text[selectionIndexEnd:]...)
		p.ed.SelectionBase = selectionIndexStart
		p.ed.SelectionExtent = selectionIndexStart
		return true
	}
	return false

}

// getSelectedText return a tuple containing: (left index of the selection, right index of the
// selection, the content of the selection)
func (p *textinputPlugin) getSelectedText() (int, int) {
	selectionIndex := []int{p.ed.SelectionBase, p.ed.SelectionExtent}
	sort.Ints(selectionIndex)
	return selectionIndex[0],
		selectionIndex[1]
}

func (p *textinputPlugin) sliceLeftChar() {
	if len(p.ed.utf16Text) > 0 && p.ed.SelectionBase > 0 {
		count := 1
		// Check if code point appear in a surrogate pair
		if utf16.IsSurrogate(rune(p.ed.utf16Text[p.ed.SelectionBase-1])) {
			count = 2
		}
		p.ed.utf16Text = append(p.ed.utf16Text[:p.ed.SelectionBase-count], p.ed.utf16Text[p.ed.SelectionBase:]...)
		p.ed.SelectionBase -= count
		p.ed.SelectionExtent = p.ed.SelectionBase
	}
}
