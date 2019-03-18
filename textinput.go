package flutter

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"

	"github.com/go-flutter-desktop/go-flutter/plugin"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/pkg/errors"
)

const textinputChannelName = "flutter/textinput"

// textinputPlugin implements flutter.Plugin and handles method calls to the
// flutter/platform channel.
type textinputPlugin struct {
	messenger plugin.BinaryMessenger
	window    *glfw.Window
	channel   *plugin.MethodChannel

	keyboardLayout KeyboardShortcuts

	modifierKey           glfw.ModifierKey
	wordTravellerKey      glfw.ModifierKey
	wordTravellerKeyShift glfw.ModifierKey

	clientID        float64
	word            []rune
	selectionBase   int
	selectionExtent int
}

// all hardcoded because theres not pluggable renderer system.
var defaultTextinputPlugin = &textinputPlugin{}

var _ Plugin = &textinputPlugin{}     // compile-time type check
var _ PluginGLFW = &textinputPlugin{} // compile-time type check

func (p *textinputPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	p.messenger = messenger

	// set modifier keys based on OS
	switch runtime.GOOS {
	case "darwin":
		p.modifierKey = glfw.ModSuper
		p.wordTravellerKey = glfw.ModAlt
		p.wordTravellerKeyShift = glfw.ModAlt | glfw.ModShift
	default:
		p.modifierKey = glfw.ModControl
		p.wordTravellerKey = glfw.ModControl
		p.wordTravellerKeyShift = glfw.ModControl | glfw.ModShift
	}

	return nil
}

func (p *textinputPlugin) InitPluginGLFW(window *glfw.Window) error {
	p.window = window
	p.channel = plugin.NewMethodChannel(p.messenger, textinputChannelName, plugin.JSONMethodCodec{})
	p.channel.HandleFunc("TextInput.setClient", p.handleSetClient)
	p.channel.HandleFunc("TextInput.clearClient", p.handleClearClient)
	p.channel.HandleFunc("TextInput.setEditingState", p.handleSetEditingState)

	return nil
}

func (p *textinputPlugin) handleSetClient(arguments interface{}) (reply interface{}, err error) {
	var args []interface{}
	err = json.Unmarshal(arguments.(json.RawMessage), &args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode json arguments for handleSetClient")
	}
	p.clientID = args[0].(float64)
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

	editingState := argsEditingState{}
	err = json.Unmarshal(arguments.(json.RawMessage), &editingState)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode json arguments for handleSetEditingState")
	}

	p.word = []rune(editingState.Text)
	p.selectionBase = editingState.SelectionBase
	p.selectionExtent = editingState.SelectionExtent
	return nil, nil
}

func (p *textinputPlugin) glfwCharCallback(w *glfw.Window, char rune) {
	if p.clientID == 0 {
		return
	}
	p.addChar([]rune{char})
}

func (p *textinputPlugin) glfwKeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	var modsIsModfifier = false
	var modsIsShift = false
	var modsIsWordModifierShift = false
	var modsIsWordModifier = false

	switch {
	case mods == p.wordTravellerKeyShift:
		modsIsWordModifierShift = true
	case mods == p.wordTravellerKey:
		modsIsWordModifier = true
	case mods == p.modifierKey:
		modsIsModfifier = true
	case mods == glfw.ModShift:
		modsIsShift = true
	}

	if key == glfw.KeyEscape && action == glfw.Press {
		_, err := defaultNavigationPlugin.channel.InvokeMethod("popRoute", nil)
		if err != nil {
			fmt.Printf("go-flutter: failed to pop route after escape key press: %v\n", err)
		}
	}

	if action == glfw.Repeat || action == glfw.Press {
		if p.clientID == 0 {
			return
		}

		switch key {
		case glfw.KeyEnter:
			if mods == p.modifierKey {
				p.performAction("done")
			} else {
				p.addChar([]rune{'\n'})
				p.performAction("newline")
			}

		case glfw.KeyHome:
			p.MoveCursorHome(modsIsModfifier, modsIsShift, modsIsWordModifierShift, modsIsWordModifier)

		case glfw.KeyEnd:
			p.MoveCursorEnd(modsIsModfifier, modsIsShift, modsIsWordModifierShift, modsIsWordModifier)

		case glfw.KeyLeft:
			p.MoveCursorLeft(modsIsModfifier, modsIsShift, modsIsWordModifierShift, modsIsWordModifier)

		case glfw.KeyRight:
			p.MoveCursorRight(modsIsModfifier, modsIsShift, modsIsWordModifierShift, modsIsWordModifier)

		case glfw.KeyDelete:
			p.Delete(modsIsModfifier, modsIsShift, modsIsWordModifierShift, modsIsWordModifier)

		case glfw.KeyBackspace:
			p.Backspace(modsIsModfifier, modsIsShift, modsIsWordModifierShift, modsIsWordModifier)

		case p.keyboardLayout.SelectAll:
			if mods == p.modifierKey {
				p.SelectAll()
			}

		case p.keyboardLayout.Copy:
			if mods == p.modifierKey && p.isSelected() {
				_, _, selectedContent := p.GetSelectedText()
				window.SetClipboardString(selectedContent)
			}

		case p.keyboardLayout.Cut:
			if mods == p.modifierKey && p.isSelected() {
				_, _, selectedContent := p.GetSelectedText()
				window.SetClipboardString(selectedContent)
				p.RemoveSelectedText()
			}

		case p.keyboardLayout.Paste:
			if mods == p.modifierKey {
				var clpString, err = window.GetClipboardString()
				if err != nil {
					log.Printf("unable to get the clipboard content: %v\n", err)
				} else {
					p.addChar([]rune(clpString))
				}
			}
		}
	}
}
