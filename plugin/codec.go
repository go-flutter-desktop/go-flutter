package plugin

// MessageCodec defines a message encoding/decoding mechanism.
type MessageCodec interface {
	// EncodeMessage encodes a message to a slice of bytes.
	EncodeMessage(message interface{}) (binaryMessage []byte, err error)
	// DecodeMessage decodes a slice of bytes to a message.
	DecodeMessage(binaryMessage []byte) (message interface{}, err error)
}

// MethodCall describes a method invocation.
type MethodCall struct {
	Method    string
	Arguments interface{}
}

// MethodCodec describes a codec for method calls and enveloped results.
type MethodCodec interface {
	// EncodeMethodCall encodes the MethodCall into binary
	// Returns an error on invalid MethodCall arguments.
	EncodeMethodCall(methodCall MethodCall) (data []byte, err error)

	// DecodeMethodCal decodes the MethodCall from binary.
	// Returns an error on invalid data.
	DecodeMethodCall(data []byte) (methodCall MethodCall, err error)

	/// Encodes a successful [result] into a binary envelope.
	EncodeSuccessEnvelope(result interface{}) (data []byte, err error)

	// EncodeErrorEnvelope encodes an error result into a binary envelope.
	// The specified error code, human-readable error message, and error
	// details correspond to the fields of Flutter's PlatformException.
	EncodeErrorEnvelope(code string, message string, details interface{}) (data []byte, err error)

	// DecodeEnvelope decodes the specified result [envelope] from binary.
	// Returns a FlutterError as error if provided envelope represents an error,
	// otherwise returns the enveloped result.
	DecodeEnvelope(envelope []byte) (result interface{}, err error)
}
