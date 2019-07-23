package flutter

import (
	"fmt"
	"sync"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/pkg/errors"

	"github.com/go-flutter-desktop/go-flutter/embedder"
)

type texturer struct {
	window       *glfw.Window
	channels     map[int64]externalTextureHanlder
	channelsLock sync.RWMutex
}

func newTexturer(window *glfw.Window) *texturer {
	return &texturer{
		window:   window,
		channels: make(map[int64]externalTextureHanlder),
	}
}

func (t *texturer) init() error {
	t.window.MakeContextCurrent()
	// Important! Call gl.Init only under the presence of an active OpenGL context,
	// i.e., after MakeContextCurrent.
	if err := gl.Init(); err != nil {
		return errors.Wrap(err, "texturer gl init")
	}
	return nil
}

// SetTextureHandler registers a handler to be invoked when the Flutter
// application want to get a PixelBuffer to draw into the scene.
//
// Registration overwrites any previous registration for the same textureID
// name. Use nil as handler to deregister.
func (t *texturer) SetTextureHandler(textureID int64, handler ExternalTextureHanlderFunc) {
	t.channelsLock.Lock()
	if handler == nil {
		delete(t.channels, textureID)
	} else {
		t.channels[textureID] = externalTextureHanlder{
			handle: handler,
		}
	}
	t.channelsLock.Unlock()
}

// ExternalTextureHanlderFunc describes the function that handles external
// Texture on a given ID.
type ExternalTextureHanlderFunc func(width int, height int) (bool, *PixelBuffer)

type externalTextureHanlder struct {
	// handle is called when flutter needs the PixelBuffer
	handle ExternalTextureHanlderFunc
	// gl texture to refer to for this handler
	texture uint32
}

// PixelBuffer is an in-memory (RGBA) image.
type PixelBuffer struct {
	// Pix holds the image's pixels, in R, G, B, A order.
	Pix []uint8
	// Width and Height of the image's bounds
	Width, Height int
}

func (t *texturer) handleExternalTexture(textureID int64,
	width int, height int) (bool, *embedder.FlutterOpenGLTexture) {

	t.channelsLock.RLock()
	registration, registrationExists := t.channels[textureID]
	t.channelsLock.RUnlock()

	if !registrationExists {
		fmt.Printf("go-flutter: no texture handler found for Texture ID: %v\n", textureID)
		return false, nil
	}
	res, pixelBuffer := registration.handle(width, height)
	if !res || pixelBuffer == nil {
		return false, nil
	}

	if len(pixelBuffer.Pix) == 0 {
		return false, nil
	}

	t.window.MakeContextCurrent()

	if registration.texture == 0 {
		gl.GenTextures(1, &registration.texture)
		gl.BindTexture(gl.TEXTURE_2D, registration.texture)
		// set the texture wrapping parameters
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)
		// set texture filtering parameters
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	}

	gl.BindTexture(gl.TEXTURE_2D, registration.texture)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(pixelBuffer.Width),
		int32(pixelBuffer.Height),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(pixelBuffer.Pix))

	return true, &embedder.FlutterOpenGLTexture{
		Target: gl.TEXTURE_2D,
		Name:   registration.texture,
		Format: gl.RGBA8,
	}

}
