package flutter

import (
	"fmt"
	"sync"

	"github.com/go-flutter-desktop/go-flutter/embedder"
	"github.com/go-flutter-desktop/go-flutter/internal/opengl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/pkg/errors"
)

// once is used for the lazy initialization of go-gl/gl.
// The initialization occur on the first requested texture's frame.
var once sync.Once

// TextureRegistry is a registry entry for a managed Texture.
type TextureRegistry struct {
	window       *glfw.Window
	engine       *embedder.FlutterEngine
	channels     map[int64]*externalTextureHanlder
	channelsLock sync.RWMutex

	texture      int64
	texturesLock sync.Mutex
}

type externalTextureHanlder struct {
	// handle is called when flutter needs the PixelBuffer
	handle ExternalTextureHanlderFunc
	// gl texture to refer to for this handler
	texture uint32
}

func newTextureRegistry(engine *embedder.FlutterEngine, window *glfw.Window) *TextureRegistry {
	return &TextureRegistry{
		window:   window,
		engine:   engine,
		channels: make(map[int64]*externalTextureHanlder),
	}
}

func (t *TextureRegistry) init() error {
	t.window.MakeContextCurrent()
	// Important! Call open.Init only under the presence of an active OpenGL context,
	// i.e., after MakeContextCurrent.
	if err := opengl.Init(); err != nil {
		return errors.Wrap(err, "TextureRegistry gl init failed")
	}
	return nil
}

// NewTexture creates a new Texture
func (t *TextureRegistry) NewTexture() Texture {
	t.texturesLock.Lock()
	defer t.texturesLock.Unlock()
	t.texture++
	return Texture{ID: t.texture, registry: t}
}

// ExternalTextureHanlderFunc describes the function that handles external
// Texture on a given ID.
type ExternalTextureHanlderFunc func(width int, height int) (bool, *PixelBuffer)

// PixelBuffer is an in-memory (RGBA) image.
type PixelBuffer struct {
	// Pix holds the image's pixels, in R, G, B, A order.
	Pix []uint8
	// Width and Height of the image's bounds
	Width, Height int
}

// setTextureHandler registers a handler to be invoked when the Flutter
// application want to get a PixelBuffer to draw into the scene.
//
// Registration overwrites any previous registration for the same textureID
// name. Use nil as handler to deregister.
func (t *TextureRegistry) setTextureHandler(textureID int64, handler ExternalTextureHanlderFunc) {
	t.channelsLock.Lock()
	if handler == nil {
		texture := t.channels[textureID]
		if texture != nil {
			opengl.DeleteTextures(1, &texture.texture)
		}
		delete(t.channels, textureID)
	} else {
		t.channels[textureID] = &externalTextureHanlder{
			handle: handler,
		}
	}
	t.channelsLock.Unlock()
}

func (t *TextureRegistry) handleExternalTexture(textureID int64,
	width int, height int) *embedder.FlutterOpenGLTexture {

	once.Do(func() {
		t.init()
	})

	t.channelsLock.RLock()
	registration, registrationExists := t.channels[textureID]
	t.channelsLock.RUnlock()

	if !registrationExists {
		fmt.Printf("go-flutter: no texture handler found for Texture ID: %v\n", textureID)
		return nil
	}
	res, pixelBuffer := registration.handle(width, height)
	if !res || pixelBuffer == nil {
		return nil
	}

	if len(pixelBuffer.Pix) == 0 {
		return nil
	}

	t.window.MakeContextCurrent()

	if registration.texture == 0 {
		opengl.CreateTexture(&registration.texture)
	}

	opengl.BindTexture(registration.texture)

	opengl.TexImage2D(
		int32(pixelBuffer.Width),
		int32(pixelBuffer.Height),
		opengl.Ptr(pixelBuffer.Pix),
	)

	return &embedder.FlutterOpenGLTexture{
		Target: opengl.TEXTURE2D,
		Name:   registration.texture,
		Format: opengl.RGBA8,
	}

}
