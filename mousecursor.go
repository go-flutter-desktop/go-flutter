package flutter

import (
	"errors"
	"fmt"
	"github.com/go-flutter-desktop/go-flutter/plugin"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const mousecursorChannelName = "flutter/mousecursor"

// mousecursorPlugin implements flutter.Plugin and handles method calls to the
// flutter/mousecursor channel.
type mousecursorPlugin struct {
	window     *glfw.Window
	lastCursor *glfw.Cursor
}

var defaultMousecursorPlugin = &mousecursorPlugin{}

func (p *mousecursorPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	channel := plugin.NewMethodChannel(messenger, mousecursorChannelName, plugin.StandardMethodCodec{})
	channel.HandleFuncSync("activateSystemCursor", p.handleActivateSystemCursor)
	return nil
}
func (p *mousecursorPlugin) InitPluginGLFW(window *glfw.Window) error {
	p.window = window
	return nil
}

func (p *mousecursorPlugin) handleActivateSystemCursor(arguments interface{}) (reply interface{}, err error) {
	args := arguments.(map[interface{}]interface{})
	var cursor *glfw.Cursor
	if args["kind"] == "none" {
		p.window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	} else {
		p.window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	}
	switch kind := args["kind"]; {
	case kind == "none":
	case kind == "basic" || kind == "forbidden" || kind == "grab" || kind == "grabbing":
		// go-gl GLFW currently (latest tagged v3.3 version) has no cursors for "forbidden", "grab" and "grabbing"
		// TODO: Wait for https://github.com/glfw/glfw/commit/7dbdd2e6a5f01d2a4b377a197618948617517b0e to appear in go-gl GLFW and implement the "forbidden" cursor
		cursor = glfw.CreateStandardCursor(glfw.ArrowCursor)
	case kind == "click":
		cursor = glfw.CreateStandardCursor(glfw.HandCursor)
	case kind == "text":
		cursor = glfw.CreateStandardCursor(glfw.IBeamCursor)
	default:
		return nil, errors.New(fmt.Sprintf("cursor kind %s not implemented", args["kind"]))
	}
	if p.lastCursor != nil {
		p.lastCursor.Destroy()
	}
	if cursor != nil {
		p.window.SetCursor(cursor)
		p.lastCursor = cursor
	}
	return nil, nil
}
