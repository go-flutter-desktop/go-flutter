package plugin

import "fmt"

// FlutterError is returned to indicate that a Flutter method invocation
// failed on the Flutter side.
type FlutterError struct {
	Code    string
	Message string
	Details interface{}
}

// Error returns a string describing the FlutterError
func (e FlutterError) Error() string {
	return fmt.Sprintf("Error %s in Flutter: %s (%v)", e.Code, e.Message, e.Details)
}

// MessageTypeError is returned when a MessageCodec implementation is asked to
// encode a message of the wrong type.
type MessageTypeError struct {
	hint string
}

// Error returns a string describing the MessageTypeError
func (e MessageTypeError) Error() string {
	return fmt.Sprintf("invalid type provided to message codec: %s", e.hint)
}
