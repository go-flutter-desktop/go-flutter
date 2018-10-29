package flutter

import "encoding/json"

// Message is the json content of a PlatformMessage
type Message struct {
	// Describe the method
	Method string `json:"method"`
	// Actual datas
	Args json.RawMessage `json:"args"`
}
