package plugin

import (
	"fmt"
	"runtime/debug"

	"github.com/pkg/errors"
)

// EventChannel provides way for flutter applications and hosts to communicate
// using event streams.
// It must be used with a codec, for example the StandardMethodCodec.
type EventChannel struct {
	messenger   BinaryMessenger
	channelName string
	methodCodec MethodCodec

	handler    StreamHandler
	activeSink *EventSink
}

// NewEventChannel creates a new event channel
func NewEventChannel(messenger BinaryMessenger, channelName string, methodCodec MethodCodec) (channel *EventChannel) {
	ec := &EventChannel{
		messenger:   messenger,
		channelName: channelName,
		methodCodec: methodCodec,
	}
	messenger.SetChannelHandler(channelName, ec.handleChannelMessage)
	return ec
}

// Handle registers a StreamHandler for a event channel.
//
// Consecutive calls override any existing handler registration.
// When given nil as handler, the previously registered
// handler for a method is unregistrered.
//
// When no handler is registered for a method, it will be handled silently by
// sending a nil reply which triggers the dart MissingPluginException exception.
func (e *EventChannel) Handle(handler StreamHandler) {
	e.handler = handler
}

// handleChannelMessage decodes incoming binary message to a method call, calls the
// handler, and encodes the outgoing reply.
func (e *EventChannel) handleChannelMessage(binaryMessage []byte, responseSender ResponseSender) (err error) {
	methodCall, err := e.methodCodec.DecodeMethodCall(binaryMessage)
	if err != nil {
		return errors.Wrap(err, "failed to decode incoming message")
	}

	if e.handler == nil {
		fmt.Printf("go-flutter: no method handler registered for event channel '%s'\n", e.channelName)
		responseSender.Send(nil)
		return nil
	}

	defer func() {
		p := recover()
		if p != nil {
			fmt.Printf("go-flutter: recovered from panic while handling message for event channel '%s': %v\n", e.channelName, p)
			debug.PrintStack()
		}
	}()

	switch methodCall.Method {
	case "listen":

		binaryReply, err := e.methodCodec.EncodeSuccessEnvelope(nil)
		if err != nil {
			fmt.Printf("go-flutter: failed to encode listen envelope for event channel '%s', error: %v\n", e.channelName, err)
		}
		responseSender.Send(binaryReply)

		if e.activeSink != nil {
			// Repeated calls to onListen may happen during hot restart.
			// We separate them with a call to onCancel.
			e.handler.OnCancel(nil)
		}

		e.activeSink = &EventSink{eventChannel: e}
		go e.handler.OnListen(methodCall.Arguments, e.activeSink)

	case "cancel":
		if e.activeSink != nil {
			e.activeSink = nil
			go e.handler.OnCancel(methodCall.Arguments)

			binaryReply, _ := e.methodCodec.EncodeSuccessEnvelope(nil)
			responseSender.Send(binaryReply)
		} else {
			fmt.Printf("go-flutter: No active stream to cancel onEventChannel '%s'\n", e.channelName)
			binaryReply, _ := e.methodCodec.EncodeErrorEnvelope("error", "No active stream to cancel", nil)
			responseSender.Send(binaryReply)
		}

	default:
		fmt.Printf("go-flutter: no StreamHandler handler registered for method '%s' on EventChannel '%s'\n", methodCall.Method, e.channelName)
		responseSender.Send(nil) // MissingPluginException
	}

	return nil
}
