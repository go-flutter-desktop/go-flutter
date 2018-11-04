package gutter

import (
	"github.com/Drakirus/go-flutter-desktop-embedder/flutter"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// PluginReceivers do stuff when receiving Message from the Engine,
// send result with `flutterEngine.SendPlatformMessageResponse`
type PluginReceivers func(
	message flutter.PlatformMessage,
	flutterEngine *flutter.EngineOpenGL,
	window *glfw.Window,
) bool
