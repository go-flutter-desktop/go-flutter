// +build openglnone

package opengl

// Don't link go-flutter against OpenGL, less-dependency at the cost of not
// supporting external texture (eg: video_plugin)
//
// compile go-flutter with: hover build|run --opengl=none

import (
	"unsafe"
)

// const exposed to go-flutter
const (
	TEXTURE2D = 0
	RGBA8     = 0
)

// Init opengl
func Init() error { return nil }

// DeleteTextures deletes named textures
func DeleteTextures(n int32, textures *uint32) {}

// CreateTexture creates a texture for go-flutter uses
func CreateTexture(texture *uint32) {}

// BindTexture binds a named texture to a texturing target
func BindTexture(texture uint32) {}

// Ptr takes a slice or pointer (to a singular scalar value or the first
// element of an array or slice) and returns its GL-compatible address.
func Ptr(data interface{}) unsafe.Pointer { return nil }

// TexImage2D specifies a two-dimensional texture image
func TexImage2D(width, height int32, pixels unsafe.Pointer) {
	panic("go-flutter: go-flutter wasn't compiled with support for external texture plugin.")
}

// GLFWWindowHint sets hints for the next call to CreateWindow.
func GLFWWindowHint() {}
