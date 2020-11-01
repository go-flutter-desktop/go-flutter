package flutter

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/go-flutter-desktop/go-flutter/embedder"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// dpPerInch defines the amount of display pixels per inch as defined for Flutter.
const dpPerInch = 160.0

// TODO (GeertJohan): better name for this, confusing with 'actual' window
// managers. Renderer interface? implemented by this type for glfw? type
// glfwRenderer or glfwManager? All the attaching to glfw.Window must be done
// during manager init in that case. Cannot be done by Application.
type windowManager struct {
	// forcedPixelRatio forces the pixelRatio to given value, when value is not zero.
	forcedPixelRatio float64

	// current pointer state
	pointerPhase          embedder.PointerPhase
	pointerButton         embedder.PointerButtonMouse
	pointerCurrentlyAdded bool

	// caching of ppsc to avoid re-calculating every event
	pixelsPerScreenCoordinate float64
}

func newWindowManager(forcedPixelRatio float64) *windowManager {
	return &windowManager{
		forcedPixelRatio:          forcedPixelRatio,
		pixelsPerScreenCoordinate: 1.0,
		pointerPhase:              embedder.PointerPhaseHover,
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

	event := embedder.PointerEvent{
		Phase:   phase,
		X:       x * m.pixelsPerScreenCoordinate,
		Y:       y * m.pixelsPerScreenCoordinate,
		Buttons: m.pointerButton,
	}

	flutterEnginePointer := *(*uintptr)(window.GetUserPointer())
	flutterEngine := (*embedder.FlutterEngine)(unsafe.Pointer(flutterEnginePointer))

	// Always send a pointer event with PhaseMove before an eventual PhaseRemove.
	// If x/y on the last move doesn't equal x/y on the PhaseRemove, the remove
	// is canceled in Flutter.
	if phase == embedder.PointerPhaseRemove {
		event.Phase = embedder.PointerPhaseHover
		flutterEngine.SendPointerEvent(event)
		event.Phase = embedder.PointerPhaseRemove
	}

	flutterEngine.SendPointerEvent(event)

	if phase == embedder.PointerPhaseAdd {
		m.pointerCurrentlyAdded = true
	} else if phase == embedder.PointerPhaseRemove {
		m.pointerCurrentlyAdded = false
	}
}

func (m *windowManager) sendPointerEventButton(window *glfw.Window, phase embedder.PointerPhase) {
	x, y := window.GetCursorPos()
	event := embedder.PointerEvent{
		Phase:      phase,
		X:          x * m.pixelsPerScreenCoordinate,
		Y:          y * m.pixelsPerScreenCoordinate,
		SignalKind: embedder.PointerSignalKindNone,
		Buttons:    m.pointerButton,
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
		SignalKind:   embedder.PointerSignalKindScroll,
		ScrollDeltaX: xDelta,
		ScrollDeltaY: yDelta,
		Buttons:      m.pointerButton,
	}

	flutterEnginePointer := *(*uintptr)(window.GetUserPointer())
	flutterEngine := (*embedder.FlutterEngine)(unsafe.Pointer(flutterEnginePointer))

	flutterEngine.SendPointerEvent(event)
}

func (m *windowManager) glfwCursorEnterCallback(window *glfw.Window, entered bool) {
	x, y := window.GetCursorPos()
	if entered {
		m.sendPointerEvent(window, embedder.PointerPhaseAdd, x, y)
		// the mouse can enter the windows while having button pressed.
		// if so, don't overwrite the phase.
		if m.pointerButton == 0 {
			m.pointerPhase = embedder.PointerPhaseHover
		}
	} else {
		// if the mouse is still in 'phaseMove' outside the window (click-drag
		// outside). Don't remove the cursor.
		if m.pointerButton == 0 {
			m.sendPointerEvent(window, embedder.PointerPhaseRemove, x, y)
		}
	}
}

func (m *windowManager) glfwCursorPosCallback(window *glfw.Window, x, y float64) {
	m.sendPointerEvent(window, m.pointerPhase, x, y)
}

func (m *windowManager) handleButtonPhase(window *glfw.Window, action glfw.Action, buttons embedder.PointerButtonMouse) {
	if action == glfw.Press {
		m.pointerButton |= buttons
		// If only one button is pressed then each bits of buttons will be equals
		// to m.pointerButton.
		if m.pointerButton == buttons {
			m.sendPointerEventButton(window, embedder.PointerPhaseDown)
		} else {
			// if any other buttons are already pressed when a new button is pressed,
			// the engine is expecting a Move phase instead of a Down phase.
			m.sendPointerEventButton(window, embedder.PointerPhaseMove)
		}
		m.pointerPhase = embedder.PointerPhaseMove
	}

	if action == glfw.Release {
		// Always send a pointer event with PhaseMove before an eventual
		// PhaseUp. Even if the last button was released. If x/y on the last
		// move doesn't equal x/y on the PhaseUp, the click is canceled in
		// Flutter. On MacOS, the Release event always has y-1 of the last move
		// event. By sending a PhaseMove here (after the release) we avoid a
		// difference in x/y.
		m.sendPointerEventButton(window, embedder.PointerPhaseMove)

		m.pointerButton ^= buttons
		// If all button are released then m.pointerButton is cleared
		if m.pointerButton == 0 {
			m.sendPointerEventButton(window, embedder.PointerPhaseUp)
			m.pointerPhase = embedder.PointerPhaseHover
		} else {
			// if any other buttons are still pressed when one button is released
			// the engine is expecting a Move phase instead of a Up phase.
			m.sendPointerEventButton(window, embedder.PointerPhaseMove)
		}
	}
}

func (m *windowManager) glfwMouseButtonCallback(window *glfw.Window, key glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	switch key {
	case glfw.MouseButtonLeft:
		m.handleButtonPhase(window, action, embedder.PointerButtonMousePrimary)
	case glfw.MouseButtonRight:
		m.handleButtonPhase(window, action, embedder.PointerButtonMouseSecondary)
	case glfw.MouseButtonMiddle:
		m.handleButtonPhase(window, action, embedder.PointerButtonMouseMiddle)
	default:
		m.handleButtonPhase(window, action, 1<<uint(key))
	}
}

func (m *windowManager) glfwScrollCallback(window *glfw.Window, xoff float64, yoff float64, scrollAmount float64) {
	scrollModifier := -scrollAmount
	m.sendPointerEventScroll(window, xoff*scrollModifier, yoff*scrollModifier)
}

// glfwRefreshCallback is called when the window needs a reresh, this
// can occur when the window is resized, was covered by another window, etc.
// When forcedPixelratio is zero, the forcedPixelratio communicated to the
// Flutter embedder is calculated based on physical and logical screen
// dimensions.
func (m *windowManager) glfwRefreshCallback(window *glfw.Window) {
	widthPx, heightPx := window.GetFramebufferSize()
	width, _ := window.GetSize()
	if width == 0 {
		fmt.Println("go-flutter: Cannot calculate pixelsPerScreenCoordinate for zero-width window.")
		return
	}
	m.pixelsPerScreenCoordinate = float64(widthPx) / float64(width)

	var pixelRatio float64
	if m.forcedPixelRatio != 0 {
		pixelRatio = m.forcedPixelRatio
	} else {
		if runtime.GOOS == "linux" {
			pixelRatio = m.getPixelRatioLinux(window)
		} else {
			pixelRatio = m.getPixelRatioOther(window)
		}
	}

	event := embedder.WindowMetricsEvent{
		Width:      widthPx,
		Height:     heightPx,
		PixelRatio: pixelRatio,
	}

	flutterEnginePointer := *(*uintptr)(window.GetUserPointer())
	flutterEngine := (*embedder.FlutterEngine)(unsafe.Pointer(flutterEnginePointer))

	flutterEngine.SendWindowMetricsEvent(event)
}

// getPixelRatioOther, getPixelRatioLinux isn't well working on other platform.
// GLFW window.GetContentScale() works better:
// https://github.com/go-flutter-desktop/go-flutter/pull/458
func (m *windowManager) getPixelRatioOther(window *glfw.Window) float64 {
	xscale, _ := window.GetContentScale()
	return float64(xscale)
}

// getPixelRatioLinux returns the Flutter pixel_ratio is defined as DPI/dp
// given framebuffer size and the current window information.
// Same as defined in the official LINUX embedder:
// https://github.com/flutter/engine/blob/master/shell/platform/glfw/flutter_glfw.cc
// Fallback to getPixelRatioOther if error occur.
func (m *windowManager) getPixelRatioLinux(window *glfw.Window) float64 {
	widthPx, heightPx := window.GetFramebufferSize()

	var selectedMonitor *glfw.Monitor
	winX, winY := window.GetPos()
	winCenterX, winCenterY := winX+widthPx/2, winY+heightPx/2

	monitors := glfw.GetMonitors()
	for _, monitor := range monitors {
		monX1, monY1 := monitor.GetPos()
		monMode := monitor.GetVideoMode()
		if monMode == nil {
			continue
		}
		monX2, monY2 := monX1+monMode.Width, monY1+monMode.Height
		if (monX1 <= winCenterX && winCenterX <= monX2) &&
			(monY1 <= winCenterY && winCenterY <= monY2) {
			selectedMonitor = monitor
			break
		}
	}

	if selectedMonitor == nil {
		// when no monitor was selected, try fallback to primary monitor
		// TODO: ? perhaps select monitor that is "closest" to the window ?
		selectedMonitor = glfw.GetPrimaryMonitor()
	}
	if selectedMonitor == nil {
		return m.getPixelRatioOther(window)
	}
	selectedMonitorMode := selectedMonitor.GetVideoMode()
	if selectedMonitorMode == nil {
		return m.getPixelRatioOther(window)
	}
	selectedMonitorWidthMM, _ := selectedMonitor.GetPhysicalSize()
	if selectedMonitorWidthMM == 0 {
		return m.getPixelRatioOther(window)
	}
	monitorScreenCoordinatesPerInch := float64(selectedMonitorMode.Width) / (float64(selectedMonitorWidthMM) / 25.4)

	dpi := m.pixelsPerScreenCoordinate * monitorScreenCoordinatesPerInch
	pixelRatio := dpi / dpPerInch

	// If the pixelRatio is lower to 1 use this scale factor to downscale the ContentScale
	if pixelRatio < 1.0 {
		pixelRatio *= m.getPixelRatioOther(window)
	}
	return pixelRatio
}
