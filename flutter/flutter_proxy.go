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
	"fmt"
	"unsafe"
)

// C proxies

//export proxy_on_platform_message
func proxy_on_platform_message(message *C.FlutterPlatformMessage, userPointer unsafe.Pointer) C.bool {
	if message.message != nil {
		str := C.GoStringN(C.c_str(message.message), C.int(message.message_size))

		FlutterPlatformMessage := PlatformMessage{}

		messageContent := Message{}
		json.Unmarshal([]byte(str), &messageContent)

		FlutterPlatformMessage.Message = messageContent
		FlutterPlatformMessage.Channel = C.GoString(message.channel)
		if message.response_handle == nil {
			fmt.Println("==================== NIL")
		}

		return C.bool(globalFlutterOpenGL.FPlatfromMessage(FlutterPlatformMessage, userPointer))
	}
	return C.bool(false)

}

//export proxy_make_current
func proxy_make_current(v unsafe.Pointer) C.bool {
	return C.bool(globalFlutterOpenGL.FMakeCurrent(v))
}

//export proxy_clear_current
func proxy_clear_current(v unsafe.Pointer) C.bool {
	return C.bool(globalFlutterOpenGL.FClearCurrent(v))
}

//export proxy_present
func proxy_present(v unsafe.Pointer) C.bool {
	return C.bool(globalFlutterOpenGL.FPresent(v))
}

//export proxy_fbo_callback
func proxy_fbo_callback(v unsafe.Pointer) C.uint32_t {
	return C.uint32_t(globalFlutterOpenGL.FFboCallback(v))
}

//export proxy_make_resource_current
func proxy_make_resource_current(v unsafe.Pointer) C.bool {
	return C.bool(globalFlutterOpenGL.FMakeResourceCurrent(v))
}
