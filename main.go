package main

/*
#cgo CFLAGS: -I${SRCDIR}/flutter/library/
#cgo LDFLAGS: -L${SRCDIR}/flutter/library/linux -lflutter_engine -Wl,-rpath,$ORIGIN/flutter/library/linux
#include "flutter_embedder.h"

#include <stdbool.h>
#include <stdint.h>



bool proxy_OnPlatformMessage(FlutterPlatformMessage *message,
                             void *window);

static char* converter(uint8_t *str, size_t size){
	str[size] = '\0'; // Prevent overFlow
	return (char *)str;
}




*/
import "C"

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

func glfwCursorPositionCallbackAtPhase(
	window *glfw.Window, phase flutter.FlutterPointerPhase,
	x float64, y float64,
) {

	event := flutter.FlutterPointerEvent{
		Phase:     phase,
		X:         x,
		Y:         y,
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
	}

	flutter.FlutterEngineSendPointerEvent(window.GetUserPointer(), event)

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

func updateEditingState(w *glfw.Window) {

	// state.word = "Лайкаа"

	editingState := editingState{
		Text:                   state.word,
		SelectionAffinity:      "TextAffinity.downstream",
		SelectionBase:          len(state.word),
		SelectionExtent:        len(state.word),
		SelectionIsDirectional: false,
	}

	editingStateMarchalled, _ := json.Marshal([]interface{}{
		state.clientID,
		editingState,
	})

	message := flutterMessage{
		Args:   editingStateMarchalled,
		Method: "TextInputClient.updateEditingState",
	}

	marshalled, _ := json.Marshal(message)
	str := string(marshalled)

	var mess = flutter.FlutterPlatformMessage{
		Channel: textChannel,
		Message: str,
	}

	flutter.FlutterEngineSendPlatformMessage(w.GetUserPointer(), mess)
}

func glfwCharCallback(w *glfw.Window, char rune) {
	if state.clientID != 0 {
		state.word += string(char)
		updateEditingState(w)
	}
}

func glfwKeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	}

	if action == glfw.Repeat || action == glfw.Press {
		if state.clientID != 0 {
			if key == glfw.KeyBackspace {
				if state.word != "" {
					if mods == glfw.ModControl {
						state.word = string(deleteLeadingWord([]rune(state.word)))
						updateEditingState(w)
					} else {
						state.word = state.word[:len(state.word)-1]
						updateEditingState(w)
					}
				}
			}
		}
	}
}

func glfwWindowSizeCallback(w *glfw.Window, width int, height int) {

	event := flutter.FlutterWindowMetricsEvent{
		Width:       width,
		Height:      height,
		Pixel_ratio: 1.2,
	}

	flutter.FlutterEngineSendWindowMetricsEvent(w.GetUserPointer(), event)
}

func runFlutter(window *glfw.Window) *flutter.FlutterOpenGL {

	flutterOGL := flutter.FlutterOpenGL{}

	result := flutterOGL.Run(window.GLFWWindow())

	if result != flutter.KSuccess {
		window.Destroy()
		panic("Couldn't launch the FlutterEngine")
	}

	window.SetUserPointer(unsafe.Pointer(flutterOGL.Engine))

	width, height := window.GetSize()
	glfwWindowSizeCallback(window, width, height)

	window.SetKeyCallback(glfwKeyCallback)
	window.SetSizeCallback(glfwWindowSizeCallback)
	window.SetMouseButtonCallback(glfwMouseButtonCallback)
	window.SetCharCallback(glfwCharCallback)
	return &flutterOGL
}

var state struct {
	clientID float64
	word     string
}

//export proxy_OnPlatformMessage
func proxy_OnPlatformMessage(message *C.FlutterPlatformMessage, window unsafe.Pointer) C.bool {
	if message.message != nil {
		// fmt.Println(C.GoString(message.channel))
		str := C.GoString(C.converter(message.message, message.message_size))
		// fmt.Println(str)
		// fmt.Println()

		windows := glfw.GoWindow(window)

		res := flutterMessage{}
		json.Unmarshal([]byte(str), &res)

		if res.Method == setDescriptionMethod {
			msgBody := appSwitcherDescription{}
			json.Unmarshal(res.Args, &msgBody)
			windows.SetTitle(msgBody.Label)
		}

		if C.GoString(message.channel) == textChannel {

			if res.Method == clearClientMethod {
				state.clientID = 0
			}

			if res.Method == setClientMethod {
				var body []interface{}
				json.Unmarshal(res.Args, &body)
				state.clientID = body[0].(float64)
			}

			if res.Method == setEditingStateMethod {
				if state.clientID != 0 {
					editingState := editingState{}
					json.Unmarshal(res.Args, &editingState)
					state.word = editingState.Text
				}

			}

		}

	}
	return C.bool(true)
}

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

	engine := runFlutter(window)

	defer engine.Shutdown()

	for !window.ShouldClose() {
		// glfw.WaitEvents()
		glfw.PollEvents()
		C.__FlutterEngineFlushPendingTasksNow()
	}

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
