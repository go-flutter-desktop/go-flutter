package flutter

import (
	"fmt"

	"github.com/pkg/errors"
)

// popBehavior defines how an application should handle the navigation pop
// event from the flutter side.
type popBehavior int

const (
	// PopBehaviorNone means the system navigation pop event is ignored.
	PopBehaviorNone popBehavior = iota
	// PopBehaviorHide hides the application window on a system navigation pop
	// event.
	PopBehaviorHide
	// PopBehaviorIconify minimizes/iconifies the application window on a system
	// navigation pop event.
	PopBehaviorIconify
	// PopBehaviorClose closes the application on a system navigation pop event.
	PopBehaviorClose
)

// PopBehavior sets the PopBehavior on the application
func PopBehavior(p popBehavior) Option {
	return func(c *config) {
		// TODO: this is a workarround because there is no renderer interface
		// yet. We rely on a platform plugin singleton to handle events from the
		// flutter side. Should go via Application and renderer abstraction
		// layer.
		//
		// Downside of this workarround is that it will configure the pop
		// behavior for all Application's within the same Go process.
		defaultPlatformPlugin.popBehavior = p
	}
}

func (p *platformPlugin) handleSystemNavigatorPop(arguments interface{}) (reply interface{}, err error) {
	switch p.popBehavior {
	case PopBehaviorNone:
		return nil, nil
	case PopBehaviorHide:
		p.glfwTasker.Do(func() {
			p.window.Hide()
		})
		return nil, nil
	case PopBehaviorIconify:
		var err error
		p.glfwTasker.Do(func() {
			err = p.window.Iconify()
		})
		if err != nil {
			fmt.Printf("go-flutter: error on iconifying window: %v\n", err)
			return nil, errors.Wrap(err, "failed to iconify window")
		}
		return nil, nil
	case PopBehaviorClose:
		p.glfwTasker.Do(func() {
			p.window.SetShouldClose(true)
		})
		return nil, nil
	default:
		return nil, errors.Errorf("unknown pop behavior %T not implemented by platform handler", p.popBehavior)
	}
}
