package plugin

import (
	"fmt"

	"github.com/pkg/errors"
)

// EventChannel provides way for flutter applications and hosts to communicate
// using event streams.
// It must be used with a codec, for example the StandardMethodCodec.
type EventChannel struct {
	messenger   BinaryMessenger
	channelName string
	methodCodec MethodCodec

	handler StreamHandler
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

// handleChannelMessage decodes incoming binary message to a method call, calls the
// handler, and encodes the outgoing reply.
func (e *EventChannel) handleChannelMessage(binaryMessage []byte, responseSender ResponseSender) (err error) {
	methodCall, err := e.methodCodec.DecodeMethodCall(binaryMessage)
	if err != nil {
		return errors.Wrap(err, "failed to decode incomming message")
	}

	if e.handler == nil {
		fmt.Printf("go-flutter: no method handler registered for event channel '%s'\n", e.channelName)
		responseSender.Send(nil)
		return nil
	}

	switch methodCall.Method {
	case "listen":
		binaryReply, err := e.methodCodec.EncodeSuccessEnvelope(nil)
		if err != nil {
			fmt.Printf("go-flutter: failed to encode listen envelope for event channel '%s', error: %v\n", e.channelName, err)
		}
		responseSender.Send(binaryReply)

		sink := &EventSink{
			messenger:   e.messenger,
			methodCodec: e.methodCodec,
			channelName: e.channelName,
		}
		go e.handler.OnListen(methodCall.Arguments, sink)

	case "cancel":
		binaryReply, err := e.methodCodec.EncodeSuccessEnvelope(nil)
		if err != nil {
			fmt.Printf("go-flutter: failed to encode cancel envelope for event channel '%s', error: %v\n", e.channelName, err)
		}
		responseSender.Send(binaryReply)

		go e.handler.OnCancel(methodCall.Arguments)
	default:
		fmt.Printf("go-flutter: no StreamHandler handler registered for method '%s' on EventChannel '%s'\n", methodCall.Method, e.channelName)
		responseSender.Send(nil) // MissingPluginException
	}

	return nil

}

// Handle registers a method handler for a event channel.
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

// HandleFunc is a shorthand for m.Handle(MethodHandlerFunc(f))
func (e *EventChannel) HandleFunc(onListen func(arguments interface{}, sink *EventSink),
	onCancel func(arguments interface{})) {
	if onListen == nil || onCancel == nil {
		e.Handle(nil)
		return
	}

	f := StreamHandlerFunc{
		onListen: onListen,
		onCancel: onCancel,
	}
	e.Handle(f)
}
