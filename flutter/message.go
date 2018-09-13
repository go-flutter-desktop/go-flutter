package flutter

import "encoding/json"

// Constant values used to read/send messages to the Flutter Engine.
const (
	// channel
	PlatformChannel  = "flutter/platform"
	TextInputChannel = "flutter/textinput"

	// Args -> struct AppSwitcherDescription
	SetDescriptionMethod = "SystemChrome.setApplicationSwitcherDescription"

	// Args -> struct ArgsEditingState
	TextUpdateStateMethod = "TextInputClient.updateEditingState"

	// text
	TextInputClientSet    = "TextInput.setClient"
	TextInputClientClear  = "TextInput.clearClient"
	TextInputSetEditState = "TextInput.setEditingState"
)

// Message is the json content of a PlatformMessage
type Message struct {
	// Describe the method
	Method string `json:"method"`
	// Actual datas
	Args json.RawMessage `json:"args"`
}

// ArgsAppSwitcherDescription Args content
type ArgsAppSwitcherDescription struct {
	Label        string `json:"label"`
	PrimaryColor int64  `json:"primaryColor"`
}

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
