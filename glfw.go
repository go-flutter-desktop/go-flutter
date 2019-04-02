package flutter

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-flutter-desktop/go-flutter/embedder"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// dpPerInch defines the amount of display pixels per inch as defined for Flutter.
const dpPerInch = 160.0

// GLFW callbacks to the Flutter Engine
func glfwCursorPositionCallbackAtPhase(
	window *glfw.Window, phase embedder.PointerPhase,
	x float64, y float64,
) {

	// Pointer events from GLFW are described using screen coordinates.
	// We need to provide Flutter with the position in pixels.
	width, _ := window.GetSize()
	if width == 0 {
		fmt.Println("go-flutter: Cannot calculate pointer position in zero-width window.")
		return
	}
	widthPx, _ := window.GetFramebufferSize()
	pixelsPerScreenCoordinate := float64(widthPx) / float64(width)

	event := embedder.PointerEvent{
		Phase:     phase,
		X:         x * pixelsPerScreenCoordinate,
		Y:         y * pixelsPerScreenCoordinate,
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
	}

	index := *(*int)(window.GetUserPointer())
	flutterEngine := embedder.FlutterEngineByIndex(index)

	flutterEngine.SendPointerEvent(event)
}

func glfwMouseButtonCallback(window *glfw.Window, key glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.MouseButton1 {
		x, y := window.GetCursorPos()

		if action == glfw.Press {
			glfwCursorPositionCallbackAtPhase(window, embedder.PointerPhaseDown, x, y)
			// TODO(#90): new kHover pointerphase in embedder.h suggests that pos callback is relevant outside "down" as well...
			window.SetCursorPosCallback(func(window *glfw.Window, x float64, y float64) {
				glfwCursorPositionCallbackAtPhase(window, embedder.PointerPhaseMove, x, y)
			})
		}

		if action == glfw.Release {
			glfwCursorPositionCallbackAtPhase(window, embedder.PointerPhaseUp, x, y)
			window.SetCursorPosCallback(nil)
		}
	}
}

// newGLFWFramebufferSizeCallback creates a func that is called on framebuffer
// resizes. When pixelRatio is zero, the pixelRatio communicated to the Flutter
// embedder is calculated based on physical and logical screen dimensions.
func newGLFWFramebufferSizeCallback(pixelRatio float64) func(*glfw.Window, int, int) {
	var oncePrintPixelRatioLimit sync.Once

	return func(window *glfw.Window, widthPx int, heightPx int) {
		index := *(*int)(window.GetUserPointer())
		flutterEngine := embedder.FlutterEngineByIndex(index)

		// calculate pixelRatio when it has not been forced.
		if pixelRatio == 0 {

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
				pixelRatio = 1.0
				goto SendWindowMetricsEvent
			}

			pixelsPerScreenCoordinate := float64(widthPx) / float64(width)
			dpi := pixelsPerScreenCoordinate * monitorScreenCoordinatesPerInch
			pixelRatio = dpi / dpPerInch

			// Limit the ratio to 1 to avoid rendering a smaller UI in standard resolution monitors.
			if pixelRatio < 1.0 {
				metrics := map[string]interface{}{
					"ppsc":           pixelsPerScreenCoordinate,
					"windowWidthPx":  widthPx,
					"windowWidthSc":  width,
					"mscpi":          monitorScreenCoordinatesPerInch,
					"dpi":            dpi,
					"pixelRatio":     pixelRatio,
					"monitorWidthMm": primaryMonitorWidthMM,
					"monitorWidthSc": primaryMonitorMode.Width,
				}
				pixelRatio = 1.0
				oncePrintPixelRatioLimit.Do(func() {
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
		flutterEngine.SendWindowMetricsEvent(event)
	}
}
