package plugin

import "fmt"

// StreamHandler defines the interface for a stream handler.
type StreamHandler interface {
	OnListen(arguments interface{}, sink *EventSink)
	OnCancel(arguments interface{})
}

// The StreamHandlerFunc type is an adapter to allow the use of
// ordinary functions as Stream handlers.
type StreamHandlerFunc struct {
	onListen func(arguments interface{}, sink *EventSink)
	onCancel func(arguments interface{})
}

// OnListen calls f(arguments, sink).
func (f StreamHandlerFunc) OnListen(arguments interface{}, sink *EventSink) {
	f.onListen(arguments, sink)
}

// OnCancel calls f(arguments).
func (f StreamHandlerFunc) OnCancel(arguments interface{}) {
	f.onCancel(arguments)
}

// EventSink defines the interface to handle asynchronous events from Stream.
type EventSink struct {
	messenger   BinaryMessenger
	methodCodec MethodCodec
	channelName string
}

func (es *EventSink) Success(event interface{}) {
	binaryMsg, err := es.methodCodec.EncodeSuccessEnvelope(event)
	if err != nil {
		fmt.Printf("go-flutter: failed to encode success envelope for event channel '%s', error: %v\n", es.channelName, err)
	}
	es.messenger.Send(es.channelName, binaryMsg)
}

func (es *EventSink) Error(errorCode string, errorMessage string, errorDetails interface{}) {
	binaryMsg, err := es.methodCodec.EncodeErrorEnvelope(errorCode, errorMessage, errorDetails)
	if err != nil {
		fmt.Printf("go-flutter: failed to encode success envelope for event channel '%s', error: %v\n", es.channelName, err)
	}
	es.messenger.Send(es.channelName, binaryMsg)
}

func (es *EventSink) EndOfStream() {
	es.messenger.Send(es.channelName, nil)
}
