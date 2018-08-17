package main

import "encoding/json"

const (
	// channel
	platformChannel = "flutter/platform"
	textChannel     = "flutter/textinput"
	// method App description. Args -> struct AppSwitcherDescription
	setDescriptionMethod = "SystemChrome.setApplicationSwitcherDescription"

	// text
	setClientMethod       = "TextInput.setClient"
	clearClientMethod     = "TextInput.clearClient"
	setEditingStateMethod = "TextInput.setEditingState"
)

type flutterMessage struct {
	Method string          `json:"method"`
	Args   json.RawMessage `json:"args"`
}

type appSwitcherDescription struct {
	Label        string `json:"label"`
	PrimaryColor int64  `json:"primaryColor"`
}

type editingState struct {
	Text                   string `json:"text"`
	SelectionBase          int    `json:"selectionBase"`
	SelectionExtent        int    `json:"selectionExtent"`
	SelectionAffinity      string `json:"selectionAffinity"`
	SelectionIsDirectional bool   `json:"selectionIsDirectional"`
	ComposingBase          int    `json:"composingBase"`
	ComposingExtent        int    `json:"composingExtent"`
}
