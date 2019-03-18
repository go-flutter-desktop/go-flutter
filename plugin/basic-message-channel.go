package plugin

import "github.com/pkg/errors"

// BasicMessageHandler defines the interfece for a basic message handler.
type BasicMessageHandler interface {
	HandleMessage(message interface{}) (reply interface{}, err error)
}

// The BasicMessageHandlerFunc type is an adapter to allow the use of
// ordinary functions as basic message handlers. If f is a function
// with the appropriate signature, BasicMessageHandlerFunc(f) is a
// BasicMessageHandler that calls f.
type BasicMessageHandlerFunc func(message interface{}) (reply interface{}, err error)

// HandleMessage calls f(message).
func (f BasicMessageHandlerFunc) HandleMessage(message interface{}) (reply interface{}, err error) {
	return f(message)
}

// BasicMessageChannel presents named channel for communicating with the Flutter
// application using basic, asynchronous message passing.
//
// Messages are encoded into binary before being sent, and binary messages
/// received are decoded into. The MessageCodec used must be compatible with the
// one used by the Flutter application. This can be achieved by creating a
// BasicMessageChannel counterpart of this channel on the Dart side.
// See: https://docs.flutter.io/flutter/services/BasicMessageChannel-class.html
//
// The static Go type of messages sent and received is interface{}, but only
// values supported by the specified MessageCodec can be used.
//
// The logical identity of the channel is given by its name. Identically named
// channels will interfere with each other's communication.
type BasicMessageChannel struct {
	messenger BinaryMessenger
	name      string
	codec     MessageCodec
	handler   BasicMessageHandler
}

// NewBasicMessageChannel creates a BasicMessageChannel.
//
// Call Handle or HandleFunc on the returned BasicMessageChannel to provide the
// channel with a handler for incomming messages.
func NewBasicMessageChannel(messenger BinaryMessenger, name string, codec MessageCodec) *BasicMessageChannel {
	b := &BasicMessageChannel{
		messenger: messenger,
		name:      name,
		codec:     codec,
	}
	messenger.SetChannelHandler(b.name, b.handleChannel)
	return b
}

// Send encodes and sends the specified message to the Flutter application and
// returns the reply, or an error.
func (b *BasicMessageChannel) Send(message interface{}) (reply interface{}, err error) {
	encodedMessage, err := b.codec.EncodeMessage(message)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode outgoing message")
	}
	encodedReply, err := b.messenger.Send(b.name, encodedMessage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send outgoing message")
	}
	reply, err = b.codec.DecodeMessage(encodedReply)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode incomming reply")
	}
	return reply, nil
}

// Handle registers a message handler on this channel for receiving
// messages sent from the Flutter application.
//
// Consecutive calls override any existing handler registration for (the name
// of) this channel.
//
// When given nil as handler, any incoming message on this channel will be
// handled silently by sending a nil reply (null on the dart side).
func (b *BasicMessageChannel) Handle(handler BasicMessageHandler) {
	b.handler = handler
}

// HandleFunc is a shorthand for b.Handle(BasicMessageHandlerFunc(f))
func (b *BasicMessageChannel) HandleFunc(f func(message interface{}) (reply interface{}, err error)) {
	if f == nil {
		b.Handle(nil)
		return
	}

	b.Handle(BasicMessageHandlerFunc(f))
}

// handleChannel decodes an incoming binary envelopes, calls the bassic
// message handler, and encodes the outgoing reply into an envelope.
func (b *BasicMessageChannel) handleChannel(binaryMessage []byte) (binaryReply []byte, err error) {
	if b.handler == nil {
		return nil, nil
	}
	message, err := b.codec.DecodeMessage(binaryMessage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode incomming message")
	}
	reply, err := b.handler.HandleMessage(message)
	if err != nil {
		return nil, errors.Wrap(err, "handler for incoming basic message failed")
	}
	binaryReply, err = b.codec.EncodeMessage(reply)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode outgoing reply")
	}
	return binaryReply, nil
}

var _ ChannelHandlerFunc = (*BasicMessageChannel)(nil).handleChannel // compile-time type check
