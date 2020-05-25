package embedder

import "C"

// #include "embedder.h"
// #include <stdlib.h>
// FlutterEngineResult runFlutter(void *user_data, FlutterEngine *engine, FlutterProjectArgs * Args);
// FlutterEngineResult
// createMessageResponseHandle(FlutterEngine engine, void *user_data,
//                             FlutterPlatformMessageResponseHandle **reply);
// const int32_t kFlutterSemanticsNodeIdBatchEnd = -1;
// const int32_t kFlutterSemanticsCustomActionIdBatchEnd = -1;
// FlutterEngineAOTDataSource* createAOTDataSource(FlutterEngineAOTDataSource *data_in, const char * elfSnapshotPath);
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/pkg/errors"
)

// Result corresponds to the C.enum retuned by the shared flutter library
// whenever we call it.
type Result int32

// Values representing the status of an Result.
const (
	ResultSuccess               Result = C.kSuccess
	ResultInvalidLibraryVersion Result = C.kInvalidLibraryVersion
	ResultInvalidArguments      Result = C.kInvalidArguments
	ResultInternalInconsistency Result = C.kInternalInconsistency
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

	// AOT ELF snopshot path
	// only required for AOT app.
	ElfSnapshotPath string
	aotDataSource   C.FlutterEngineAOTData
}

// GoError convert a FlutterEngineResult to a golang readable error
func (res Result) GoError(caller string) error {
	switch res {
	case ResultSuccess:
		return nil
	case ResultInvalidLibraryVersion:
		return errors.Errorf("%s returned result code %d (invalid library version)", caller, res)
	case ResultInvalidArguments:
		return errors.Errorf("%s returned result code %d (invalid arguments)", caller, res)
	case ResultInternalInconsistency:
		return errors.Errorf("%s returned result code %d (internal inconsistency)", caller, res)
	case ResultEngineNotRunning:
		return errors.Errorf("%s returned result code %d (engine not running)", caller, res)
	default:
		return errors.Errorf("%s returned result code %d (unknown result code)", caller, res)
	}
}

// NewFlutterEngine creates an empty FlutterEngine.
func NewFlutterEngine() *FlutterEngine {
	return &FlutterEngine{}
}

// Run launches the Flutter Engine in a background thread.
func (flu *FlutterEngine) Run(userData unsafe.Pointer, vmArgs []string) error {

	cVMArgs := C.malloc(C.size_t(len(vmArgs)) * C.size_t(unsafe.Sizeof(uintptr(0))))
	defer C.free(cVMArgs)

	a := (*[1<<30 - 1]*C.char)(cVMArgs)

	for idx, substring := range vmArgs {
		a[idx] = C.CString(substring)
		defer C.free(unsafe.Pointer(a[idx]))
	}

	assetsPath := C.CString(flu.AssetsPath)
	icuDataPath := C.CString(flu.IcuDataPath)
	defer C.free(unsafe.Pointer(assetsPath))
	defer C.free(unsafe.Pointer(icuDataPath))

	args := C.FlutterProjectArgs{
		assets_path:                assetsPath,
		icu_data_path:              icuDataPath,
		command_line_argv:          (**C.char)(cVMArgs),
		command_line_argc:          C.int(len(vmArgs)),
		shutdown_dart_vm_when_done: true,
	}

	if C.FlutterEngineRunsAOTCompiledDartCode() {
		// elfSnapshotPath := C.CString(flu.ElfSnapshotPath)
		// defer C.free(unsafe.Pointer(elfSnapshotPath))

		// dataIn := C.FlutterEngineAOTDataSource{}

		// C.createAOTDataSource(&dataIn, elfSnapshotPath)
		// res := (Result)(C.FlutterEngineCreateAOTData(&dataIn, &flu.aotDataSource))
		// if res != ResultSuccess {
		// return res.GoError("C.FlutterEngineCreateAOTData()")
		// }
		// args.aot_data = flu.aotDataSource
	}

	args.struct_size = C.size_t(unsafe.Sizeof(args))

	res := (Result)(C.runFlutter(userData, &flu.Engine, &args))
	if flu.Engine == nil {
		return ResultInvalidArguments.GoError("engine.Run()")
	}

	return res.GoError("engine.Run()")
}

// Shutdown stops the Flutter engine.
func (flu *FlutterEngine) Shutdown() error {
	flu.sync.Lock()
	defer flu.sync.Unlock()
	flu.closed = true

	res := (Result)(C.FlutterEngineShutdown(flu.Engine))
	if res != ResultSuccess {
		return res.GoError("engine.Shutdown()")
	}

	if C.FlutterEngineRunsAOTCompiledDartCode() {
		// res := (Result)(C.FlutterEngineCollectAOTData(flu.aotDataSource))
		// if res != ResultSuccess {
		// return res.GoError("engine.Shutdown()")
		// }
	}
	return nil
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
	X            float64
	Y            float64
	SignalKind   PointerSignalKind
	ScrollDeltaX float64
	ScrollDeltaY float64
	Buttons      PointerButtonMouse
}

// SendPointerEvent is used to send an PointerEvent to the Flutter engine.
func (flu *FlutterEngine) SendPointerEvent(event PointerEvent) error {
	cPointerEvent := C.FlutterPointerEvent{
		phase:          (C.FlutterPointerPhase)(event.Phase),
		x:              C.double(event.X),
		y:              C.double(event.Y),
		timestamp:      C.size_t(FlutterEngineGetCurrentTime()),
		signal_kind:    (C.FlutterPointerSignalKind)(event.SignalKind),
		device_kind:    (C.FlutterPointerDeviceKind)(PointerDeviceKindMouse),
		scroll_delta_x: C.double(event.ScrollDeltaX),
		scroll_delta_y: C.double(event.ScrollDeltaY),
		buttons:        C.int64_t(event.Buttons),
	}
	cPointerEvent.struct_size = C.size_t(unsafe.Sizeof(cPointerEvent))

	res := C.FlutterEngineSendPointerEvent(flu.Engine, &cPointerEvent, 1)

	return (Result)(res).GoError("engine.SendPointerEvent")
}

// WindowMetricsEvent represents a window's resolution.
type WindowMetricsEvent struct {
	Width      int
	Height     int
	PixelRatio float64
}

// SendWindowMetricsEvent is used to send a WindowMetricsEvent to the Flutter
// Engine.
func (flu *FlutterEngine) SendWindowMetricsEvent(event WindowMetricsEvent) error {
	cMetricEvent := C.FlutterWindowMetricsEvent{
		width:       C.size_t(event.Width),
		height:      C.size_t(event.Height),
		pixel_ratio: C.double(event.PixelRatio),
	}
	cMetricEvent.struct_size = C.size_t(unsafe.Sizeof(cMetricEvent))

	res := C.FlutterEngineSendWindowMetricsEvent(flu.Engine, &cMetricEvent)

	return (Result)(res).GoError("engine.SendWindowMetricsEvent()")
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

// UpdateSystemLocale is used to update the flutter locale to the system locale
func (flu *FlutterEngine) UpdateSystemLocale(lang, country, script string) error {
	languageCode := C.CString(lang)
	countryCode := C.CString(country)
	scriptCode := C.CString(script)
	defer C.free(unsafe.Pointer(languageCode))
	defer C.free(unsafe.Pointer(countryCode))
	defer C.free(unsafe.Pointer(scriptCode))

	locale := C.FlutterLocale{
		language_code: languageCode,
		country_code:  countryCode,
		script_code:   scriptCode,
	}

	locale.struct_size = C.size_t(unsafe.Sizeof(locale))
	arr := (**C.FlutterLocale)(C.malloc(locale.struct_size))
	defer C.free(unsafe.Pointer(arr))
	(*[1]*C.FlutterLocale)(unsafe.Pointer(arr))[0] = &locale

	res := C.FlutterEngineUpdateLocales(flu.Engine, arr, (C.size_t)(1))

	return (Result)(res).GoError("engine.UpdateSystemLocale()")
}

// SendPlatformMessage is used to send a PlatformMessage to the Flutter engine.
func (flu *FlutterEngine) SendPlatformMessage(msg *PlatformMessage) error {
	flu.sync.Lock()
	defer flu.sync.Unlock()
	if flu.closed {
		return ResultEngineNotRunning.GoError("engine.SendPlatformMessage()")
	}

	message := C.CBytes(msg.Message)
	channel := C.CString(msg.Channel)
	defer C.free(message)
	defer C.free(unsafe.Pointer(channel))

	cPlatformMessage := C.FlutterPlatformMessage{
		channel:      channel,
		message:      (*C.uint8_t)(message),
		message_size: C.size_t(len(msg.Message)),

		response_handle: (*C.FlutterPlatformMessageResponseHandle)(unsafe.Pointer(msg.ResponseHandle)),
	}

	cPlatformMessage.struct_size = C.size_t(unsafe.Sizeof(cPlatformMessage))

	res := C.FlutterEngineSendPlatformMessage(
		flu.Engine,
		&cPlatformMessage,
	)

	return (Result)(res).GoError("engine.SendPlatformMessage()")
}

// SendPlatformMessageResponse is used to send a message to the Flutter side
// using the correct ResponseHandle.
func (flu *FlutterEngine) SendPlatformMessageResponse(
	responseTo PlatformMessageResponseHandle,
	encodedMessage []byte,
) error {

	message := C.CBytes(encodedMessage)
	defer C.free(message)

	res := C.FlutterEngineSendPlatformMessageResponse(
		flu.Engine,
		(*C.FlutterPlatformMessageResponseHandle)(unsafe.Pointer(responseTo)),
		(*C.uint8_t)(message),
		(C.size_t)(len(encodedMessage)),
	)

	return (Result)(res).GoError("engine.SendPlatformMessageResponse()")
}

// RunTask inform the engine to run the specified task.
func (flu *FlutterEngine) RunTask(task *FlutterTask) error {
	res := C.FlutterEngineRunTask(flu.Engine, task)
	return (Result)(res).GoError("engine.RunTask()")
}

// RegisterExternalTexture registers an external texture with a unique identifier.
func (flu *FlutterEngine) RegisterExternalTexture(textureID int64) error {
	flu.sync.Lock()
	defer flu.sync.Unlock()
	if flu.closed {
		return ResultEngineNotRunning.GoError("engine.RegisterExternalTexture()")
	}
	res := C.FlutterEngineRegisterExternalTexture(flu.Engine, C.int64_t(textureID))
	return (Result)(res).GoError("engine.RegisterExternalTexture()")
}

// UnregisterExternalTexture unregisters a previous texture registration.
func (flu *FlutterEngine) UnregisterExternalTexture(textureID int64) error {
	flu.sync.Lock()
	defer flu.sync.Unlock()
	if flu.closed {
		return ResultEngineNotRunning.GoError("engine.UnregisterExternalTexture()")
	}
	res := C.FlutterEngineUnregisterExternalTexture(flu.Engine, C.int64_t(textureID))
	return (Result)(res).GoError("engine.UnregisterExternalTexture()")
}

// MarkExternalTextureFrameAvailable marks that a new texture frame is
// available for a given texture identifier.
func (flu *FlutterEngine) MarkExternalTextureFrameAvailable(textureID int64) error {
	flu.sync.Lock()
	defer flu.sync.Unlock()
	if flu.closed {
		return ResultEngineNotRunning.GoError("engine.MarkExternalTextureFrameAvailable()")
	}
	res := C.FlutterEngineMarkExternalTextureFrameAvailable(flu.Engine, C.int64_t(textureID))
	return (Result)(res).GoError("engine.MarkExternalTextureFrameAvailable()")
}

// DataCallback is a function called when a PlatformMessage response send back
// to the embedder.
type DataCallback struct {
	// Handle func
	Handle func(binaryReply []byte)
}

// CreatePlatformMessageResponseHandle creates a platform message response
// handle that allows the embedder to set a native callback for a response to a
// message.
// Must be collected via `ReleasePlatformMessageResponseHandle` after the call
// to `SendPlatformMessage`.
func (flu *FlutterEngine) CreatePlatformMessageResponseHandle(callback *DataCallback) (PlatformMessageResponseHandle, error) {
	var responseHandle *C.FlutterPlatformMessageResponseHandle

	callbackPointer := uintptr(unsafe.Pointer(callback))
	defer func() {
		runtime.KeepAlive(callbackPointer)
	}()

	res := C.createMessageResponseHandle(flu.Engine, unsafe.Pointer(&callbackPointer), &responseHandle)
	return PlatformMessageResponseHandle(unsafe.Pointer(responseHandle)), (Result)(res).GoError("engine.CreatePlatformMessageResponseHandle()")
}

// ReleasePlatformMessageResponseHandle collects a platform message response
// handle.
func (flu *FlutterEngine) ReleasePlatformMessageResponseHandle(responseHandle PlatformMessageResponseHandle) {
	cResponseHandle := (*C.FlutterPlatformMessageResponseHandle)(unsafe.Pointer(responseHandle))
	res := C.FlutterPlatformMessageReleaseResponseHandle(flu.Engine, cResponseHandle)
	if (Result)(res) != ResultSuccess {
		fmt.Printf("go-flutter: failed to collect platform response message handle\n")
	}
}

// FlutterEngineGetCurrentTime gets the current time in nanoseconds from the clock used by the flutter
// engine.
func FlutterEngineGetCurrentTime() uint64 {
	return uint64(C.FlutterEngineGetCurrentTime())
}
