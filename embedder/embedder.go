package embedder

// #include "embedder.h"
// FlutterEngineResult runFlutter(void *user_data, FlutterEngine *engine, FlutterProjectArgs * Args,
//						 const char *const * vmArgs, int nVmAgrs);
// FlutterEngineResult
// createMessageResponseHandle(FlutterEngine engine, void *user_data,
//                             FlutterPlatformMessageResponseHandle **reply);
// char** makeCharArray(int size);
// void setArrayString(char **a, char *s, int n);
// const int32_t kFlutterSemanticsNodeIdBatchEnd = -1;
// const int32_t kFlutterSemanticsCustomActionIdBatchEnd = -1;
import "C"
import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"unsafe"
)

// A list of flutter engines that are managed by this embedder.
var flutterEngines []*FlutterEngine
var flutterEnginesLock sync.RWMutex

// FlutterEngineByIndex returns an existing FlutterEngine by its index in this embedder.
// Deprecated 2019-04-05, this is not used by go-flutter anymore and may be removed in the future.
func FlutterEngineByIndex(index int) (engine *FlutterEngine) {
	fmt.Println("go-flutter: FlutterEngineByIndex(..) is deprecated")
	flutterEnginesLock.RLock()
	if index <= len(flutterEngines)-1 {
		engine = flutterEngines[index]
	}
	flutterEnginesLock.RUnlock()
	if engine == nil {
		fmt.Printf("go-flutter: no flutterEngine found for index %d\n", index)
		debug.PrintStack()
		return nil
	}
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
	ResultEngineNotRunning      Result = -1
)

// FlutterOpenGLTexture corresponds to the C.FlutterOpenGLTexture struct.
type FlutterOpenGLTexture struct {
	// Target texture of the active texture unit (example GL_TEXTURE_2D)
	Target uint32
	// The name of the texture
	Name uint32
	// The texture format (example GL_RGBA8)
	Format uint32
}

// FlutterTask is a type alias to C.FlutterTask
type FlutterTask = C.FlutterTask

// FlutterEngine corresponds to the C.FlutterEngine with his associated callback's method.
type FlutterEngine struct {
	// Flutter Engine.
	Engine C.FlutterEngine

	// closed indicates if the engine has Shutdown
	closed bool
	sync   sync.Mutex
	// index of the engine in the global flutterEngines slice
	index int

	// GL callback functions
	GLMakeCurrent                  func() bool
	GLClearCurrent                 func() bool
	GLPresent                      func() bool
	GLFboCallback                  func() int32
	GLMakeResourceCurrent          func() bool
	GLProcResolver                 func(procName string) unsafe.Pointer
	GLExternalTextureFrameCallback func(textureID int64, width int, height int) *FlutterOpenGLTexture

	// task runner interop
	TaskRunnerRunOnCurrentThread func() bool
	TaskRunnerPostTask           func(trask FlutterTask, targetTimeNanos uint64)

	// platform message callback function
	PlatfromMessage func(message *PlatformMessage)

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
// Deprecated 2019-04-05, this is not used by go-flutter anymore and may be removed in the future.
func (flu *FlutterEngine) Index() int {
	fmt.Println("go-flutter: engine.Index() is deprecated")
	return flu.index
}

// Run launches the Flutter Engine in a background thread.
func (flu *FlutterEngine) Run(userData unsafe.Pointer, vmArgs []string) Result {
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

	res := C.runFlutter(userData, &flu.Engine, &args, cVMArgs, C.int(len(vmArgs)))
	if flu.Engine == nil {
		return ResultInvalidArguments
	}

	return (Result)(res)
}

// Shutdown stops the Flutter engine.
func (flu *FlutterEngine) Shutdown() Result {
	flu.sync.Lock()
	defer flu.sync.Unlock()
	flu.closed = true
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
	PointerPhaseAdd    PointerPhase = C.kAdd
	PointerPhaseRemove PointerPhase = C.kRemove
	PointerPhaseHover  PointerPhase = C.kHover
)

// PointerButtonMouse corresponds to the C.enum describing the mouse buttons.
type PointerButtonMouse int64

// Values representing the mouse buttons.
const (
	PointerButtonMousePrimary   PointerButtonMouse = C.kFlutterPointerButtonMousePrimary
	PointerButtonMouseSecondary PointerButtonMouse = C.kFlutterPointerButtonMouseSecondary
	PointerButtonMouseMiddle    PointerButtonMouse = C.kFlutterPointerButtonMouseMiddle
)

// PointerSignalKind corresponds to the C.enum describing signal kind of the mouse pointer.
type PointerSignalKind int32

// Values representing the pointer signal kind.
const (
	PointerSignalKindNone   PointerSignalKind = C.kFlutterPointerSignalKindNone
	PointerSignalKindScroll PointerSignalKind = C.kFlutterPointerSignalKindScroll
)

// PointerDeviceKind corresponds to the C.enum describing device kind of the mouse pointer.
type PointerDeviceKind int32

// Values representing the pointer signal kind.
const (
	PointerDeviceKindMouse PointerDeviceKind = C.kFlutterPointerDeviceKindMouse
	PointerDeviceKindTouch PointerDeviceKind = C.kFlutterPointerDeviceKindTouch
)

// PointerEvent represents the position and phase of the mouse at a given time.
type PointerEvent struct {
	Phase        PointerPhase
	Timestamp    int64
	X            float64
	Y            float64
	SignalKind   PointerSignalKind
	ScrollDeltaX float64
	ScrollDeltaY float64
	Buttons      PointerButtonMouse
}

// SendPointerEvent is used to send an PointerEvent to the Flutter engine.
func (flu *FlutterEngine) SendPointerEvent(event PointerEvent) Result {
	cPointerEvent := C.FlutterPointerEvent{
		phase:          (C.FlutterPointerPhase)(event.Phase),
		x:              C.double(event.X),
		y:              C.double(event.Y),
		timestamp:      C.size_t(event.Timestamp),
		signal_kind:    (C.FlutterPointerSignalKind)(event.SignalKind),
		device_kind:    (C.FlutterPointerDeviceKind)(PointerDeviceKindMouse),
		scroll_delta_x: C.double(event.ScrollDeltaX),
		scroll_delta_y: C.double(event.ScrollDeltaY),
		buttons:        C.int64_t(event.Buttons),
	}
	cPointerEvent.struct_size = C.size_t(unsafe.Sizeof(cPointerEvent))

	res := C.FlutterEngineSendPointerEvent(flu.Engine, &cPointerEvent, 1)

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
func (flu *FlutterEngine) SendWindowMetricsEvent(event WindowMetricsEvent) Result {
	cMetricEvent := C.FlutterWindowMetricsEvent{
		width:       C.size_t(event.Width),
		height:      C.size_t(event.Height),
		pixel_ratio: C.double(event.PixelRatio),
	}
	cMetricEvent.struct_size = C.size_t(unsafe.Sizeof(cMetricEvent))

	res := C.FlutterEngineSendWindowMetricsEvent(flu.Engine, &cMetricEvent)

	return (Result)(res)
}

// PlatformMessage represents a binary message sent from or to the flutter
// application.
type PlatformMessage struct {
	Channel string
	Message []byte

	// ResponseHandle is set on some received platform message. All
	// PlatformMessage received with this attribute must send a response with
	// `SendPlatformMessageResponse`.
	// ResponseHandle can also be created from the embedder side when a
	// platform(golang) message needs native callback.
	ResponseHandle PlatformMessageResponseHandle
}

// PlatformMessageResponseHandle is a pointer that is used to wire a platform
// message response to the original platform message.
type PlatformMessageResponseHandle uintptr

// ExpectsResponse indicates whether the platform message should receive a
// response.
func (p PlatformMessage) ExpectsResponse() bool {
	return p.ResponseHandle != 0
}

// SendPlatformMessage is used to send a PlatformMessage to the Flutter engine.
func (flu *FlutterEngine) SendPlatformMessage(msg *PlatformMessage) Result {
	flu.sync.Lock()
	defer flu.sync.Unlock()
	if flu.closed {
		return ResultEngineNotRunning
	}

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

// RunTask inform the engine to run the specified task.
func (flu *FlutterEngine) RunTask(task *FlutterTask) Result {
	res := C.FlutterEngineRunTask(flu.Engine, task)
	return (Result)(res)
}

// RegisterExternalTexture registers an external texture with a unique identifier.
func (flu *FlutterEngine) RegisterExternalTexture(textureID int64) Result {
	flu.sync.Lock()
	defer flu.sync.Unlock()
	if flu.closed {
		return ResultEngineNotRunning
	}
	res := C.FlutterEngineRegisterExternalTexture(flu.Engine, C.int64_t(textureID))
	return (Result)(res)
}

// UnregisterExternalTexture unregisters a previous texture registration.
func (flu *FlutterEngine) UnregisterExternalTexture(textureID int64) Result {
	flu.sync.Lock()
	defer flu.sync.Unlock()
	if flu.closed {
		return ResultEngineNotRunning
	}
	res := C.FlutterEngineUnregisterExternalTexture(flu.Engine, C.int64_t(textureID))
	return (Result)(res)
}

// MarkExternalTextureFrameAvailable marks that a new texture frame is
// available for a given texture identifier.
func (flu *FlutterEngine) MarkExternalTextureFrameAvailable(textureID int64) Result {
	flu.sync.Lock()
	defer flu.sync.Unlock()
	if flu.closed {
		return ResultEngineNotRunning
	}
	res := C.FlutterEngineMarkExternalTextureFrameAvailable(flu.Engine, C.int64_t(textureID))
	return (Result)(res)
}

// DataCallback is a function called when a PlatformMessage response send back
// to the embedder.
type DataCallback func(binaryReply []byte)

// CreatePlatformMessageResponseHandle creates a platform message response
// handle that allows the embedder to set a native callback for a response to a
// message.
// Must be collected via `ReleasePlatformMessageResponseHandle` after the call
// to `SendPlatformMessage`.
func (flu *FlutterEngine) CreatePlatformMessageResponseHandle(callback DataCallback) (PlatformMessageResponseHandle, error) {
	var responseHandle *C.FlutterPlatformMessageResponseHandle

	callbackPointer := uintptr(unsafe.Pointer(&callback))
	defer func() {
		runtime.KeepAlive(callbackPointer)
	}()

	res := C.createMessageResponseHandle(flu.Engine, unsafe.Pointer(&callbackPointer), &responseHandle)
	if (Result)(res) != ResultSuccess {
		return 0, errors.New("failed to create a response handle")
	}
	return PlatformMessageResponseHandle(unsafe.Pointer(responseHandle)), nil
}

// ReleasePlatformMessageResponseHandle collects a platform message response
// handle.
func (flu *FlutterEngine) ReleasePlatformMessageResponseHandle(responseHandle PlatformMessageResponseHandle) {
	cResponseHandle := (*C.FlutterPlatformMessageResponseHandle)(unsafe.Pointer(responseHandle))
	res := C.FlutterPlatformMessageReleaseResponseHandle(flu.Engine, cResponseHandle)
	if (Result)(res) != ResultSuccess {
		fmt.Printf("go-flutter: failed to collect platform response message handle")
	}
}

// FlutterEngineGetCurrentTime gets the current time in nanoseconds from the clock used by the flutter
// engine.
func FlutterEngineGetCurrentTime() uint64 {
	return uint64(C.FlutterEngineGetCurrentTime())
}
