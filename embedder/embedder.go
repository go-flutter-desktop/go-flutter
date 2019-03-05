package embedder

// #include "flutter_embedder.h"
// FlutterEngineResult runFlutter(uintptr_t window, FlutterEngine *engine, FlutterProjectArgs * Args,
//						 const char *const * vmArgs, int nVmAgrs);
// char** makeCharArray(int size);
// void setArrayString(char **a, char *s, int n);
import "C"
import (
	"sync"
	"unsafe"
)

// A list of flutter engines that are managed by this embedder.
var flutterEngines []*FlutterEngine
var flutterEnginesLock sync.RWMutex

// FlutterEngineByIndex returns an existing FlutterEngine by its index in this embedder.
func FlutterEngineByIndex(index int) (engine *FlutterEngine) {
	flutterEnginesLock.RLock()
	// TODO(#89): Remove this workarround check on slice length to handle
	// for invalid indexes that are sometimes returned by glfw proxies because
	// they cannot guarantee to execute glfw functions on the main thread.
	if index <= len(flutterEngines)-1 {
		engine = flutterEngines[index]
	}
	flutterEnginesLock.RUnlock()
	return engine
}

// CountFlutterEngines return the number of engines registered in this embedder.
func CountFlutterEngines() int {
	return len(flutterEngines)
}

// Result corresponds to the C.enum retuned by the shared flutter library
// whenever we call it.
type Result int32

// Values representing the status of an Result.
const (
	ResultSuccess               Result = C.kSuccess
	ResultInvalidLibraryVersion Result = C.kInvalidLibraryVersion
	ResultInvalidArguments      Result = C.kInvalidArguments
)

// FlutterEngine corresponds to the C.FlutterEngine with his associated callback's method.
type FlutterEngine struct {
	// Flutter Engine.
	Engine C.FlutterEngine

	// index of the engine in the global flutterEngines slice
	index int

	// Necessary callbacks for rendering.
	FMakeCurrent         func(v unsafe.Pointer) bool
	FClearCurrent        func(v unsafe.Pointer) bool
	FPresent             func(v unsafe.Pointer) bool
	FFboCallback         func(v unsafe.Pointer) int32
	FMakeResourceCurrent func(v unsafe.Pointer) bool

	// platform message callback.
	FPlatfromMessage func(message *PlatformMessage)

	// Engine arguments
	AssetsPath  string
	IcuDataPath string
}

// NewFlutterEngine creates an empty FlutterEngine
// and assigns it an index for global lookup.
func NewFlutterEngine() *FlutterEngine {
	fe := &FlutterEngine{}
	flutterEnginesLock.Lock()
	flutterEngines = append(flutterEngines, fe)
	fe.index = len(flutterEngines) - 1
	flutterEnginesLock.Unlock()
	return fe
}

// Index returns the index of the engine in the global flutterEngines slice
func (flu *FlutterEngine) Index() int {
	return flu.index
}

// Run launches the Flutter Engine in a background thread.
func (flu *FlutterEngine) Run(window uintptr, vmArgs []string) Result {
	// validate this FlutterEngine was created correctly
	flutterEnginesLock.RLock()
	if len(flutterEngines) <= flu.index || flutterEngines[flu.index] != flu {
		panic("FlutterEngine was wrongly created. Use embedder.NewFlutterEngine().")
	}
	flutterEnginesLock.RUnlock()

	args := C.FlutterProjectArgs{
		assets_path:   C.CString(flu.AssetsPath),
		icu_data_path: C.CString(flu.IcuDataPath),
	}

	args.struct_size = C.size_t(unsafe.Sizeof(args))

	cVMArgs := C.makeCharArray(C.int(len(vmArgs)))
	for i, s := range vmArgs {
		C.setArrayString(cVMArgs, C.CString(s), C.int(i))
	}

	res := C.runFlutter(C.uintptr_t(window), &flu.Engine, &args, cVMArgs, C.int(len(vmArgs)))
	if flu.Engine == nil {
		return ResultInvalidArguments
	}

	return (Result)(res)
}

// Shutdown stops the Flutter engine.
func (flu *FlutterEngine) Shutdown() Result {
	res := C.FlutterEngineShutdown(flu.Engine)
	return (Result)(res)
}

// PointerPhase corresponds to the C.enum describing phase of the mouse pointer.
type PointerPhase int32

// Values representing the mouse phase.
const (
	PointerPhaseCancel PointerPhase = C.kCancel
	PointerPhaseUp     PointerPhase = C.kUp
	PointerPhaseDown   PointerPhase = C.kDown
	PointerPhaseMove   PointerPhase = C.kMove
)

// PointerEvent represents the position and phase of the mouse at a given time.
type PointerEvent struct {
	Phase     PointerPhase
	Timestamp int64
	X         float64
	Y         float64
}

// SendPointerEvent is used to send an PointerEvent to the Flutter engine.
func (flu *FlutterEngine) SendPointerEvent(Event PointerEvent) Result {

	cEvents := C.FlutterPointerEvent{
		phase:     (C.FlutterPointerPhase)(Event.Phase),
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

// SendWindowMetricsEvent is used to send a WindowMetricsEvent to the Flutter
// Engine.
func (flu *FlutterEngine) SendWindowMetricsEvent(Metric WindowMetricsEvent) Result {
	cMetric := C.FlutterWindowMetricsEvent{
		width:       C.size_t(Metric.Width),
		height:      C.size_t(Metric.Height),
		pixel_ratio: C.double(Metric.PixelRatio),
	}
	cMetric.struct_size = C.size_t(unsafe.Sizeof(cMetric))

	res := C.FlutterEngineSendWindowMetricsEvent(flu.Engine, &cMetric)

	return (Result)(res)
}

// PlatformMessage represents a binary message sent from or to the flutter
// application.
type PlatformMessage struct {
	Channel string
	Message []byte

	// ResponseHandle is only set when receiving a platform message. // TODO: is comment true?
	ResponseHandle PlatformMessageResponseHandle
}

type PlatformMessageResponseHandle uintptr

func (p PlatformMessage) ExpectsReply() bool {
	return p.ResponseHandle != 0
}

// SendPlatformMessage is used to send a PlatformMessage to the Flutter engine.
func (flu *FlutterEngine) SendPlatformMessage(msg *PlatformMessage) Result {
	cPlatformMessage := C.FlutterPlatformMessage{
		channel: C.CString(msg.Channel),
		// TODO: who is responsible for free-ing this C alloc? And can they be
		// freed when this call returns? Or are they stil used at that time?
		message:      (*C.uint8_t)(C.CBytes(msg.Message)),
		message_size: C.size_t(len(msg.Message)),

		response_handle: (*C.FlutterPlatformMessageResponseHandle)(unsafe.Pointer(msg.ResponseHandle)),
	}

	cPlatformMessage.struct_size = C.size_t(unsafe.Sizeof(cPlatformMessage))

	res := C.FlutterEngineSendPlatformMessage(
		flu.Engine,
		&cPlatformMessage,
	)

	return (Result)(res)
}

// SendPlatformMessageResponse is used to send a message to the Flutter side
// using the correct ResponseHandle.
func (flu *FlutterEngine) SendPlatformMessageResponse(
	responseTo PlatformMessageResponseHandle,
	encodedMessage []byte,
) Result {
	res := C.FlutterEngineSendPlatformMessageResponse(
		flu.Engine,
		(*C.FlutterPlatformMessageResponseHandle)(unsafe.Pointer(responseTo)),
		// TODO: who is responsible for free-ing this C alloc? And can they be
		// freed when this call returns? Or are they stil used at that time?
		(*C.uint8_t)(C.CBytes(encodedMessage)),
		(C.size_t)(len(encodedMessage)),
	)

	return (Result)(res)
}

// FlutterEngineFlushPendingTasksNow flush tasks on a  message loop not
// controlled by the Flutter engine.
//
// deprecated soon.
func FlutterEngineFlushPendingTasksNow() {
	C.__FlutterEngineFlushPendingTasksNow()
}
