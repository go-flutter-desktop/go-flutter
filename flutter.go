package flutter

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"time"
	"unsafe"

	"github.com/go-flutter-desktop/go-flutter/embedder"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/pkg/errors"
)

// dpPerInch defines the amount of display pixels per inch as defined for Flutter.
const dpPerInch = 160.0

// Run executes a flutter application with the provided options.
// given limitations this method must be called by the main function directly.
func Run(options ...Option) (err error) {
	var (
		window *glfw.Window
		c      config
	)

	// The Windows Title Handler and the TextInput handler come by default
	options = append(options, addHandlerWindowTitle())
	options = append(options, addHandlerTextInput())
	options = append(options, addHandlerClipboard())

	c = c.merge(options...)

	err = glfw.Init()
	if err != nil {
		return errors.Wrap(err, "glfw init")
	}
	defer glfw.Terminate()

	window, err = glfw.CreateWindow(c.WindowDimension.x, c.WindowDimension.y, "Loading..", nil, nil)
	if err != nil {
		return errors.Wrap(err, "creating glfw window")
	}
	defer window.Destroy()

	if c.WindowIconProvider != nil {
		images, err := c.WindowIconProvider()
		if err != nil {
			return errors.Wrap(err, "getting images from icon provider")
		}
		window.SetIcon(images)
	}

	if c.WindowInitializerDeprecated != nil {
		err = c.WindowInitializerDeprecated(window)
		if err != nil {
			return errors.Wrap(err, "executing window initializer")
		}
	}

	flu := runFlutter(window, c)

	defer flu.Shutdown()

	for !window.ShouldClose() {
		// glfw.WaitEvents()
		glfw.PollEvents()
		embedder.FlutterEngineFlushPendingTasksNow()
	}

	return nil
}

// GLFW callbacks to the Flutter Engine
func glfwCursorPositionCallbackAtPhase(
	window *glfw.Window, phase embedder.PointerPhase,
	x float64, y float64,
) {
	winWidth, _ := window.GetSize()
	frameBuffWidth, _ := window.GetFramebufferSize()
	contentScale := float64(frameBuffWidth / winWidth)
	event := embedder.PointerEvent{
		Phase:     phase,
		X:         x * contentScale,
		Y:         y * contentScale,
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
	}

	index := *(*int)(window.GetUserPointer())
	flutterEngine := embedder.FlutterEngineByIndex(index)

	flutterEngine.SendPointerEvent(event)
}

func glfwMouseButtonCallback(window *glfw.Window, key glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.MouseButton1 {
		x, y := window.GetCursorPos()

		// recalculate x and y from screen cordinates to pixels
		widthPx, _ := window.GetFramebufferSize()
		width, _ := window.GetSize()
		pixelsPerScreenCoordinate := float64(widthPx) / float64(width)
		x = x * pixelsPerScreenCoordinate
		y = y * pixelsPerScreenCoordinate

		if action == glfw.Press {
			glfwCursorPositionCallbackAtPhase(window, embedder.KDown, x, y)
			window.SetCursorPosCallback(func(window *glfw.Window, x float64, y float64) {
				x = x * pixelsPerScreenCoordinate
				y = y * pixelsPerScreenCoordinate
				glfwCursorPositionCallbackAtPhase(window, embedder.KMove, x, y)
			})
		}

		if action == glfw.Release {
			glfwCursorPositionCallbackAtPhase(window, embedder.KUp, x, y)
			window.SetCursorPosCallback(nil)
		}
	}
}

var state = textModel{}

func glfwKey(keyboardLayout KeyboardShortcuts) func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

	var modifierKey glfw.ModifierKey
	var wordTravellerKey int
	var wordTravellerKeyShift int

	switch runtime.GOOS {
	case "darwin":
		modifierKey = glfw.ModSuper
		wordTravellerKey = ModAlt
		wordTravellerKeyShift = ModShiftAlt

	default:
		modifierKey = glfw.ModControl
		wordTravellerKey = ModControl
		wordTravellerKeyShift = ModShiftControl

	}

	return func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		var modsIsModfifier = false
		var modsIsShift = false
		var modsIsWordModifierShift = false
		var modsIsWordModifier = false

		switch {
		case int(mods) == wordTravellerKeyShift:
			modsIsWordModifierShift = true
		case int(mods) == wordTravellerKey:
			modsIsWordModifier = true
		case mods == modifierKey:
			modsIsModfifier = true
		case int(mods) == ModShift:
			modsIsShift = true
		}

		if key == glfw.KeyEscape && action == glfw.Press {
			w.SetShouldClose(true)
		}

		if action == glfw.Repeat || action == glfw.Press {
			if state.clientID != 0 {

				switch key {
				case glfw.KeyEnter:
					if mods == modifierKey {
						performAction(w, "done")
					} else {
						state.addChar([]rune{'\n'})
						performAction(w, "newline")
					}

				case glfw.KeyHome:
					state.MoveCursorHome(modsIsModfifier, modsIsShift, modsIsWordModifierShift, modsIsWordModifier)

				case glfw.KeyEnd:
					state.MoveCursorEnd(modsIsModfifier, modsIsShift, modsIsWordModifierShift, modsIsWordModifier)

				case glfw.KeyLeft:
					state.MoveCursorLeft(modsIsModfifier, modsIsShift, modsIsWordModifierShift, modsIsWordModifier)

				case glfw.KeyRight:
					state.MoveCursorRight(modsIsModfifier, modsIsShift, modsIsWordModifierShift, modsIsWordModifier)

				case glfw.KeyDelete:
					state.Delete(modsIsModfifier, modsIsShift, modsIsWordModifierShift, modsIsWordModifier)

				case glfw.KeyBackspace:
					state.Backspace(modsIsModfifier, modsIsShift, modsIsWordModifierShift, modsIsWordModifier)

				case keyboardLayout.SelectAll:
					if mods == modifierKey {
						state.SelectAll()
					}

				case keyboardLayout.Copy:
					if mods == modifierKey && state.isSelected() {
						_, _, selectedContent := state.GetSelectedText()
						w.SetClipboardString(selectedContent)
					}

				case keyboardLayout.Cut:
					if mods == modifierKey && state.isSelected() {
						_, _, selectedContent := state.GetSelectedText()
						w.SetClipboardString(selectedContent)
						state.RemoveSelectedText()
					}

				case keyboardLayout.Paste:
					if mods == modifierKey {
						var clpString, err = w.GetClipboardString()
						if err != nil {
							log.Printf("unable to get the clipboard content: %v\n", err)
						} else {
							state.addChar([]rune(clpString))
						}
					}
				}
			}
		}
	}
}

// newGLFWFramebufferSizeCallback creates a func that is called on framebuffer resizes.
// When pixelRatio is set, the pixelRatio communicated to the Flutter embedder is not calculated.
func newGLFWFramebufferSizeCallback(pixelRatio float64, monitorScreenCoordinatesPerInch float64) func(*glfw.Window, int, int) {
	return func(window *glfw.Window, widthPx int, heightPx int) {
		index := *(*int)(window.GetUserPointer())
		flutterEngine := embedder.FlutterEngineByIndex(index)

		// calculate pixelRatio when it has not been forced.
		if pixelRatio == 0 {
			width, _ := window.GetSize()
			pixelsPerScreenCoordinate := float64(widthPx) / float64(width)
			dpi := pixelsPerScreenCoordinate * monitorScreenCoordinatesPerInch
			pixelRatio = dpi / dpPerInch

			// Limit the ratio to 1 to avoid rendering a smaller UI in standard resolution monitors.
			if pixelRatio < 1.0 {
				fmt.Println("calculated pixelRatio limited to a minimum of 1.0")
				pixelRatio = 1.0
			}
		}

		event := embedder.WindowMetricsEvent{
			Width:      widthPx,
			Height:     heightPx,
			PixelRatio: pixelRatio,
		}
		flutterEngine.SendWindowMetricsEvent(event)
	}
}

func glfwCharCallback(w *glfw.Window, char rune) {
	if state.clientID != 0 {
		state.addChar([]rune{char})
	}
}

// Flutter Engine
func runFlutter(window *glfw.Window, c config) *embedder.FlutterEngine {
	flutterEngine := embedder.NewFlutterEngine()

	// Engine arguments
	flutterEngine.AssetsPath = c.AssetsPath
	flutterEngine.IcuDataPath = c.ICUDataPath

	// Render callbacks
	flutterEngine.FMakeCurrent = func(v unsafe.Pointer) bool {
		w := glfw.GoWindow(v)
		w.MakeContextCurrent()
		return true
	}
	flutterEngine.FClearCurrent = func(v unsafe.Pointer) bool {
		glfw.DetachCurrentContext()
		return true
	}
	flutterEngine.FPresent = func(v unsafe.Pointer) bool {
		w := glfw.GoWindow(v)
		w.SwapBuffers()
		return true
	}
	flutterEngine.FFboCallback = func(v unsafe.Pointer) int32 {
		return 0
	}
	flutterEngine.FMakeResourceCurrent = func(v unsafe.Pointer) bool {
		return false
	}

	// PlatformMessage
	flutterEngine.FPlatfromMessage = func(platMessage *embedder.PlatformMessage, window unsafe.Pointer) bool {
		windows := glfw.GoWindow(window)

		hasDispatched := false

		// Dispatch the message from the Flutter Engine, to all of the PluginReceivers
		// having the same embedder.PlatformMessage.Channel name
		for _, receivers := range c.PlatformMessageReceivers[platMessage.Channel] {
			hasDispatched = receivers(platMessage, flutterEngine, windows) || hasDispatched
		}

		return hasDispatched
	}

	state.notifyState = func() {
		// log.Printf("Text: Sending to the flutter engine %v", state)
		updateEditingState(window)
	}

	flutterEngineIndex := flutterEngine.Index()
	window.SetUserPointer(unsafe.Pointer(&flutterEngineIndex))
	result := flutterEngine.Run(window.GLFWWindow(), c.VMArguments)

	if result != embedder.KSuccess {
		window.Destroy()
		panic("Couldn't launch the FlutterEngine")
	}

	glfwFramebufferSizeCallback := newGLFWFramebufferSizeCallback(c.ForcePixelRatio, getScreenCoordinatesPerInch())
	width, height := window.GetFramebufferSize()
	glfwFramebufferSizeCallback(window, width, height)
	var glfwKeyCallback func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)

	if c.KeyboardLayout != nil {
		glfwKeyCallback = glfwKey(*c.KeyboardLayout)
	} else {
		glfwKeyCallback = glfwKey(KeyboardQwertyLayout)
	}

	window.SetKeyCallback(glfwKeyCallback)
	window.SetFramebufferSizeCallback(glfwFramebufferSizeCallback)
	window.SetMouseButtonCallback(glfwMouseButtonCallback)
	window.SetCharCallback(glfwCharCallback)
	return flutterEngine
}

// getScreenCoordinatesPerInch returns the number of screen coordinates per
// inch for the main monitor. If the information is unavailable it returns
// a default value that assumes that a screen coordinate is one dp.
func getScreenCoordinatesPerInch() float64 {
	// TODO: multi-monitor support (#74)
	primaryMonitor := glfw.GetPrimaryMonitor()
	if primaryMonitor == nil {
		return dpPerInch
	}
	primaryMonitorMode := primaryMonitor.GetVideoMode()
	if primaryMonitorMode == nil {
		return dpPerInch
	}
	primaryMonitorWidthMM, _ := primaryMonitor.GetPhysicalSize()
	if primaryMonitorWidthMM == 0 {
		return dpPerInch
	}
	return float64(primaryMonitorMode.Width) / (float64(primaryMonitorWidthMM) / 25.4)
}

// Update the TextInput with the current state
func updateEditingState(window *glfw.Window) {
	editingState := argsEditingState{
		Text:                   string(state.word),
		SelectionAffinity:      "TextAffinity.downstream",
		SelectionBase:          state.selectionBase,
		SelectionExtent:        state.selectionExtent,
		SelectionIsDirectional: false,
	}

	editingStateMarchalled, _ := json.Marshal([]interface{}{
		state.clientID,
		editingState,
	})

	message := embedder.Message{
		Args:   editingStateMarchalled,
		Method: textUpdateStateMethod,
	}

	var mess = &embedder.PlatformMessage{
		Channel: textInputChannel,
		Message: message,
	}

	index := *(*int)(window.GetUserPointer())
	flutterEngine := embedder.FlutterEngineByIndex(index)

	flutterEngine.SendPlatformMessage(mess)
}

func performAction(window *glfw.Window, action string) {
	actionArgs, _ := json.Marshal([]interface{}{
		state.clientID,
		"TextInputAction." + action,
	})
	message := embedder.Message{
		Args:   actionArgs,
		Method: "TextInputClient.performAction",
	}
	var mess = &embedder.PlatformMessage{
		Channel: textInputChannel,
		Message: message,
	}

	index := *(*int)(window.GetUserPointer())
	flutterEngine := embedder.FlutterEngineByIndex(index)

	flutterEngine.SendPlatformMessage(mess)
}
