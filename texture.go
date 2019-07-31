package flutter

import (
	"errors"
	"fmt"

	"github.com/go-flutter-desktop/go-flutter/embedder"
)

// Texture is an identifier for texture declaration
type Texture struct {
	ID       int64
	registry *TextureRegistry
}

// Register registers a textureID with his associated handler
func (t *Texture) Register(handler ExternalTextureHanlderFunc) error {
	t.registry.setTextureHandler(t.ID, handler)
	result := t.registry.engine.RegisterExternalTexture(t.ID)
	if result != embedder.ResultSuccess {
		t.registry.setTextureHandler(t.ID, nil)
		return errors.New("'go-flutter' couldn't register texture with id: " + fmt.Sprint(t.ID))
	}
	return nil
}

// FrameAvailable mark a texture buffer is ready to be draw in the flutter scene
func (t *Texture) FrameAvailable() error {
	result := t.registry.engine.MarkExternalTextureFrameAvailable(t.ID)
	if result != embedder.ResultSuccess {
		return errors.New("'go-flutter' couldn't mark frame available of texture with id: " + fmt.Sprint(t.ID))
	}
	return nil
}

// UnRegister unregisters a textureID with his associated handler
func (t *Texture) UnRegister() error {
	result := t.registry.engine.UnregisterExternalTexture(t.ID)
	if result != embedder.ResultSuccess {
		return errors.New("'go-flutter' couldn't unregisters texture with id: " + fmt.Sprint(t.ID))
	}
	t.registry.setTextureHandler(t.ID, nil)
	return nil
}
