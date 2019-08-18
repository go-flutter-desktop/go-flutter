package plugin

import (
	"fmt"
	"sync"
)

// StreamHandler defines the interface for a stream handler setup and tear-down
// requests.
type StreamHandler interface {
	// OnListen handles a request to set up an event stream.
	OnListen(arguments interface{}, sink *EventSink)
	// OnCancel handles a request to tear down the most recently created event
	// stream.
	OnCancel(arguments interface{})
}

// EventSink defines the interface for producers of events to send message to
// Flutter. StreamHandler act as a clients of EventSink for sending events.
type EventSink struct {
	eventChannel *EventChannel

	hasEnded bool
	sync.Mutex
}

// Success consumes a successful event.
func (es *EventSink) Success(event interface{}) {
	es.Lock()
	defer es.Unlock()
	if es.hasEnded || es != es.eventChannel.activeSink {
		return
	}

	binaryMsg, err := es.eventChannel.methodCodec.EncodeSuccessEnvelope(event)
	if err != nil {
		fmt.Printf("go-flutter: failed to encode success envelope for event channel '%s', error: %v\n", es.eventChannel.channelName, err)
	}
	err = es.eventChannel.messenger.Send(es.eventChannel.channelName, binaryMsg)
	if err != nil {
		fmt.Printf("go-flutter: failed to send Success message on event channel '%s', error: %v\n", es.eventChannel.channelName, err)
	}
}

// Error consumes an error event.
func (es *EventSink) Error(errorCode string, errorMessage string, errorDetails interface{}) {
	es.Lock()
	defer es.Unlock()
	if es.hasEnded || es != es.eventChannel.activeSink {
		return
	}

	binaryMsg, err := es.eventChannel.methodCodec.EncodeErrorEnvelope(errorCode, errorMessage, errorDetails)
	if err != nil {
		fmt.Printf("go-flutter: failed to encode success envelope for event channel '%s', error: %v\n", es.eventChannel.channelName, err)
	}
	err = es.eventChannel.messenger.Send(es.eventChannel.channelName, binaryMsg)
	if err != nil {
		fmt.Printf("go-flutter: failed to send Error message on event channel '%s', error: %v\n", es.eventChannel.channelName, err)
	}
}

// EndOfStream consumes end of stream.
func (es *EventSink) EndOfStream() {
	es.Lock()
	defer es.Unlock()
	if es.hasEnded || es != es.eventChannel.activeSink {
		return
	}
	es.hasEnded = true

	err := es.eventChannel.messenger.Send(es.eventChannel.channelName, nil)
	if err != nil {
		fmt.Printf("go-flutter: failed to send EndOfStream message on event channel '%s', error: %v\n", es.eventChannel.channelName, err)
	}
}
