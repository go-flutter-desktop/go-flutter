package plugin

import (
	"errors"
	"unicode/utf8"
)

// StringCodec implements a MessageCodec using UTF-8 encoded string messages.
type StringCodec struct{}

// Compiler test to assert that StringCodec implements MessageCodec
var _ MessageCodec = &StringCodec{}

// EncodeMessage expects message to be a string.
func (StringCodec) EncodeMessage(message interface{}) ([]byte, error) {
	if message == nil {
		return nil, nil
	}

	s, ok := message.(string)
	if !ok {
		return nil, MessageTypeError{"expected message to be of type string"}
	}
	if !utf8.ValidString(s) {
		return nil, errors.New("error encoding message to bytes, string message is not valid UTF-8 encoded")
	}
	return []byte(s), nil
}

// DecodeMessage decodes binary data into a string message.
func (StringCodec) DecodeMessage(data []byte) (message interface{}, err error) {
	if data == nil {
		return nil, nil
	}

	s := string(data)
	if !utf8.ValidString(s) {
		return nil, errors.New("error decoding bytes to message, bytes are not valid UTF-8 encoded")
	}
	return s, nil
}
