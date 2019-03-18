package embedder

// #include "flutter_embedder.h"
import "C"
import (
	"unsafe"

	"github.com/go-gl/glfw/v3.2/glfw"
)

//export proxy_platform_message_callback
func proxy_platform_message_callback(message *C.FlutterPlatformMessage, v unsafe.Pointer) {
	msg := &PlatformMessage{
		Channel:        C.GoString(message.channel),
		Message:        C.GoBytes(unsafe.Pointer(message.message), C.int(message.message_size)),
		ResponseHandle: PlatformMessageResponseHandle(unsafe.Pointer(message.response_handle)),
	}
	w := glfw.GoWindow(v)
	index := *(*int)(w.GetUserPointer())
	flutterEngine := FlutterEngineByIndex(index)
	if flutterEngine == nil {
		return
	}
	flutterEngine.FPlatfromMessage(msg)
}

//export proxy_make_current
func proxy_make_current(v unsafe.Pointer) C.bool {
	w := glfw.GoWindow(v)
	// TODO(#89): Move this to main thread when we have a renderer
	// interface (which must implement these callbacks)
	index := *(*int)(w.GetUserPointer())
	flutterEngine := FlutterEngineByIndex(index)
	// TODO(#89): remove this workarround, flutter engine should never be nil, but
	// this can still happen because we cannot guarantee
	// (*glfw.Window).GetUserPointer to be executed on the main thread.
	if flutterEngine == nil {
		return C.bool(false)
	}
	return C.bool(flutterEngine.FMakeCurrent(v))
}

//export proxy_clear_current
func proxy_clear_current(v unsafe.Pointer) C.bool {
	w := glfw.GoWindow(v)
	// TODO(#89): Move this to main thread when we have a renderer
	// interface (which must implement these callbacks)
	index := *(*int)(w.GetUserPointer())
	flutterEngine := FlutterEngineByIndex(index)
	// TODO(#89): remove this workarround, flutter engine should never be nil, but
	// this can still happen because we cannot guarantee
	// (*glfw.Window).GetUserPointer to be executed on the main thread.
	if flutterEngine == nil {
		return C.bool(false)
	}
	return C.bool(flutterEngine.FClearCurrent(v))
}

//export proxy_present
func proxy_present(v unsafe.Pointer) C.bool {
	w := glfw.GoWindow(v)
	// TODO(#89): Move this to main thread when we have a renderer
	// interface (which must implement these callbacks)
	index := *(*int)(w.GetUserPointer())
	flutterEngine := FlutterEngineByIndex(index)
	// TODO(#89): remove this workarround, flutter engine should never be nil, but
	// this can still happen because we cannot guarantee
	// (*glfw.Window).GetUserPointer to be executed on the main thread.
	if flutterEngine == nil {
		return C.bool(false)
	}
	return C.bool(flutterEngine.FPresent(v))
}

//export proxy_fbo_callback
func proxy_fbo_callback(v unsafe.Pointer) C.uint32_t {
	w := glfw.GoWindow(v)
	// TODO(#89): Move this to main thread when we have a renderer
	// interface (which must implement these callbacks)
	index := *(*int)(w.GetUserPointer())
	flutterEngine := FlutterEngineByIndex(index)
	// TODO(#89): remove this workarround, flutter engine should never be nil, but
	// this can still happen because we cannot guarantee
	// (*glfw.Window).GetUserPointer to be executed on the main thread.
	if flutterEngine == nil {
		panic("note me")
		return C.uint32_t(0)
	}
	return C.uint32_t(flutterEngine.FFboCallback(v))
}

//export proxy_make_resource_current
func proxy_make_resource_current(v unsafe.Pointer) C.bool {
	w := glfw.GoWindow(v)
	// TODO(#89): Move this to main thread when we have a renderer
	// interface (which must implement these callbacks)
	index := *(*int)(w.GetUserPointer())
	flutterEngine := FlutterEngineByIndex(index)
	// TODO(#89): remove this workarround, flutter engine should never be nil, but
	// this can still happen because we cannot guarantee
	// (*glfw.Window).GetUserPointer to be executed on the main thread.
	if flutterEngine == nil {
		return C.bool(false)
	}
	return C.bool(flutterEngine.FMakeResourceCurrent(v))
}

//export proxy_gl_proc_resolver
func proxy_gl_proc_resolver(v unsafe.Pointer, procname *C.char) unsafe.Pointer {
	return glfw.GetProcAddress(C.GoString(procname))
}
