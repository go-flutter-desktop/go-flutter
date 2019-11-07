// +build !openglnone

package opengl

// The default version (3.3) of OpenGL used by go-flutter.
// If you want to support other version, copy/pase this file, change the import
// statement, add builds constraints and open a PR.

import (
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// const exposed to go-flutter
const (
	TEXTURE2D = gl.TEXTURE_2D
	RGBA8     = gl.RGBA8
)

// Init opengl
func Init() error {
	return gl.Init()
}

// DeleteTextures deletes named textures
func DeleteTextures(n int32, textures *uint32) {
	gl.DeleteTextures(n, textures)
}

// CreateTexture creates a texture for go-flutter uses
func CreateTexture(texture *uint32) {
	gl.GenTextures(1, texture)
	gl.BindTexture(gl.TEXTURE_2D, *texture)
	// set the texture wrapping parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)
	// set texture filtering parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
}

// BindTexture binds a named texture to a texturing target
func BindTexture(texture uint32) {
	gl.BindTexture(gl.TEXTURE_2D, texture)
}

// Ptr takes a slice or pointer (to a singular scalar value or the first
// element of an array or slice) and returns its GL-compatible address.
func Ptr(data interface{}) unsafe.Pointer {
	return gl.Ptr(data)
}

// TexImage2D specifies a two-dimensional texture image
func TexImage2D(width, height int32, pixels unsafe.Pointer) {
	// It the current flutter/engine can only support RGBA texture.
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		width,
		height,
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		pixels,
	)
}

// GLFWWindowHint sets hints for the next call to CreateWindow.
func GLFWWindowHint() {
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
}
