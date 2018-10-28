package flutter

/*
#include "flutter_embedder.h"

static char* converter(uint8_t *str, size_t size){
	str[size] = '\0'; // Prevent overFlow
	return (char *)str;
}


#include <stdio.h>
static int am_i_null(void* pointer) {
  if (NULL == pointer) {
    printf("NULL");
  }
}

bool proxy_make_current(void *v);

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
		str := C.GoString(C.converter(message.message, message.message_size))

		FlutterPlatformMessage := PlatformMessage{}

		messageContent := Message{}
		json.Unmarshal([]byte(str), &messageContent)

		FlutterPlatformMessage.Message = messageContent
		FlutterPlatformMessage.Channel = C.GoString(message.channel)
		// fmt.Println(unsafe.Pointer(message.response_handle))
		// fmt.Println(str)
		if message.response_handle == nil {
			fmt.Println("==================== NIL")
		}

		response := "[ 1 ]"

		C.FlutterEngineSendPlatformMessageResponse(
			flutterEngines[0].Engine,
			message.response_handle,
			(*C.uint8_t)(unsafe.Pointer(C.CString(response))), (C.ulong)(len(response)))

		return C.bool(flutterEngines[0].FPlatfromMessage(FlutterPlatformMessage, userPointer))
	}
	return C.bool(false)

}

//export proxy_make_current
func proxy_make_current(v unsafe.Pointer) C.bool {
	return C.bool(flutterEngines[0].FMakeCurrent(v))
}

//export proxy_make_current_test
func proxy_make_current_test(engineID int, v unsafe.Pointer) C.bool {
	return C.bool(flutterEngines[engineID].FMakeCurrent(v))
}

//export proxy_make_current_func
func proxy_make_current_func(engineID int) unsafe.Pointer {

	print(engineID)
	return C.proxy_make_current
}

//export proxy_clear_current
func proxy_clear_current(v unsafe.Pointer) C.bool {
	return C.bool(flutterEngines[0].FClearCurrent(v))
}

//export proxy_present
func proxy_present(v unsafe.Pointer) C.bool {
	return C.bool(flutterEngines[0].FPresent(v))
}

//export proxy_fbo_callback
func proxy_fbo_callback(v unsafe.Pointer) C.uint32_t {
	return C.uint32_t(flutterEngines[0].FFboCallback(v))
}

//export proxy_make_resource_current
func proxy_make_resource_current(v unsafe.Pointer) C.bool {
	return C.bool(flutterEngines[0].FMakeResourceCurrent(v))
}
