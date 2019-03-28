package plugin

// MethodHandler defines the interface for a method handler.
type MethodHandler interface {
	// HandleMethod is called whenever an incoming
	HandleMethod(arguments interface{}) (reply interface{}, err error)
}

// The MethodHandlerFunc type is an adapter to allow the use of
// ordinary functions as method handlers. If f is a function
// with the appropriate signature, MethodHandlerFunc(f) is a
// MethodHandler that calls f.
type MethodHandlerFunc func(arguments interface{}) (reply interface{}, err error)

// HandleMethod calls f(arguments).
func (f MethodHandlerFunc) HandleMethod(arguments interface{}) (reply interface{}, err error) {
	return f(arguments)
}
