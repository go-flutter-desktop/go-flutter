package flutter

import (
	"github.com/pkg/errors"
)

// Texture is an identifier for texture declaration
type Texture struct {
	ID       int64
	registry *TextureRegistry
}

// Register registers a textureID with his associated handler
func (t *Texture) Register(handler ExternalTextureHanlderFunc) error {
	t.registry.setTextureHandler(t.ID, handler)
	err := t.registry.engine.RegisterExternalTexture(t.ID)
	if err != nil {
		t.registry.setTextureHandler(t.ID, nil)
		return errors.Errorf("'go-flutter' couldn't register texture with id: '%v': %v", t.ID, err)
	}
	return nil
}

// FrameAvailable mark a texture buffer is ready to be draw in the flutter scene
func (t *Texture) FrameAvailable() error {
	err := t.registry.engine.MarkExternalTextureFrameAvailable(t.ID)
	if err != nil {
		return errors.Errorf("'go-flutter' couldn't mark frame available of texture with id: '%v': %v", t.ID, err)
	}
	return nil
}

// UnRegister unregisters a textureID with his associated handler
func (t *Texture) UnRegister() error {
	err := t.registry.engine.UnregisterExternalTexture(t.ID)
	if err != nil {
		return errors.Errorf("'go-flutter' couldn't unregisters texture with id: '%v': %v", t.ID, err)
	}
	t.registry.setTextureHandler(t.ID, nil)
	return nil
}
