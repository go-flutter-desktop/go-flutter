package flutter

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-flutter-desktop/go-flutter/embedder"
	"github.com/go-flutter-desktop/go-flutter/internal/tasker"
	"github.com/go-flutter-desktop/go-flutter/plugin"
)

type messenger struct {
	engine *embedder.FlutterEngine

	channels     map[string]plugin.ChannelHandlerFunc
	channelsLock sync.RWMutex

	// engineTasker holds tasks which must be executed in the engine thread
	engineTasker *tasker.Tasker
}

var _ plugin.BinaryMessenger = &messenger{}

func newMessenger(engine *embedder.FlutterEngine) *messenger {
	return &messenger{
		engine:       engine,
		channels:     make(map[string]plugin.ChannelHandlerFunc),
		engineTasker: tasker.New(),
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

	// NOTE: Response from engine is not yet supported by embedder.
	// https://github.com/flutter/flutter/issues/18852
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

	if channelHander == nil {
		// print a log, but continue on to send a nil reply when required
		fmt.Println("go-flutter: no handler found for channel " + message.Channel)
		return
	}

	var err error
	err = channelHander(message.Message, responseSender{
		engine:       m.engine,
		message:      message,
		engineTasker: m.engineTasker,
	})
	if err != nil {
		fmt.Printf("go-flutter: handling message on channel "+message.Channel+" failed: %v\n", err)
	}
}

type responseSender struct {
	engine       *embedder.FlutterEngine
	message      *embedder.PlatformMessage
	engineTasker *tasker.Tasker
}

var _ plugin.ResponseSender = responseSender{} // compile-time type check

func (r responseSender) Send(binaryReply []byte) {
	if !r.message.ExpectsResponse() {
		return // quick path when no response should be sent
	}

	// TODO: detect multiple responses on the same message and spam the log
	// about it.

	// It would be preferable to replace this with channels so sending
	// doesn't have to wait on the main loop to come around.
	go r.engineTasker.Do(func() {
		res := r.engine.SendPlatformMessageResponse(r.message.ResponseHandle, binaryReply)
		if res != embedder.ResultSuccess {
			fmt.Println("go-flutter: failed sending response for message on channel " + r.message.Channel)
		}
	})
}
