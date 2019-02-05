package gutter

import (
	"encoding/json"
	"log"
	"runtime"
	"time"
	"unsafe"

	"github.com/Drakirus/go-flutter-desktop-embedder/flutter"
	"github.com/go-gl/glfw/v3.2/glfw"
)

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

	if err = glfw.Init(); err != nil {
		return err
	}
	defer glfw.Terminate()

	window, err = glfw.CreateWindow(c.WindowDimension.x, c.WindowDimension.y, "Loading..", nil, nil)
	if err != nil {
		return err
	}
	defer window.Destroy()

	if err = c.WindowInitializer(window); err != nil {
		return err
	}

	engine := runFlutter(window, c)

	defer engine.Shutdown()

	for !window.ShouldClose() {
		// glfw.WaitEvents()
		glfw.PollEvents()
		flutter.EngineFlushPendingTasksNow()
	}

	return nil
}

// GLFW callbacks to the Flutter Engine
func glfwCursorPositionCallbackAtPhase(
	window *glfw.Window, phase flutter.PointerPhase,
	x float64, y float64,
) {
	winWidth, _ := window.GetSize()
	frameBuffWidth, _ := window.GetFramebufferSize()
	contentScale := float64(frameBuffWidth / winWidth)
	event := flutter.PointerEvent{
		Phase:     phase,
		X:         x * contentScale,
		Y:         y * contentScale,
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
	}

	flutterOGL := flutter.SelectEngine(0)

	flutterOGL.EngineSendPointerEvent(event)
}

func glfwMouseButtonCallback(window *glfw.Window, key glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {

	if key == glfw.MouseButton1 && action == glfw.Press {
		x, y := window.GetCursorPos()
		glfwCursorPositionCallbackAtPhase(window, flutter.KDown, x, y)
		window.SetCursorPosCallback(func(window *glfw.Window, x float64, y float64) {
			glfwCursorPositionCallbackAtPhase(window, flutter.KMove, x, y)
		})
	}

	if key == glfw.MouseButton1 && action == glfw.Release {
		x, y := window.GetCursorPos()
		glfwCursorPositionCallbackAtPhase(window, flutter.KUp, x, y)
		window.SetCursorPosCallback(nil)
	}

}

var state = textModel{}

func glfwKey(keyboardLayout KeyboardShortcuts) func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

	var modifierKey glfw.ModifierKey
	var wordTravellerKey int

	switch runtime.GOOS {
	case "darwin":
		modifierKey = glfw.ModSuper
		wordTravellerKey = ModAlt

	default:
		modifierKey = glfw.ModControl
		wordTravellerKey = ModControl

	}

	return func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		var modsIsModfifier = false
		var modsIsShift = false
		var modsIsAltShift = false
		var modsIsAlt = false

		switch {
		case int(mods) == ModShiftAlt:
			modsIsAltShift = true
		case mods == modifierKey:
			modsIsModfifier = true
		case int(mods) == ModShift:
			modsIsShift = true
		case int(mods) == wordTravellerKey:
			modsIsAlt = true
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
					state.MoveCursorHome(modsIsModfifier, modsIsShift, modsIsAltShift, modsIsAlt)

				case glfw.KeyEnd:
					state.MoveCursorEnd(modsIsModfifier, modsIsShift, modsIsAltShift, modsIsAlt)

				case glfw.KeyLeft:
					state.MoveCursorLeft(modsIsModfifier, modsIsShift, modsIsAltShift, modsIsAlt)

				case glfw.KeyRight:
					state.MoveCursorRight(modsIsModfifier, modsIsShift, modsIsAltShift, modsIsAlt)

				case glfw.KeyDelete:
					state.Delete(modsIsModfifier, modsIsShift, modsIsAltShift, modsIsAlt)

				case glfw.KeyBackspace:
					state.Backspace(modsIsModfifier, modsIsShift, modsIsAltShift, modsIsAlt)

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

func glfwWindowSizeCallback(window *glfw.Window, width int, height int) {

	flutterOGL := flutter.SelectEngine(0)

	event := flutter.WindowMetricsEvent{
		Width:      width,
		Height:     height,
		PixelRatio: flutterOGL.PixelRatio,
	}
	flutterOGL.EngineSendWindowMetricsEvent(event)
}

func glfwCharCallback(w *glfw.Window, char rune) {
	if state.clientID != 0 {
		state.addChar([]rune{char})
	}
}

// Flutter Engine
func runFlutter(window *glfw.Window, c config) *flutter.EngineOpenGL {

	flutterOGL := flutter.EngineOpenGL{
		// Engine arguments
		AssetsPath:  c.AssetPath,
		IcuDataPath: c.ICUDataPath,
		// Render callbacks
		FMakeCurrent: func(v unsafe.Pointer) bool {
			w := glfw.GoWindow(v)
			w.MakeContextCurrent()
			return true
		},
		FClearCurrent: func(v unsafe.Pointer) bool {
			glfw.DetachCurrentContext()
			return true
		},
		FPresent: func(v unsafe.Pointer) bool {
			w := glfw.GoWindow(v)
			w.SwapBuffers()
			return true
		},
		FFboCallback: func(v unsafe.Pointer) int32 {
			return 0
		},
		FMakeResourceCurrent: func(v unsafe.Pointer) bool {
			return false
		},
		PixelRatio: c.PixelRatio,
	}

	// PlatformMessage
	flutterOGL.FPlatfromMessage = func(platMessage *flutter.PlatformMessage, window unsafe.Pointer) bool {
		windows := glfw.GoWindow(window)

		hasDispatched := false

		// Dispatch the message from the Flutter Engine, to all of the PluginReceivers
		// having the same flutter.PlatformMessage.Channel name
		for _, receivers := range c.PlatformMessageReceivers[platMessage.Channel] {
			hasDispatched = receivers(platMessage, &flutterOGL, windows) || hasDispatched
		}

		return hasDispatched
	}

	state.notifyState = func() {
		// log.Printf("Text: Sending to the flutter engine %v", state)
		updateEditingState(window)
	}

	NbEngine := flutter.NumberOfEngines()
	window.SetUserPointer(unsafe.Pointer(&NbEngine))
	result := flutterOGL.Run(window.GLFWWindow(), c.VMArguments)

	if result != flutter.KSuccess {
		window.Destroy()
		panic("Couldn't launch the FlutterEngine")
	}

	width, height := window.GetFramebufferSize()
	glfwWindowSizeCallback(window, width, height)
	var glfwKeyCallback func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)

	if c.KeyboardLayout != nil {
		glfwKeyCallback = glfwKey(*c.KeyboardLayout)
	} else {
		glfwKeyCallback = glfwKey(KeyboardQwertyLayout)
	}

	window.SetKeyCallback(glfwKeyCallback)
	window.SetFramebufferSizeCallback(glfwWindowSizeCallback)
	window.SetMouseButtonCallback(glfwMouseButtonCallback)
	window.SetCharCallback(glfwCharCallback)
	return &flutterOGL
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

	message := flutter.Message{
		Args:   editingStateMarchalled,
		Method: textUpdateStateMethod,
	}

	var mess = &flutter.PlatformMessage{
		Channel: textInputChannel,
		Message: message,
	}

	flutterOGL := flutter.SelectEngine(0)
	flutterOGL.SendPlatformMessage(mess)
}

func performAction(window *glfw.Window, action string) {
	actionArgs, _ := json.Marshal([]interface{}{
		state.clientID,
		"TextInputAction." + action,
	})
	message := flutter.Message{
		Args:   actionArgs,
		Method: "TextInputClient.performAction",
	}
	var mess = &flutter.PlatformMessage{
		Channel: textInputChannel,
		Message: message,
	}
	flutterOGL := flutter.SelectEngine(0)
	flutterOGL.SendPlatformMessage(mess)
}
