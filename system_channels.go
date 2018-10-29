package gutter

import (
	"encoding/json"

	"github.com/Drakirus/go-flutter-desktop-embedder/flutter"
	"github.com/go-gl/glfw/v3.2/glfw"
)

////////////////////
//  Window Title  //
////////////////////

// const for `addHandlerWindowTitle`
const (
	// Args -> struct ArgsAppSwitcherDescription
	setDescriptionMethod = "SystemChrome.setApplicationSwitcherDescription"
)

// ArgsAppSwitcherDescription Args content
type ArgsAppSwitcherDescription struct {
	Label        string `json:"label"`
	PrimaryColor int64  `json:"primaryColor"`
}

func addHandlerWindowTitle() Option {

	var handler PluginReceivers = func(
		platMessage flutter.PlatformMessage,
		flutterEngine *flutter.EngineOpenGL,
		window *glfw.Window,
	) bool {
		message := platMessage.Message

		if message.Method == setDescriptionMethod {
			msgBody := ArgsAppSwitcherDescription{}
			json.Unmarshal(message.Args, &msgBody)
			window.SetTitle(msgBody.Label)
			return true
		}
		return false
	}

	return OptionAddPluginReceiver(handler)

}

/////////////////
//  TextInput  //
/////////////////

// const for `addHandlerTextInput`
const (
	// channel
	PlatformChannel  = "flutter/platform"
	TextInputChannel = "flutter/textinput"

	// Args -> struct ArgsEditingState
	TextUpdateStateMethod = "TextInputClient.updateEditingState"

	// text
	TextInputClientSet    = "TextInput.setClient"
	TextInputClientClear  = "TextInput.clearClient"
	TextInputSetEditState = "TextInput.setEditingState"
)

// ArgsEditingState Args content
type ArgsEditingState struct {
	Text                   string `json:"text"`
	SelectionBase          int    `json:"selectionBase"`
	SelectionExtent        int    `json:"selectionExtent"`
	SelectionAffinity      string `json:"selectionAffinity"`
	SelectionIsDirectional bool   `json:"selectionIsDirectional"`
	ComposingBase          int    `json:"composingBase"`
	ComposingExtent        int    `json:"composingExtent"`
}

func addHandlerTextInput() Option {

	var handler PluginReceivers = func(
		platMessage flutter.PlatformMessage,
		flutterEngine *flutter.EngineOpenGL,
		window *glfw.Window,
	) bool {

		message := platMessage.Message

		if platMessage.Channel == TextInputChannel {
			switch message.Method {
			case TextInputClientClear:
				state.clientID = 0
			case TextInputClientSet:
				var body []interface{}
				json.Unmarshal(message.Args, &body)
				state.clientID = body[0].(float64)
			case TextInputSetEditState:
				if state.clientID != 0 {
					editingState := ArgsEditingState{}
					json.Unmarshal(message.Args, &editingState)
					state.word = editingState.Text
					state.selectionBase = editingState.SelectionBase
					state.selectionExtent = editingState.SelectionExtent
				}
			default:
			}
		}

		return true

	}

	return OptionAddPluginReceiver(handler)

}
