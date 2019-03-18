package plugin

// BinaryCodec implements a MessageCodec using unencoded binary messages, represented as byte slices.
type BinaryCodec struct{}

// Compiler test to assert that BinaryCodec implements MessageCodec
var _ MessageCodec = &BinaryCodec{}

// EncodeMessage expects message to be a slice of bytes.
func (BinaryCodec) EncodeMessage(message interface{}) ([]byte, error) {
	if message == nil {
		return nil, nil
	}

	b, ok := message.([]byte)
	if !ok {
		return nil, MessageTypeError{"expected message to be of type []byte"}
	}
	return b, nil
}

// DecodeMessage decodes binary data into binary data.
func (BinaryCodec) DecodeMessage(data []byte) (message interface{}, err error) {
	if data == nil {
		return nil, nil
	}

	return data, nil
}
