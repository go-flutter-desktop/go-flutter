package flutter

import (
	"encoding/json"

	"github.com/go-flutter-desktop/go-flutter/embedder"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// Talks to the dart side
// https://github.com/flutter/flutter/blob/master/packages/flutter/lib/src/services/system_channels.dart

////////////////////
//  Window Title  //
////////////////////

// const for `addHandlerWindowTitle`
const (
	// Channel
	platformChannel = "flutter/platform"
	// Args -> struct ArgsAppSwitcherDescription
	setDescriptionMethod = "SystemChrome.setApplicationSwitcherDescription"

	clipboardSetData = "Clipboard.setData"
	clipboardGetData = "Clipboard.getData"
)

// ArgsAppSwitcherDescription Args content
type ArgsAppSwitcherDescription struct {
	Label        string `json:"label"`
	PrimaryColor int64  `json:"primaryColor"`
}

func addHandlerWindowTitle() Option {
	var handler PluginReceivers = func(
		platMessage *embedder.PlatformMessage,
		flutterEngine *embedder.FlutterEngine,
		window *glfw.Window,
	) bool {
		message := &platMessage.Message

		if message.Method == setDescriptionMethod {
			msgBody := ArgsAppSwitcherDescription{}
			json.Unmarshal(message.Args, &msgBody)
			window.SetTitle(msgBody.Label)
			return true
		}
		return false
	}

	return OptionAddPluginReceiver(handler, platformChannel)
}

func addHandlerClipboard() Option {
	handler := func(platMessage *embedder.PlatformMessage,
		flutterEngine *embedder.FlutterEngine,
		window *glfw.Window) bool {

		message := &platMessage.Message
		switch message.Method {
		case clipboardSetData:
			newClipboard := struct {
				Text string `json:"text"`
			}{}
			json.Unmarshal(message.Args, &newClipboard)
			window.SetClipboardString(newClipboard.Text)
		case clipboardGetData:
			requestedMime := ""
			json.Unmarshal(message.Args, &requestedMime)
			if requestedMime == "text/plain" {
				clipText, _ := window.GetClipboardString()

				retBytes, _ := json.Marshal([]struct {
					Text string `json:"text"`
				}{{clipText}})

				flutterEngine.SendPlatformMessageResponse(platMessage, retBytes)
				return true
			} else {
				// log.Printf("Don't know how to acquire type #v from the clipboard", requestedMime)
			}

		default:
			// log.Printf("unhandled platform method: %#v\n", platMessage.Message)
		}
		return false

	}
	return OptionAddPluginReceiver(handler, platformChannel)
}

/////////////////
//  TextInput  //
/////////////////

// const for `addHandlerTextInput`
const (
	// channel
	textInputChannel = "flutter/textinput"

	// Args -> struct argsEditingState
	textUpdateStateMethod = "TextInputClient.updateEditingState"

	// text
	textInputClientSet    = "TextInput.setClient"
	textInputClientClear  = "TextInput.clearClient"
	textInputSetEditState = "TextInput.setEditingState"
)

// argsEditingState Args content
// To update the embedder text use `flutter.SendPlatformMessage` whenever a keys is pressed
type argsEditingState struct {
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
		platMessage *embedder.PlatformMessage,
		flutterEngine *embedder.FlutterEngine,
		window *glfw.Window,
	) bool {

		message := &platMessage.Message

		switch message.Method {
		case textInputClientClear:
			state.clientID = 0
		case textInputClientSet:
			var body []interface{}
			json.Unmarshal(message.Args, &body)
			state.clientID = body[0].(float64)
		case textInputSetEditState:
			if state.clientID != 0 {
				editingState := argsEditingState{}
				json.Unmarshal(message.Args, &editingState)
				state.word = []rune(editingState.Text)
				state.selectionBase = editingState.SelectionBase
				state.selectionExtent = editingState.SelectionExtent
			}
		default:
		}

		return true

	}

	return OptionAddPluginReceiver(handler, textInputChannel)
}
