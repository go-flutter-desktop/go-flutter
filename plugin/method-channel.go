package plugin

import (
	"fmt"
	"os"
	"regexp"
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

	methods     map[string]methodHandlerRegistration
	prefixRoute []*route
	methodsLock sync.RWMutex
}

// PrefixMethodChannel provides a way of defining MethodHandler based on a
// regex instead of a hard defined string.
type PrefixMethodChannel struct {
	methodChannel *MethodChannel

	// the route regex used by the MethodHandler
	route *route
}

type methodHandlerRegistration struct {
	handler MethodHandler
	sync    bool
}

type route struct {
	name        string
	routeRegexp *regexp.Regexp
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

// InvokeMethod sends a methodcall to the binary messenger and waits for a
// result. Results from the Flutter side are not yet implemented in the
// embedder. Until then, InvokeMethod will always return nil as reult.
// https://github.com/flutter/flutter/issues/18852
func (m *MethodChannel) InvokeMethod(name string, arguments interface{}) (result interface{}, err error) {
	encodedMessage, err := m.methodCodec.EncodeMethodCall(MethodCall{
		Method:    name,
		Arguments: arguments,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode methodcall")
	}
	encodedReply, err := m.messenger.Send(m.channelName, encodedMessage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send methodcall")
	}
	// TODO(GeertJohan): InvokeMethod may not return any JSON. In Java this is
	// handled by not having a callback handler, which means no reponse is
	// expected and reponse is never unmarshalled. We should perhaps define
	// InvokeMethod(..) and InovkeMethodNoResponse(..) to avoid errors when no
	// response is given.
	// https://github.com/go-flutter-desktop/go-flutter/issues/141
	result, err = m.methodCodec.DecodeEnvelope(encodedReply)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *MethodChannel) addRouteSync(r *route, handler MethodHandler) {
	m.pushNewRoute(r, handler, true)
}
func (m *MethodChannel) addRoute(r *route, handler MethodHandler) {
	m.pushNewRoute(r, handler, false)
}

func (m *MethodChannel) pushNewRoute(r *route, handler MethodHandler, sync bool) {
	m.methodsLock.Lock()
	if handler == nil {
		// delete from map
		delete(m.methods, r.name)
		// delete from slice
		i := 0
		for _, x := range m.prefixRoute {
			if x.name != r.name {
				m.prefixRoute[i] = x
				i++
			}
		}
		m.prefixRoute = m.prefixRoute[:i]
	} else {
		if r.routeRegexp != nil {
			m.prefixRoute = append(m.prefixRoute, r)
		}
		m.methods[r.name] = methodHandlerRegistration{
			handler: handler,
			sync:    sync,
		}
	}
	m.methodsLock.Unlock()
}

// Handle registers a method handler for method calls with given name.
//
// Consecutive calls override any existing handler registration for (the name
// of) this method. When given nil as handler, the previously registered
// handler for a method is unregistrered.
//
// When no handler is registered for a method, it will be handled silently by
// sending a nil reply which triggers the dart MissingPluginException exception.
func (m *MethodChannel) Handle(methodName string, handler MethodHandler) {
	route := &route{
		name:        methodName,
		routeRegexp: nil}
	m.addRoute(route, handler)
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
	route := &route{
		name:        methodName,
		routeRegexp: nil}
	m.addRouteSync(route, handler)
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

		// finding a route that matches
		var registration methodHandlerRegistration
		m.methodsLock.RLock()
		for _, re := range m.prefixRoute {
			if re.routeRegexp.MatchString(methodCall.Method) {
				registration = m.methods[re.name]
				break
			}
		}
		m.methodsLock.RUnlock()

		if registration == (methodHandlerRegistration{}) {
			fmt.Printf("go-flutter: no method handler registered for method '%s' on channel '%s'\n", methodCall.Method, m.channelName)
			responseSender.Send(nil)
			return nil
		}

		if registration.sync {
			m.handleMethodCall(registration.handler, methodCall.Method, methodCall, responseSender)
		} else {
			go m.handleMethodCall(registration.handler, methodCall.Method, methodCall, responseSender)
		}
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
		binaryReply, err := m.methodCodec.EncodeErrorEnvelope("error", err.Error(), nil)
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

// PathPrefix registers a new route with a regex for the methodName.
// PrefixMethodChannel expose a similar interface as MethodChannel.
//
// Instead of having a hard defined methodName string. The methodName is
// matched to the tpl regex, if the match succeed,
// PrefixMethodChannel.MethodHandler is used.
//
// If you use multiples PathPrefix in your plugin, the selection of the
// MethodHandler follows the same as the order as the HandleFunc definition.
func (m *MethodChannel) PathPrefix(tpl string) *PrefixMethodChannel {
	regex, err := regexp.Compile(tpl)
	if err != nil {
		fmt.Printf("go-flutter: couldn't compile the PathPrefix regex: '%s', error: %v\n", tpl, err)
		os.Exit(1)
	}

	route := &route{
		name:        tpl,
		routeRegexp: regex}

	pmc := &PrefixMethodChannel{
		route: route,
	}
	pmc.methodChannel = m

	return pmc
}

// Handle registers a method handler for method calls with a name that matches
// the routeRegexp.
//
// The MethodHandler defined on PrefixMethodChannel are only called if no
// MethodChannel MethodHandler are found.
// PrefixMethodChannel has a lower priority than MethodChannel.
//
// Consecutive calls override any existing handler registration for (the name
// of) this method. When given nil as handler, the previously registered
// handler for a method is unregistrered.
//
// When no handler is registered for a method, it will be handled silently by
// sending a nil reply which triggers the dart MissingPluginException exception.
func (pcm *PrefixMethodChannel) Handle(handler MethodHandler) {
	pcm.methodChannel.addRoute(pcm.route, handler)
}

// HandleFunc is a shorthand for m.Handle(MethodHandlerFunc(f))
//
// The argument of the function f is a MethodCall struct
func (pcm *PrefixMethodChannel) HandleFunc(f func(methodCall interface{}) (reply interface{}, err error)) {
	if f == nil {
		pcm.Handle(nil)
		return
	}

	pcm.Handle(MethodHandlerFunc(f))
}

// HandleSync is like Handle, but messages for given method are handled
// synchronously.
func (pcm *PrefixMethodChannel) HandleSync(handler MethodHandler) {
	pcm.methodChannel.addRouteSync(pcm.route, handler)

}

// HandleFuncSync is a shorthand for m.HandleSync(MethodHandlerFunc(f))
//
// The argument of the function f is a MethodCall struct
func (pcm *PrefixMethodChannel) HandleFuncSync(f func(methodCall interface{}) (reply interface{}, err error)) {
	if f == nil {
		pcm.HandleSync(nil)
		return
	}

	pcm.HandleSync(MethodHandlerFunc(f))
}
