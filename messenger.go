package flutter

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-flutter-desktop/go-flutter/embedder"
	"github.com/go-flutter-desktop/go-flutter/plugin"
)

type messenger struct {
	engine *embedder.FlutterEngine

	channels     map[string]plugin.ChannelHandlerFunc
	channelsLock sync.RWMutex
}

var _ plugin.BinaryMessenger = &messenger{}

func newMessenger(engine *embedder.FlutterEngine) *messenger {
	return &messenger{
		engine:   engine,
		channels: make(map[string]plugin.ChannelHandlerFunc),
	}
}

// Send pushes a binary message on a channel to the Flutter side. Replies are
// not supported yet (https://github.com/flutter/flutter/issues/18852). This
// means that currently, binaryReply will be nil on success.
func (m *messenger) Send(channel string, binaryMessage []byte) (binaryReply []byte, err error) {
	msg := &embedder.PlatformMessage{
		Channel: channel,
		Message: binaryMessage,
	}
	res := m.engine.SendPlatformMessage(msg)
	if err != nil {
		if ferr, ok := err.(*plugin.FlutterError); ok {
			return nil, ferr
		}
	}
	if res != embedder.ResultSuccess {
		return nil, errors.New("failed to send message")
	}
	return nil, nil
}

// SetChannelHandler satisfies plugin.BinaryMessenger
func (m *messenger) SetChannelHandler(channel string, channelHandler plugin.ChannelHandlerFunc) {
	m.channelsLock.Lock()
	if channelHandler == nil {
		delete(m.channels, channel)
	} else {
		m.channels[channel] = channelHandler
	}
	m.channelsLock.Unlock()
}

func (m *messenger) handlePlatformMessage(message *embedder.PlatformMessage) {
	m.channelsLock.RLock()
	channelHander := m.channels[message.Channel]
	m.channelsLock.RUnlock()

	var encodedReply []byte

	if channelHander == nil {
		// print a log, but continue on to send a nil reply when required
		fmt.Println("go-flutter: no handler found for channel " + message.Channel)
	} else {
		var err error
		encodedReply, err = channelHander(message.Message)
		if err != nil {
			// print a log, but continue on to send a nil reply when required
			fmt.Printf("go-flutter: handling message on channel "+message.Channel+" failed: %v\n", err)
			// force encodedReply to be nil, it may be sent when the dart side expects a reply.
			encodedReply = nil
		}
	}

	if message.ExpectsReply() {
		res := m.engine.SendPlatformMessageResponse(message.ResponseHandle, encodedReply)
		if res != embedder.ResultSuccess {
			fmt.Println("go-flutter: failed sending reply for message on channel " + message.Channel)
		}
	}
}
