package plugin

import (
	"errors"
	"sync"
)

// TestingBinaryMessenger implements the BinaryMessenger interface for testing
//  purposes. It can be used as a backend in tests for BasicMessageChannel and
// StandardMethodChannel.
// TODO: perhaps this can be exported (non-test) into a subpackage `plugintest`.
//   The TestingBinaryMessenger may then be used to test other plugins.
//   s/TestingBinaryMessenger/MockBinaryMessenger/ ?
type TestingBinaryMessenger struct {
	channelHandlersLock sync.Mutex
	channelHandlers     map[string]ChannelHandlerFunc

	// handlers mocking the other side of the BinaryMessenger
	mockChannelHandlersLock sync.Mutex
	mockChannelHandlers     map[string]ChannelHandlerFunc
}

func NewTestingBinaryMessenger() *TestingBinaryMessenger {
	return &TestingBinaryMessenger{
		channelHandlers:     make(map[string]ChannelHandlerFunc),
		mockChannelHandlers: make(map[string]ChannelHandlerFunc),
	}
}

var _ BinaryMessenger = &TestingBinaryMessenger{} // compile-time type check

func (t *TestingBinaryMessenger) Send(channel string, message []byte) (err error) {
	err = t.Send(channel, message)
	return err
}

// Send sends the bytes onto the given channel.
// In this testing implementation of a BinaryMessenger, the handler for the
// channel may be set using MockSetMessageHandler
func (t *TestingBinaryMessenger) SendWithReply(channel string, message []byte) (reply []byte, err error) {
	t.mockChannelHandlersLock.Lock()
	handler := t.mockChannelHandlers[channel]
	t.mockChannelHandlersLock.Unlock()
	if handler == nil {
		return nil, errors.New("no handler set")
	}

	r := mockResponseSender{}
	handler(message, &r)
	return r.binaryReply, nil
}

// SetMessageHandler registers a binary message handler on given channel.
// In this testing implementation of a BinaryMessenger, the handler may be
// executed by calling MockSend(..).
func (t *TestingBinaryMessenger) SetChannelHandler(channel string, handler ChannelHandlerFunc) {
	t.channelHandlersLock.Lock()
	t.channelHandlers[channel] = handler
	t.channelHandlersLock.Unlock()
}

// MockSend imitates a send method call from the other end of the binary
// messenger. It calls a method that was registered through SetMessageHandler.
func (t *TestingBinaryMessenger) MockSend(channel string, message []byte) (reply []byte, err error) {
	t.channelHandlersLock.Lock()
	handler := t.channelHandlers[channel]
	t.channelHandlersLock.Unlock()
	if handler == nil {
		return nil, errors.New("no handler set")
	}

	r := mockResponseSender{}
	handler(message, &r)
	return r.binaryReply, nil
}

// MockSetChannelHandler imitates a handler set at the other end of the
// binary messenger.
func (t *TestingBinaryMessenger) MockSetChannelHandler(channel string, handler ChannelHandlerFunc) {
	t.mockChannelHandlersLock.Lock()
	t.mockChannelHandlers[channel] = handler
	t.mockChannelHandlersLock.Unlock()
}

type mockResponseSender struct {
	binaryReply []byte
}

func (m *mockResponseSender) Send(binaryReply []byte) {
	m.binaryReply = binaryReply
}
