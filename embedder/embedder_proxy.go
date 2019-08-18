package embedder

// #include "embedder.h"
import "C"
import (
	"unsafe"
)

//export proxy_platform_message_callback
func proxy_platform_message_callback(message *C.FlutterPlatformMessage, userData unsafe.Pointer) {
	msg := &PlatformMessage{
		Channel:        C.GoString(message.channel),
		Message:        C.GoBytes(unsafe.Pointer(message.message), C.int(message.message_size)),
		ResponseHandle: PlatformMessageResponseHandle(unsafe.Pointer(message.response_handle)),
	}
	flutterEnginePointer := *(*uintptr)(userData)
	flutterEngine := (*FlutterEngine)(unsafe.Pointer(flutterEnginePointer))
	flutterEngine.PlatfromMessage(msg)
}

//export proxy_make_current
func proxy_make_current(userData unsafe.Pointer) C.bool {
	flutterEnginePointer := *(*uintptr)(userData)
	flutterEngine := (*FlutterEngine)(unsafe.Pointer(flutterEnginePointer))
	return C.bool(flutterEngine.GLMakeCurrent())
}

//export proxy_clear_current
func proxy_clear_current(userData unsafe.Pointer) C.bool {
	flutterEnginePointer := *(*uintptr)(userData)
	flutterEngine := (*FlutterEngine)(unsafe.Pointer(flutterEnginePointer))
	return C.bool(flutterEngine.GLClearCurrent())
}

//export proxy_present
func proxy_present(userData unsafe.Pointer) C.bool {
	flutterEnginePointer := *(*uintptr)(userData)
	flutterEngine := (*FlutterEngine)(unsafe.Pointer(flutterEnginePointer))
	return C.bool(flutterEngine.GLPresent())
}

//export proxy_fbo_callback
func proxy_fbo_callback(userData unsafe.Pointer) C.uint32_t {
	flutterEnginePointer := *(*uintptr)(userData)
	flutterEngine := (*FlutterEngine)(unsafe.Pointer(flutterEnginePointer))
	return C.uint32_t(flutterEngine.GLFboCallback())
}

//export proxy_make_resource_current
func proxy_make_resource_current(userData unsafe.Pointer) C.bool {
	flutterEnginePointer := *(*uintptr)(userData)
	flutterEngine := (*FlutterEngine)(unsafe.Pointer(flutterEnginePointer))
	return C.bool(flutterEngine.GLMakeResourceCurrent())
}

//export proxy_gl_proc_resolver
func proxy_gl_proc_resolver(userData unsafe.Pointer, procname *C.char) unsafe.Pointer {
	flutterEnginePointer := *(*uintptr)(userData)
	flutterEngine := (*FlutterEngine)(unsafe.Pointer(flutterEnginePointer))
	return flutterEngine.GLProcResolver(C.GoString(procname))
}

//export proxy_gl_external_texture_frame_callback
func proxy_gl_external_texture_frame_callback(
	userData unsafe.Pointer,
	textureID int64,
	width C.size_t,
	height C.size_t,
	texture *C.FlutterOpenGLTexture,
) C.bool {
	flutterEnginePointer := *(*uintptr)(userData)
	flutterEngine := (*FlutterEngine)(unsafe.Pointer(flutterEnginePointer))
	embedderGLTexture := flutterEngine.GLExternalTextureFrameCallback(textureID, int(width), int(height))
	if embedderGLTexture == nil {
		return C.bool(false)
	}
	texture.target = C.uint32_t(embedderGLTexture.Target)
	texture.name = C.uint32_t(embedderGLTexture.Name)
	texture.format = C.uint32_t(embedderGLTexture.Format)
	return C.bool(true)
}

//export proxy_runs_task_on_current_thread_callback
func proxy_runs_task_on_current_thread_callback(userData unsafe.Pointer) C.bool {
	flutterEnginePointer := *(*uintptr)(userData)
	flutterEngine := (*FlutterEngine)(unsafe.Pointer(flutterEnginePointer))
	return C.bool(flutterEngine.TaskRunnerRunOnCurrentThread())
}

//export proxy_post_task_callback
func proxy_post_task_callback(task C.FlutterTask, targetTimeNanos C.uint64_t, userData unsafe.Pointer) {
	flutterEnginePointer := *(*uintptr)(userData)
	flutterEngine := (*FlutterEngine)(unsafe.Pointer(flutterEnginePointer))
	flutterEngine.TaskRunnerPostTask(task, uint64(targetTimeNanos))
}

//export proxy_desktop_binary_reply
func proxy_desktop_binary_reply(data *C.uint8_t, dataSize C.size_t, userData unsafe.Pointer) {
	callbackPointer := *(*uintptr)(userData)
	handler := *(*DataCallback)(unsafe.Pointer(callbackPointer))
	handler(C.GoBytes(unsafe.Pointer(data), C.int(dataSize)))
}
