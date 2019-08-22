package plugin

import (
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/pkg/errors"
)

// MethodChannel provides way for flutter applications and hosts to communicate.
// It must be used with a codec, for example the StandardMethodCodec. For more
// information please read
// https://flutter.dev/docs/development/platform-integration/platform-channels
type MethodChannel struct {
	messenger   BinaryMessenger
	channelName string
	methodCodec MethodCodec

	methods         map[string]methodHandlerRegistration
	catchAllhandler MethodHandler
	methodsLock     sync.RWMutex
}

type methodHandlerRegistration struct {
	handler MethodHandler
	sync    bool
}

// NewMethodChannel creates a new method channel
func NewMethodChannel(messenger BinaryMessenger, channelName string, methodCodec MethodCodec) (channel *MethodChannel) {
	mc := &MethodChannel{
		messenger:   messenger,
		channelName: channelName,
		methodCodec: methodCodec,

		methods: make(map[string]methodHandlerRegistration),
	}
	messenger.SetChannelHandler(channelName, mc.handleChannelMessage)
	return mc
}

// InvokeMethod sends a methodcall to the binary messenger without waiting for
// a reply.
func (m *MethodChannel) InvokeMethod(name string, arguments interface{}) error {
	encodedMessage, err := m.methodCodec.EncodeMethodCall(MethodCall{
		Method:    name,
		Arguments: arguments,
	})
	if err != nil {
		return errors.Wrap(err, "failed to encode methodcall")
	}
	err = m.messenger.Send(m.channelName, encodedMessage)
	if err != nil {
		return errors.Wrap(err, "failed to send methodcall")
	}
	return nil
}

// InvokeMethodWithReply sends a methodcall to the binary messenger and wait
// for a reply.
//
// NOTE: If no value are returned by the handler setted in the
// setMethodCallHandler flutter method, the function will wait forever. In case
// you don't want to wait for reply, use InvokeMethod or launch the
// function in a goroutine.
func (m *MethodChannel) InvokeMethodWithReply(name string, arguments interface{}) (result interface{}, err error) {
	encodedMessage, err := m.methodCodec.EncodeMethodCall(MethodCall{
		Method:    name,
		Arguments: arguments,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode methodcall")
	}
	encodedReply, err := m.messenger.SendWithReply(m.channelName, encodedMessage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send methodcall")
	}
	result, err = m.methodCodec.DecodeEnvelope(encodedReply)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Handle registers a method handler for method calls with given name.
//
// Consecutive calls override any existing handler registration for (the name
// of) this method. When given nil as handler, the previously registered
// handler for a method is unregistered.
//
// When no handler is registered for a method, it will be handled silently by
// sending a nil reply which triggers the dart MissingPluginException exception.
func (m *MethodChannel) Handle(methodName string, handler MethodHandler) {
	m.methodsLock.Lock()
	if handler == nil {
		delete(m.methods, methodName)
	} else {
		m.methods[methodName] = methodHandlerRegistration{
			handler: handler,
		}
	}
	m.methodsLock.Unlock()
}

// HandleFunc is a shorthand for m.Handle("name", MethodHandlerFunc(f))
//
// The argument of the function f is an interface corresponding to the
// MethodCall.Arguments values
func (m *MethodChannel) HandleFunc(methodName string, f func(arguments interface{}) (reply interface{}, err error)) {
	if f == nil {
		m.Handle(methodName, nil)
		return
	}

	m.Handle(methodName, MethodHandlerFunc(f))
}

// HandleSync is like Handle, but messages for given method are handled
// synchronously.
func (m *MethodChannel) HandleSync(methodName string, handler MethodHandler) {
	m.methodsLock.Lock()
	if handler == nil {
		delete(m.methods, methodName)
	} else {
		m.methods[methodName] = methodHandlerRegistration{
			handler: handler,
			sync:    true,
		}
	}
	m.methodsLock.Unlock()
}

// HandleFuncSync is a shorthand for m.HandleSync("name", MethodHandlerFunc(f))
//
// The argument of the function f is an interface corresponding to the
// MethodCall.Arguments values
func (m *MethodChannel) HandleFuncSync(methodName string, f func(arguments interface{}) (reply interface{}, err error)) {
	if f == nil {
		m.HandleSync(methodName, nil)
		return
	}

	m.HandleSync(methodName, MethodHandlerFunc(f))
}

// CatchAllHandle registers a default method handler.
// When no Handle are found, the handler provided in CatchAllHandle will be
// used. If no CatchAllHandle is provided, a MissingPluginException exception
// is sent when no handler is registered for a method name.
//
// Consecutive calls override any existing handler registration for (the name
// of) this method. When given nil as handler, the previously registered
// handler for a method is unregistered.
func (m *MethodChannel) CatchAllHandle(handler MethodHandler) {
	m.methodsLock.Lock()
	m.catchAllhandler = handler
	m.methodsLock.Unlock()
}

// CatchAllHandleFunc is a shorthand for m.CatchAllHandle(MethodHandlerFunc(f))
//
// The argument of the function f is a MethodCall struct
func (m *MethodChannel) CatchAllHandleFunc(f func(arguments interface{}) (reply interface{}, err error)) {
	m.CatchAllHandle(MethodHandlerFunc(f))
}

// handleChannelMessage decodes incoming binary message to a method call, calls the
// handler, and encodes the outgoing reply.
func (m *MethodChannel) handleChannelMessage(binaryMessage []byte, responseSender ResponseSender) (err error) {
	methodCall, err := m.methodCodec.DecodeMethodCall(binaryMessage)
	if err != nil {
		return errors.Wrap(err, "failed to decode incoming message")
	}

	m.methodsLock.RLock()
	registration, registrationExists := m.methods[methodCall.Method]
	m.methodsLock.RUnlock()
	if !registrationExists {

		if m.catchAllhandler != nil {
			go m.handleMethodCall(m.catchAllhandler, methodCall.Method, methodCall, responseSender)
			return nil
		}

		fmt.Printf("go-flutter: no method handler registered for method '%s' on channel '%s'\n", methodCall.Method, m.channelName)
		responseSender.Send(nil)
		return nil
	}

	if registration.sync {
		m.handleMethodCall(registration.handler, methodCall.Method, methodCall.Arguments, responseSender)
	} else {
		go m.handleMethodCall(registration.handler, methodCall.Method, methodCall.Arguments, responseSender)
	}

	return nil
}

// handleMethodCall handles the methodcall and sends a response.
func (m *MethodChannel) handleMethodCall(handler MethodHandler, methodName string, methodArgs interface{}, responseSender ResponseSender) {
	defer func() {
		p := recover()
		if p != nil {
			fmt.Printf("go-flutter: recovered from panic while handling call for method '%s' on channel '%s': %v\n", methodName, m.channelName, p)
			debug.PrintStack()
		}
	}()

	reply, err := handler.HandleMethod(methodArgs)
	if err != nil {
		fmt.Printf("go-flutter: handler for method '%s' on channel '%s' returned an error: %v\n", methodName, m.channelName, err)

		var errorCode string
		switch t := err.(type) {
		case *Error:
			errorCode = t.code
		default:
			errorCode = "error"
		}

		binaryReply, err := m.methodCodec.EncodeErrorEnvelope(errorCode, err.Error(), nil)
		if err != nil {
			fmt.Printf("go-flutter: failed to encode error envelope for method '%s' on channel '%s', error: %v\n", methodName, m.channelName, err)
		}
		responseSender.Send(binaryReply)
		return
	}
	binaryReply, err := m.methodCodec.EncodeSuccessEnvelope(reply)
	if err != nil {
		fmt.Printf("go-flutter: failed to encode success envelope for method '%s' on channel '%s', error: %v\n", methodName, m.channelName, err)
	}
	responseSender.Send(binaryReply)
}

// Error implement the Go error interface, can be thrown from a go-flutter
// method channel plugin to return custom error codes.
// Normal Go error can also be used, the error code will default to "error".
type Error struct {
	err  string
	code string
}

// Error is needed to comply with the Golang error interface.
func (e *Error) Error() string {
	return e.err
}

// NewError create an error with an specific error code.
func NewError(code string, err error) *Error {
	pe := &Error{
		code: code,
		err:  err.Error(),
	}
	return pe
}
