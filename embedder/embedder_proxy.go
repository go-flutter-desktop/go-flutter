package embedder

// #include "embedder.h"
// void proxy_texture_destruction_callback(void* user_data);
import "C"
import (
	"runtime"
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
func proxy_gl_external_texture_frame_callback(userData unsafe.Pointer,
	textureID int64,
	width C.size_t,
	height C.size_t,
	texture *C.FlutterOpenGLTexture) C.bool {
	flutterEnginePointer := *(*uintptr)(userData)
	flutterEngine := (*FlutterEngine)(unsafe.Pointer(flutterEnginePointer))
	embedderGLTexture := flutterEngine.GLExternalTextureFrameCallback(textureID, int(width), int(height))
	if embedderGLTexture == nil {
		return C.bool(false)
	}
	texture.target = C.uint32_t(embedderGLTexture.Target)
	texture.name = C.uint32_t(embedderGLTexture.Name)
	texture.format = C.uint32_t(embedderGLTexture.Format)
	collectPointer := (unsafe.Pointer(&embedderGLTexture.Collect))
	texture.user_data = collectPointer
	texture.destruction_callback = (C.VoidCallback)(C.proxy_texture_destruction_callback)
	defer func() {
		runtime.KeepAlive(collectPointer)
	}()
	return C.bool(true)
}

//export proxy_texture_destruction_callback
func proxy_texture_destruction_callback(userData unsafe.Pointer) {
	defer recover()
	destroyCB := (*func())(userData)
	if destroyCB != nil {
		destroyCB := *destroyCB
		destroyCB()
	}
}
