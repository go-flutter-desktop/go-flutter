package flutter

import (
	"github.com/go-flutter-desktop/go-flutter/embedder"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// PluginReceivers do stuff when receiving Message from the Engine,
// send result with `flutterEngine.SendPlatformMessageResponse`
type PluginReceivers func(
	message *embedder.PlatformMessage,
	flutterEngine *embedder.FlutterEngine,
	window *glfw.Window,
) bool
