package gutter

import (
	"encoding/json"
	"log"
	"time"
	"unsafe"

	"C"

	"github.com/Drakirus/go-flutter-desktop-embedder/flutter"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// Option for gutter
type Option func(*config)

// OptionAssetPath specify the flutter asset directory.
func OptionAssetPath(p string) Option {
	return func(c *config) {
		c.AssetPath = p
	}
}

// OptionICUDataPath specify the path to the ICUData.
func OptionICUDataPath(p string) Option {
	return func(c *config) {
		c.ICUDataPath = p
	}
}

// OptionVMArguments specify the arguments to the Dart VM.
func OptionVMArguments(a []string) Option {
	return func(c *config) {
		// First should be argument is argv[0]
		c.VMArguments = append([]string{""}, a...)
	}
}

// OptionWindowDimension specify the startup's dimention of the window.
func OptionWindowDimension(x int, y int) Option {
	return func(c *config) {
		c.WindowDimension.x = x
		c.WindowDimension.y = y
	}
}

// OptionWindowInitializer allow initializing the window.
func OptionWindowInitializer(ini func(*glfw.Window) error) Option {
	return func(c *config) {
		c.WindowInitializer = ini
	}
}

// OptionPixelRatio specify the scale factor for the physical screen.
func OptionPixelRatio(ratio float64) Option {
	return func(c *config) {
		c.PixelRatio = ratio
	}
}

type config struct {
	WindowDimension struct {
		x int
		y int
	}
	AssetPath         string
	ICUDataPath       string
	WindowInitializer func(*glfw.Window) error
	PixelRatio        float64
	VMArguments       []string
}

func (t config) merge(options ...Option) config {
	for _, opt := range options {
		opt(&t)
	}

	return t
}

// Run executes a flutter application with the provided options.
// given limitations this method must be called by the main function directly.
func Run(options ...Option) (err error) {
	var (
		window *glfw.Window
		c      config
	)
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

	flutterOGL := *(*flutter.EngineOpenGL)(window.GetUserPointer())
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

func glfwKeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	}

	if action == glfw.Repeat || action == glfw.Press {
		if state.clientID != 0 {

			switch key {
			case glfw.KeyHome:
				state.MoveCursorHome(int(mods))

			case glfw.KeyEnd:
				state.MoveCursorEnd(int(mods))

			case glfw.KeyLeft:
				state.MoveCursorLeft(int(mods))

			case glfw.KeyRight:
				state.MoveCursorRight(int(mods))

			case glfw.KeyDelete:
				state.Delete(int(mods))

			case glfw.KeyBackspace:
				state.Backspace(int(mods))

			case glfw.KeyA:
				if mods == glfw.ModControl {
					state.SelectAll()
				}

			case glfw.KeyC:
				if mods == glfw.ModControl && state.isSelected() {
					_, _, selectedContent := state.GetSelectedText()
					w.SetClipboardString(selectedContent)
				}

			case glfw.KeyX:
				if mods == glfw.ModControl && state.isSelected() {
					_, _, selectedContent := state.GetSelectedText()
					w.SetClipboardString(selectedContent)
					state.RemoveSelectedText()
				}

			case glfw.KeyV:
				if mods == glfw.ModControl {
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

func glfwWindowSizeCallback(window *glfw.Window, width int, height int) {
	flutterOGL := *(*flutter.EngineOpenGL)(window.GetUserPointer())
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
		AssetsPath:  unsafe.Pointer(C.CString(c.AssetPath)),
		IcuDataPath: unsafe.Pointer(C.CString(c.ICUDataPath)),
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
		// Messaging (TextInput)
		FPlatfromMessage: onPlatformMessage,
		PixelRatio:       c.PixelRatio,
	}

	state.notifyState = func() {
		// log.Printf("Text: Sending to the flutter engine %v", state)
		updateEditingState(window)
	}

	result := flutterOGL.Run(window.GLFWWindow(), c.VMArguments)

	if result != flutter.KSuccess {
		window.Destroy()
		panic("Couldn't launch the FlutterEngine")
	}

	window.SetUserPointer(unsafe.Pointer(&flutterOGL))

	width, height := window.GetFramebufferSize()
	glfwWindowSizeCallback(window, width, height)

	window.SetKeyCallback(glfwKeyCallback)
	window.SetFramebufferSizeCallback(glfwWindowSizeCallback)
	window.SetMouseButtonCallback(glfwMouseButtonCallback)
	window.SetCharCallback(glfwCharCallback)
	return &flutterOGL
}

// Message from the Flutter Engine

func onPlatformMessage(platMessage flutter.PlatformMessage, window unsafe.Pointer) bool {

	windows := glfw.GoWindow(window)
	message := platMessage.Message

	// fmt.Println(string(platMessage.Message.Args))

	if message.Method == flutter.SetDescriptionMethod {
		msgBody := flutter.ArgsAppSwitcherDescription{}
		json.Unmarshal(message.Args, &msgBody)
		windows.SetTitle(msgBody.Label)
	}

	if platMessage.Channel == flutter.TextInputChannel {
		switch message.Method {
		case flutter.TextInputClientClear:
			state.clientID = 0
		case flutter.TextInputClientSet:
			var body []interface{}
			json.Unmarshal(message.Args, &body)
			state.clientID = body[0].(float64)
		case flutter.TextInputSetEditState:
			if state.clientID != 0 {
				editingState := flutter.ArgsEditingState{}
				json.Unmarshal(message.Args, &editingState)
				state.word = editingState.Text
				state.selectionBase = editingState.SelectionBase
				state.selectionExtent = editingState.SelectionExtent
			}
		default:
			// log.Printf("unhandled text input method: %#v\n", platMessage.Message)
		}
	}

	return true
}

// Update the TextInput with the current state
func updateEditingState(window *glfw.Window) {

	// state.word = "Лайкаа"

	editingState := flutter.ArgsEditingState{
		Text:                   state.word,
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
		Method: flutter.TextUpdateStateMethod,
	}

	var mess = flutter.PlatformMessage{
		Channel: flutter.TextInputChannel,
		Message: message,
	}

	flutterOGL := *(*flutter.EngineOpenGL)(window.GetUserPointer())
	flutterOGL.EngineSendPlatformMessage(mess)
}
