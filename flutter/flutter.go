package flutter

// #include "flutter_embedder.h"
// FlutterResult runFlutter(uintptr_t window, FlutterEngine *engine, FlutterProjectArgs * Args,
//						 const char *const * vmArgs, int nVmAgrs);
// char** makeCharArray(int size);
// void setArrayString(char **a, char *s, int n);
import "C"
import (
	"encoding/json"
	"unsafe"
)

// the current FlutterEngine running (associated with his callback)
var flutterEngines []*EngineOpenGL

// SelectEngine return a EngineOpenGL from an index
func SelectEngine(index int) *EngineOpenGL {
	return flutterEngines[index]
}

// NumberOfEngines return the number of engine registered into this embedder
func NumberOfEngines() C.int {
	return C.int(len(flutterEngines))
}

// Result corresponds to the C.enum retuned by the shared flutter library
// whenever we call it.
type Result int32

// Values representing the status of an Result.
const (
	KSuccess               Result = C.kSuccess
	KInvalidLibraryVersion Result = C.kInvalidLibraryVersion
	KInvalidArguments      Result = C.kInvalidArguments
)

// EngineOpenGL corresponds to the C.FlutterEngine with his associated callback's method.
type EngineOpenGL struct {
	// Flutter Engine.
	Engine C.FlutterEngine

	// Necessary callbacks for rendering.
	FMakeCurrent         func(v unsafe.Pointer) bool
	FClearCurrent        func(v unsafe.Pointer) bool
	FPresent             func(v unsafe.Pointer) bool
	FFboCallback         func(v unsafe.Pointer) int32
	FMakeResourceCurrent func(v unsafe.Pointer) bool

	// platform message callback.
	FPlatfromMessage func(message *PlatformMessage, window unsafe.Pointer) bool

	// Engine arguments
	PixelRatio  float64
	AssetsPath  string
	IcuDataPath string
}

// Run launches the Flutter Engine in a background thread.
func (flu *EngineOpenGL) Run(window uintptr, vmArgs []string) Result {

	flutterEngines = append(flutterEngines, flu)

	args := C.FlutterProjectArgs{
		assets_path:   C.CString(flu.AssetsPath),
		main_path:     C.CString(""),
		packages_path: C.CString(""),
		icu_data_path: C.CString(flu.IcuDataPath),
	}

	args.struct_size = C.size_t(unsafe.Sizeof(args))

	cVMArgs := C.makeCharArray(C.int(len(vmArgs)))
	for i, s := range vmArgs {
		C.setArrayString(cVMArgs, C.CString(s), C.int(i))
	}

	res := C.runFlutter(C.uintptr_t(window), &flu.Engine, &args, cVMArgs, C.int(len(vmArgs)))
	if flu.Engine == nil {
		return KInvalidArguments
	}

	return (Result)(res)
}

// Shutdown stops the Flutter engine.
func (flu *EngineOpenGL) Shutdown() Result {
	res := C.FlutterEngineShutdown(flu.Engine)
	return (Result)(res)
}

// PointerPhase corresponds to the C.enum describing phase of the mouse pointer.
type PointerPhase int32

// Values representing the mouse phase.
const (
	KCancel PointerPhase = C.kCancel
	KUp     PointerPhase = C.kUp
	KDown   PointerPhase = C.kDown
	KMove   PointerPhase = C.kMove
)

// PointerEvent represents the position and phase of the mouse at a given time.
type PointerEvent struct {
	Phase     PointerPhase
	Timestamp int64
	X         float64
	Y         float64
}

// EngineSendPointerEvent is used to send an PointerEvent to the Flutter engine.
func (flu *EngineOpenGL) EngineSendPointerEvent(Event PointerEvent) Result {

	cEvents := C.FlutterPointerEvent{
		phase:     (_Ctype_FlutterPointerPhase)(Event.Phase),
		x:         C.double(Event.X),
		y:         C.double(Event.Y),
		timestamp: C.size_t(Event.Timestamp),
	}
	cEvents.struct_size = C.size_t(unsafe.Sizeof(cEvents))

	res := C.FlutterEngineSendPointerEvent(flu.Engine, &cEvents, 1)

	return (Result)(res)
}

// WindowMetricsEvent represents a window's resolution.
type WindowMetricsEvent struct {
	Width      int
	Height     int
	PixelRatio float64
}

// EngineSendWindowMetricsEvent is used to send a WindowMetricsEvent to the Flutter Engine.
func (flu *EngineOpenGL) EngineSendWindowMetricsEvent(Metric WindowMetricsEvent) Result {

	cMetric := C.FlutterWindowMetricsEvent{
		width:       C.size_t(Metric.Width),
		height:      C.size_t(Metric.Height),
		pixel_ratio: C.double(Metric.PixelRatio),
	}
	cMetric.struct_size = C.size_t(unsafe.Sizeof(cMetric))

	res := C.FlutterEngineSendWindowMetricsEvent(flu.Engine, &cMetric)

	return (Result)(res)
}

// PlatformMessage represents a `MethodChannel` serialized with the `JSONMethodCodec`
// TODO Support for `StandardMethodCodec`
type PlatformMessage struct {
	Channel        string
	Message        Message
	ResponseHandle *C.FlutterPlatformMessageResponseHandle
}

// Message is the json content of a PlatformMessage
type Message struct {
	// Describe the method
	Method string `json:"method"`
	// Actual datas
	Args json.RawMessage `json:"args"`
}

// SendPlatformMessage is used to send a PlatformMessage to the Flutter engine.
func (flu *EngineOpenGL) SendPlatformMessage(Message *PlatformMessage) Result {

	marshalled, err := json.Marshal(&Message.Message)
	if err != nil {
		panic("Cound not send a message to the flutter engine: Error while creating the JSON")
	}
	strMessage := string(marshalled)

	cPlatformMessage := C.FlutterPlatformMessage{
		channel:      C.CString(Message.Channel),
		message:      (*C.uint8_t)(unsafe.Pointer(C.CString(strMessage))),
		message_size: C.uintptr_t(len(strMessage)),
	}

	cPlatformMessage.struct_size = C.size_t(unsafe.Sizeof(cPlatformMessage))

	res := C.FlutterEngineSendPlatformMessage(
		flu.Engine,
		&cPlatformMessage,
	)

	return (Result)(res)
}

// SendPlatformMessageResponse is used to send a message to the Flutter side using the correct ResponseHandle!
func (flu *EngineOpenGL) SendPlatformMessageResponse(
	responseTo *PlatformMessage,
	data []byte,
) Result {

	res := C.FlutterEngineSendPlatformMessageResponse(
		flu.Engine,
		(*C.FlutterPlatformMessageResponseHandle)(responseTo.ResponseHandle),
		(*C.uint8_t)(unsafe.Pointer(C.CBytes(data))),
		(C.size_t)(len(data)))

	return (Result)(res)

}

// EngineFlushPendingTasksNow flush tasks on a  message loop not controlled by the Flutter engine.
// deprecated soon.
func EngineFlushPendingTasksNow() {
	C.__FlutterEngineFlushPendingTasksNow()
}
