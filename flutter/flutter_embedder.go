package flutter

/*
#cgo CFLAGS: -I${SRCDIR}/library
#cgo LDFLAGS: -L${SRCDIR}/library/linux -lflutter_engine -Wl,-rpath,$ORIGIN/library/linux
#include "flutter_embedder.h"

bool proxy_make_current(void *v);
bool proxy_clear_current(void *v);
bool proxy_present(void *v);
uint32_t proxy_fbo_callback(void *v);
bool proxy_make_resource_current(void *v);

static FlutterResult runFlutter(uintptr_t window, FlutterEngine *engine, FlutterOpenGLRendererConfig * fop){

	FlutterRendererConfig config = {};
	config.type = kOpenGL;

	config.open_gl.struct_size = sizeof(FlutterOpenGLRendererConfig);

	config.open_gl.make_current = fop->make_current;
	config.open_gl.clear_current = fop->clear_current;
	config.open_gl.present = fop->present;
	config.open_gl.fbo_callback = fop->fbo_callback;
	config.open_gl.make_resource_current = fop->make_resource_current;

	const char *args_arr[] = {
		"",
		"--dart-non-checked-mode",
		NULL,
	};

	FlutterProjectArgs args = {};
	args.struct_size = sizeof(FlutterProjectArgs);
	args.main_path = "";
	// args.assets_path = "/opt/flutter-desktop-embedding/example_flutter/build/flutter_assets";
	// args.assets_path = "/home/drakirus/lab/flutter/smart_access/build/flutter_assets";
	args.assets_path = "/home/drakirus/lab/go/src/flutter_desktop_go_embedding/example_flutter/build/flutter_assets";
	args.packages_path = "";
	args.icu_data_path = "/home/drakirus/lab/go/src/flutter_desktop_go_embedding/icudtl.dat";
	args.command_line_argc = 2;
	args.command_line_argv = args_arr;
	// args.platform_message_callback = (FlutterPlatformMessageCallback)proxy_OnPlatformMessage;


	return FlutterEngineRun(FLUTTER_ENGINE_VERSION, &config, &args, (void*)window, engine);

}

*/
import "C"
import (
	"unsafe"

	"github.com/go-gl/glfw/v3.2/glfw"
)

//export proxy_make_current
func proxy_make_current(v unsafe.Pointer) C.bool {
	a := glfw.GoWindow(v)
	a.MakeContextCurrent()
	return C.bool(true)
}

//export proxy_clear_current
func proxy_clear_current(v unsafe.Pointer) C.bool {
	glfw.DetachCurrentContext()
	return C.bool(true)
}

//export proxy_present
func proxy_present(v unsafe.Pointer) C.bool {
	a := glfw.GoWindow(v)
	a.SwapBuffers()
	return C.bool(true)
}

//export proxy_fbo_callback
func proxy_fbo_callback(v unsafe.Pointer) C.uint32_t {
	return C.uint32_t(0)
}

//export proxy_make_resource_current
func proxy_make_resource_current(v unsafe.Pointer) C.bool {
	return C.bool(false)
}

type FlutterOpenGL struct {
	Engine               C.FlutterEngine
	FMakeCurrent         func(v unsafe.Pointer) bool
	FClearCurrent        func(v unsafe.Pointer) bool
	FPresent             func(v unsafe.Pointer) bool
	FFboCallback         func(v unsafe.Pointer) int32
	FMakeResourceCurrent func(v unsafe.Pointer) bool
}

func (flu *FlutterOpenGL) Run(window uintptr) FlutterResult {

	a := C.FlutterOpenGLRendererConfig{
		make_current:          (C.BoolCallback)(C.proxy_make_current),
		clear_current:         (C.BoolCallback)(C.proxy_clear_current),
		present:               (C.BoolCallback)(C.proxy_present),
		fbo_callback:          (C.UIntCallback)(C.proxy_fbo_callback),
		make_resource_current: (C.BoolCallback)(C.proxy_make_resource_current),
	}

	result := C.runFlutter(C.ulong(window), &flu.Engine, &a)
	if flu.Engine == nil {
		return KInvalidArguments
	}

	return (FlutterResult)(result)
}

func (flu *FlutterOpenGL) Shutdown() FlutterResult {
	result := C.FlutterEngineShutdown(flu.Engine)
	return (FlutterResult)(result)
}

type FlutterResult int32

const (
	KSuccess               FlutterResult = C.kSuccess
	KInvalidLibraryVersion FlutterResult = C.kInvalidLibraryVersion
	KInvalidArguments      FlutterResult = C.kInvalidArguments
)

type FlutterPointerPhase int32

const (
	KCancel FlutterPointerPhase = C.kCancel
	KUp     FlutterPointerPhase = C.kUp
	KDown   FlutterPointerPhase = C.kDown
	KMove   FlutterPointerPhase = C.kMove
)

type FlutterPointerEvent struct {
	Phase     FlutterPointerPhase
	Timestamp int64
	X         float64
	Y         float64
}

func FlutterEngineSendPointerEvent(Engine unsafe.Pointer, Event FlutterPointerEvent) FlutterResult {
	cEngine := *(*C.FlutterEngine)(unsafe.Pointer(&Engine))

	cEvents := C.FlutterPointerEvent{
		phase:     (_Ctype_FlutterPointerPhase)(Event.Phase),
		x:         C.double(Event.X),
		y:         C.double(Event.Y),
		timestamp: C.size_t(Event.Timestamp),
	}
	cEvents.struct_size = C.size_t(unsafe.Sizeof(cEvents))

	ret := C.FlutterEngineSendPointerEvent(cEngine, &cEvents, 1)

	return (FlutterResult)(ret)
}

type FlutterWindowMetricsEvent struct {
	Width       int
	Height      int
	Pixel_ratio float64
}

func FlutterEngineSendWindowMetricsEvent(Engine unsafe.Pointer, Metric FlutterWindowMetricsEvent) FlutterResult {
	cEngine := *(*C.FlutterEngine)(unsafe.Pointer(&Engine))

	cMetric := C.FlutterWindowMetricsEvent{
		width:       C.size_t(Metric.Width),
		height:      C.size_t(Metric.Height),
		pixel_ratio: C.double(Metric.Pixel_ratio),
	}
	cMetric.struct_size = C.size_t(unsafe.Sizeof(cMetric))

	ret := C.FlutterEngineSendWindowMetricsEvent(cEngine, &cMetric)

	return (FlutterResult)(ret)
}

type FlutterPlatformMessageResponseHandle C.FlutterPlatformMessageResponseHandle
type FlutterPlatformMessage struct {
	Channel         string
	Message         string
	Response_handle *FlutterPlatformMessageResponseHandle
}

func FlutterEngineSendPlatformMessage(Engine unsafe.Pointer, Message FlutterPlatformMessage) FlutterResult {
	cEngine := *(*C.FlutterEngine)(unsafe.Pointer(&Engine))

	cPlatformMessage := C.FlutterPlatformMessage{
		channel:      C.CString(Message.Channel),
		message:      (*C.uint8_t)(unsafe.Pointer(C.CString(Message.Message))),
		message_size: C.ulong(len(Message.Message)),
	}

	cPlatformMessage.struct_size = C.size_t(unsafe.Sizeof(cPlatformMessage))

	ret := C.FlutterEngineSendPlatformMessage(
		cEngine,
		&cPlatformMessage,
	)

	return (FlutterResult)(ret)
}
