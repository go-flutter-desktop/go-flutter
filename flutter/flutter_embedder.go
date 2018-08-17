package flutter

/*
#cgo CFLAGS: -I${SRCDIR}/library
#cgo LDFLAGS: -L${SRCDIR}/library/linux -lflutter_engine -Wl,-rpath,$ORIGIN/library/linux
#include "flutter_embedder.h"

*/
import "C"
import (
	"unsafe"
)

type OpenGLFlutterEngine struct {
	engine               C.FlutterEngine
	FMakeCurrent         func(v unsafe.Pointer) bool
	FClearCurrent        func(v unsafe.Pointer) bool
	FPresent             func(v unsafe.Pointer) bool
	FFboCallback         func(v unsafe.Pointer) int32
	FMakeResourceCurrent func(v unsafe.Pointer) bool
}

type FlutterResult int32

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
