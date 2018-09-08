package main

import (
	"encoding/json"
	"flutter_desktop_go_embedding/flutter"
	"image"
	_ "image/png"
	"log"
	"os"
	"runtime"
	"time"
	"unsafe"

	"github.com/go-gl/glfw/v3.2/glfw"
)

// TextInput model
var state = textModel{}

func init() {
	runtime.LockOSThread()
}

func main() {
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	window, err := glfw.CreateWindow(800, 600, "Loading..", nil, nil)
	if err != nil {
		panic(err)
	}

	defer window.Destroy()

	// set icon
	if err := setIcon(window); err != nil {
		log.Printf("unable to set window icon: %v\n", err)
	}

	assetsPath := "./flutter_project/stocks/build/flutter_assets"
	icuDataPath := "./flutter/library/icudtl.dat"

	engine := runFlutter(window, assetsPath, icuDataPath)

	defer engine.Shutdown()

	for !window.ShouldClose() {
		// glfw.WaitEvents()
		glfw.PollEvents()
		flutter.EngineFlushPendingTasksNow()
	}

}

// GLFW callbacks to the Flutter Engine

func glfwCursorPositionCallbackAtPhase(
	window *glfw.Window, phase flutter.PointerPhase,
	x float64, y float64,
) {

	event := flutter.PointerEvent{
		Phase:     phase,
		X:         x,
		Y:         y,
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

	event := flutter.WindowMetricsEvent{
		Width:      width,
		Height:     height,
		PixelRatio: 1.2,
	}

	flutterOGL := *(*flutter.EngineOpenGL)(window.GetUserPointer())
	flutterOGL.EngineSendWindowMetricsEvent(event)
}

func glfwCharCallback(w *glfw.Window, char rune) {
	if state.clientID != 0 {
		state.addChar([]rune{char})
	}
}

// Flutter Engine

func runFlutter(window *glfw.Window, assetsPath string, icuDataPath string) *flutter.EngineOpenGL {

	flutterOGL := flutter.EngineOpenGL{
		// Engine arguments
		AssetsPath:  assetsPath,
		IcuDataPath: icuDataPath,
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
	}

	state.notifyState = func() {
		// log.Printf("Text: Sending to the flutter engine %v", state)
		updateEditingState(window)
	}

	result := flutterOGL.Run(window.GLFWWindow())

	if result != flutter.KSuccess {
		window.Destroy()
		panic("Couldn't launch the FlutterEngine")
	}

	window.SetUserPointer(unsafe.Pointer(&flutterOGL))

	width, height := window.GetSize()
	glfwWindowSizeCallback(window, width, height)

	window.SetKeyCallback(glfwKeyCallback)
	window.SetSizeCallback(glfwWindowSizeCallback)
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

		if message.Method == flutter.ClearClientMethod {
			state.clientID = 0
		}

		if message.Method == flutter.SetClientMethod {
			var body []interface{}
			json.Unmarshal(message.Args, &body)
			state.clientID = body[0].(float64)
		}

		if message.Method == flutter.SetEditingStateMethod {
			if state.clientID != 0 {
				editingState := flutter.ArgsEditingState{}
				json.Unmarshal(message.Args, &editingState)
				state.word = editingState.Text
			}

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

func setIcon(window *glfw.Window) error {
	imgFile, err := os.Open("assets/icon.png")
	if err != nil {
		return err
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return err
	}
	window.SetIcon([]image.Image{img})
	return nil
}
