package flutter

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/go-flutter-desktop/go-flutter/embedder"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// dpPerInch defines the amount of display pixels per inch as defined for Flutter.
const dpPerInch = 160.0

// TODO (GeertJohan): better name for this, confusing with 'actual' window
// managers. Renderer interface? implemented by this type for glfw? type
// glfwRenderer or glfwManager? All the attaching to glfw.Window must be done
// during manager init in that case. Cannot be done by Application.
type windowManager struct {
	forcedPixelRatio          float64
	oncePrintPixelRatioLimit  sync.Once
	pointerPhase              embedder.PointerPhase
	pixelsPerScreenCoordinate float64
	pointerCurrentlyAdded     bool
}

func newWindowManager() *windowManager {
	return &windowManager{
		pointerPhase: embedder.PointerPhaseHover,
	}
}

func (m *windowManager) sendPointerEvent(window *glfw.Window, phase embedder.PointerPhase, x, y float64) {
	// synthesize an PointerPhaseAdd if the pointer isn't already added
	if !m.pointerCurrentlyAdded && phase != embedder.PointerPhaseAdd {
		m.sendPointerEvent(window, embedder.PointerPhaseAdd, x, y)
	}

	// Don't double-add the pointer
	if m.pointerCurrentlyAdded && phase == embedder.PointerPhaseAdd {
		return
	}

	// TODO(GeertJohan): sometimes the x and/or y given by glfw is negative or over window size, could this cause an issue?
	// spew.Dump(event)
	event := embedder.PointerEvent{
		Phase:     phase,
		X:         x * m.pixelsPerScreenCoordinate,
		Y:         y * m.pixelsPerScreenCoordinate,
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
	}

	flutterEnginePointer := *(*uintptr)(window.GetUserPointer())
	flutterEngine := (*embedder.FlutterEngine)(unsafe.Pointer(flutterEnginePointer))

	flutterEngine.SendPointerEvent(event)

	if phase == embedder.PointerPhaseAdd {
		m.pointerCurrentlyAdded = true
	} else if phase == embedder.PointerPhaseRemove {
		m.pointerCurrentlyAdded = false
	}
}

func (m *windowManager) sendPointerEventButton(window *glfw.Window, phase embedder.PointerPhase, buttons embedder.PointerButtonMouse) {
	x, y := window.GetCursorPos()
	event := embedder.PointerEvent{
		Phase:      phase,
		X:          x * m.pixelsPerScreenCoordinate,
		Y:          y * m.pixelsPerScreenCoordinate,
		Timestamp:  time.Now().UnixNano() / int64(time.Millisecond),
		SignalKind: embedder.PointerSignalKindNone,
		Buttons:    buttons,
	}

	flutterEnginePointer := *(*uintptr)(window.GetUserPointer())
	flutterEngine := (*embedder.FlutterEngine)(unsafe.Pointer(flutterEnginePointer))

	flutterEngine.SendPointerEvent(event)
}

func (m *windowManager) sendPointerEventScroll(window *glfw.Window, xDelta, yDelta float64) {
	x, y := window.GetCursorPos()
	event := embedder.PointerEvent{
		Phase:        m.pointerPhase,
		X:            x * m.pixelsPerScreenCoordinate,
		Y:            y * m.pixelsPerScreenCoordinate,
		Timestamp:    time.Now().UnixNano() / int64(time.Millisecond),
		SignalKind:   embedder.PointerSignalKindScroll,
		ScrollDeltaX: xDelta,
		ScrollDeltaY: yDelta,
	}

	flutterEnginePointer := *(*uintptr)(window.GetUserPointer())
	flutterEngine := (*embedder.FlutterEngine)(unsafe.Pointer(flutterEnginePointer))

	flutterEngine.SendPointerEvent(event)
}

func (m *windowManager) glfwCursorEnterCallback(window *glfw.Window, entered bool) {
	x, y := window.GetCursorPos()
	if entered {
		m.sendPointerEvent(window, embedder.PointerPhaseAdd, x, y)
		m.pointerPhase = embedder.PointerPhaseHover
	} else {
		m.sendPointerEvent(window, embedder.PointerPhaseRemove, x, y)
	}
}

func (m *windowManager) glfwCursorPosCallback(window *glfw.Window, x, y float64) {
	m.sendPointerEvent(window, m.pointerPhase, x, y)
}

func (m *windowManager) sendButtonEvent(window *glfw.Window, action glfw.Action, buttons embedder.PointerButtonMouse) {
	if action == glfw.Press {
		m.sendPointerEventButton(window, embedder.PointerPhaseDown, buttons)
		m.pointerPhase = embedder.PointerPhaseMove
	}

	if action == glfw.Release {
		m.sendPointerEventButton(window, embedder.PointerPhaseUp, buttons)
		m.pointerPhase = embedder.PointerPhaseHover
	}
}

func (m *windowManager) glfwMouseButtonCallback(window *glfw.Window, key glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	switch key {
	case glfw.MouseButtonLeft:
		m.sendButtonEvent(window, action, embedder.PointerButtonMousePrimary)
	case glfw.MouseButtonRight:
		m.sendButtonEvent(window, action, embedder.PointerButtonMouseSecondary)
	case glfw.MouseButtonMiddle:
		m.sendButtonEvent(window, action, embedder.PointerButtonMouseMiddle)
	default:
		m.sendButtonEvent(window, action, 1<<uint(key))
	}
}

func (m *windowManager) glfwScrollCallback(window *glfw.Window, xoff float64, yoff float64) {
	const scrollModifier = -50
	m.sendPointerEventScroll(window, xoff*scrollModifier, yoff*scrollModifier)
}

// glfwRefreshCallback is called when the window needs a reresh, this
// can occur when the window is resized, was covered by another window, etc.
// When forcedPixelratio is zero, the forcedPixelratio communicated to the
// Flutter embedder is calculated based on physical and logical screen
// dimensions.
func (m *windowManager) glfwRefreshCallback(window *glfw.Window) {
	widthPx, heightPx := window.GetFramebufferSize()
	var pixelRatio float64
	if m.forcedPixelRatio != 0 {
		pixelRatio = m.forcedPixelRatio
	} else {
		// TODO(#74): multi-monitor support
		primaryMonitor := glfw.GetPrimaryMonitor()
		if primaryMonitor == nil {
			pixelRatio = 1.0
			goto SendWindowMetricsEvent
		}
		primaryMonitorMode := primaryMonitor.GetVideoMode()
		primaryMonitor.GetVideoModes()
		if primaryMonitorMode == nil {
			pixelRatio = 1.0
			goto SendWindowMetricsEvent
		}
		primaryMonitorWidthMM, _ := primaryMonitor.GetPhysicalSize()
		if primaryMonitorWidthMM == 0 {
			pixelRatio = 1.0
			goto SendWindowMetricsEvent
		}
		monitorScreenCoordinatesPerInch := float64(primaryMonitorMode.Width) / (float64(primaryMonitorWidthMM) / 25.4)

		width, _ := window.GetSize()
		if width == 0 {
			fmt.Println("go-flutter: Cannot calculate pixelsPerScreenCoordinate for zero-width window.")
			pixelRatio = 1.0
			goto SendWindowMetricsEvent
		}

		m.pixelsPerScreenCoordinate = float64(widthPx) / float64(width)
		dpi := m.pixelsPerScreenCoordinate * monitorScreenCoordinatesPerInch
		pixelRatio = dpi / dpPerInch

		// Limit the ratio to 1 to avoid rendering a smaller UI in standard resolution monitors.
		if pixelRatio < 1.0 {
			metrics := map[string]interface{}{
				"ppsc":           m.pixelsPerScreenCoordinate,
				"windowWidthPx":  widthPx,
				"windowWidthSc":  width,
				"mscpi":          monitorScreenCoordinatesPerInch,
				"dpi":            dpi,
				"pixelRatio":     pixelRatio,
				"monitorWidthMm": primaryMonitorWidthMM,
				"monitorWidthSc": primaryMonitorMode.Width,
			}
			pixelRatio = 1.0
			m.oncePrintPixelRatioLimit.Do(func() {
				metricsBytes, _ := json.Marshal(metrics)
				fmt.Println("go-flutter: calculated pixelRatio limited to a minimum of 1.0. metrics: " + string(metricsBytes))
			})
		}
	}

SendWindowMetricsEvent:
	event := embedder.WindowMetricsEvent{
		Width:      widthPx,
		Height:     heightPx,
		PixelRatio: pixelRatio,
	}

	flutterEnginePointer := *(*uintptr)(window.GetUserPointer())
	flutterEngine := (*embedder.FlutterEngine)(unsafe.Pointer(flutterEnginePointer))

	flutterEngine.SendWindowMetricsEvent(event)
}
