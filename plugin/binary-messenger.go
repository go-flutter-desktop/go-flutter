package plugin

// BinaryMessenger defines a bidirectional binary messenger.
type BinaryMessenger interface {
	// Send sends a binary message to the Flutter application.
	Send(channel string, binaryMessage []byte) (binaryReply []byte, err error)

	// SetChannelHandler registers a handler to be invoked when the Flutter
	// application sends a message to its host platform on given channel.
	//
	// Registration overwrites any previous registration for the same channel
	// name. Use nil as handler to deregister.
	SetChannelHandler(channel string, handler ChannelHandlerFunc)
}

type ChannelHandlerFunc func(binaryMessage []byte) (binaryReply []byte, err error)
