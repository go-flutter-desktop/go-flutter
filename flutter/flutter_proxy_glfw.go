package flutter

/*
#include "flutter_embedder.h"

static char *c_str(uint8_t *str){
	return (char *)str;
}
*/
import "C"
import (
	"encoding/json"
	"unsafe"

	"github.com/go-gl/glfw/v3.2/glfw"
)

// C proxies

//export proxy_on_platform_message
func proxy_on_platform_message(message *C.FlutterPlatformMessage, userPointer unsafe.Pointer) C.bool {
	if message.message != nil {
		str := C.GoStringN(C.c_str(message.message), C.int(message.message_size))

		messageContent := Message{}
		json.Unmarshal([]byte(str), &messageContent)

		FlutterPlatformMessage := &PlatformMessage{
			Message:        messageContent,
			Channel:        C.GoString(message.channel),
			ResponseHandle: message.response_handle,
		}
		return C.bool(flutterEngines[0].FPlatfromMessage(FlutterPlatformMessage, userPointer))
	}
	return C.bool(false)

}

//export proxy_make_current
func proxy_make_current(v unsafe.Pointer) C.bool {
	w := glfw.GoWindow(v)
	index := *(*C.int)(w.GetUserPointer())
	return C.bool(flutterEngines[index].FMakeCurrent(v))
}

//export proxy_clear_current
func proxy_clear_current(v unsafe.Pointer) C.bool {
	w := glfw.GoWindow(v)
	index := *(*C.int)(w.GetUserPointer())
	return C.bool(flutterEngines[index].FClearCurrent(v))
}

//export proxy_present
func proxy_present(v unsafe.Pointer) C.bool {
	w := glfw.GoWindow(v)
	index := *(*C.int)(w.GetUserPointer())
	return C.bool(flutterEngines[index].FPresent(v))
}

//export proxy_fbo_callback
func proxy_fbo_callback(v unsafe.Pointer) C.uint32_t {
	w := glfw.GoWindow(v)
	index := *(*C.int)(w.GetUserPointer())
	return C.uint32_t(flutterEngines[index].FFboCallback(v))
}

//export proxy_make_resource_current
func proxy_make_resource_current(v unsafe.Pointer) C.bool {
	w := glfw.GoWindow(v)
	index := *(*C.int)(w.GetUserPointer())
	return C.bool(flutterEngines[index].FMakeResourceCurrent(v))
}
