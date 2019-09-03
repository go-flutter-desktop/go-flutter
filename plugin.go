package flutter

import (
	"github.com/go-gl/glfw/v3.2/glfw"

	"github.com/go-flutter-desktop/go-flutter/plugin"
)

// TODO: move type Plugin into package plugin?

// Plugin defines the interface that each plugin must implement.
// When InitPlugin is called, the plugin may execute setup operations.
// The BinaryMessenger is passed to allow the plugin to register channels.
// A plugin may optionally implement PluginGLFW.
type Plugin interface {
	// InitPlugin is called during the startup of the flutter application. The
	// plugin is responsible for setting up channels using the BinaryMessenger.
	// If an error is returned it is printend the application is stopped.
	InitPlugin(messenger plugin.BinaryMessenger) error
}

// PluginGLFW defines the interface for plugins that are GLFW-aware. Plugins may
// implement this interface to receive access to the *glfw.Window. Note that
// plugins must still implement the Plugin interface. The call to InitPluginGLFW
// is made after the call to InitPlugin.
//
// PluginGLFW is separated because not all plugins need to know about glfw,
// Adding glfw.Window to the InitPlugin call would add glfw as dependency to
// every plugin implementation. Also, this helps in a scenarion where glfw is
// moved into a separate renderer/glfw package.
//
// The PluginGLFW interface is not stable and may change at any time.
type PluginGLFW interface {
	// Any type inmplementing PluginGLFW must also implement Plugin.
	Plugin
	// InitPluginGLFW is called after the call to InitPlugin. When an error is
	// returned it is printend the application is stopped.
	InitPluginGLFW(window *glfw.Window) error
}

// PluginTexture defines the interface for plugins that needs to create and
// manage backend textures. Plugins may implement this interface to receive
// access to the TextureRegistry. Note that plugins must still implement the
// Plugin interface. The call to PluginTexture is made after the call to
// PluginGLFW.
//
// PluginTexture is separated because not all plugins need to send raw pixel to
// the Flutter scene.
type PluginTexture interface {
	// Any type inmplementing PluginTexture must also implement Plugin.
	Plugin
	// InitPluginTexture is called after the call to InitPlugin. When an error is
	// returned it is printend the application is stopped.
	InitPluginTexture(registry *TextureRegistry) error
}
