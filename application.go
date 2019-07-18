package flutter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/pkg/errors"

	"github.com/go-flutter-desktop/go-flutter/embedder"
	"github.com/go-flutter-desktop/go-flutter/internal/execpath"
	"github.com/go-flutter-desktop/go-flutter/internal/tasker"
)

// Run executes a flutter application with the provided options.
// given limitations this method must be called by the main function directly.
//
// Run(opt) is short for NewApplication(opt).Run()
func Run(opt ...Option) (err error) {
	return NewApplication(opt...).Run()
}

// Application provides the flutter engine in a user friendly matter.
type Application struct {
	config         config
	engine         *embedder.FlutterEngine
	window         *glfw.Window
	resourceWindow *glfw.Window
}

// NewApplication creates a new application with provided options.
func NewApplication(opt ...Option) *Application {
	app := &Application{
		config: defaultApplicationConfig,
	}

	// The platformPlugin, textinputPlugin, etc. are currently hardcoded as we
	// have a hard link with GLFW. The plugins must be singleton and are
	// accessed directly from the flutter package to wire up with glfw. If
	// there's going to be a renderer interface, it's init would replace this
	// configuration.
	opt = append(opt, AddPlugin(defaultNavigationPlugin))
	opt = append(opt, AddPlugin(defaultPlatformPlugin))
	opt = append(opt, AddPlugin(defaultTextinputPlugin))
	opt = append(opt, AddPlugin(defaultLifecyclePlugin))
	opt = append(opt, AddPlugin(defaultKeyeventsPlugin))
	opt = append(opt, AddPlugin(defaultAccessibilityPlugin))

	// apply all configs
	for _, o := range opt {
		o(&app.config)
	}

	return app
}

// createResourceWindow creates an invisible GLFW window that shares the 'view'
// window's resource context. This window is used to upload resources in the
// background. Must be call after the 'view' window is created.
func createResourceWindow(window *glfw.Window) (*glfw.Window, error) {
	glfw.WindowHint(glfw.Decorated, glfw.False)
	glfw.WindowHint(glfw.Visible, glfw.False)
	resourceWindow, err := glfw.CreateWindow(1, 1, "", nil, window)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the resource window")
	}
	glfw.DefaultWindowHints()
	return resourceWindow, nil
}

// Run starts the application and waits for it to finish.
func (a *Application) Run() error {
	runtime.LockOSThread()

	err := glfw.Init()
	if err != nil {
		return errors.Wrap(err, "glfw init")
	}
	defer glfw.Terminate()

	var monitor *glfw.Monitor
	switch a.config.windowMode {
	case WindowModeDefault:
		// nothing
	case WindowModeBorderless:
		glfw.WindowHint(glfw.Decorated, glfw.False)
	case WindowModeBorderlessFullscreen:
		monitor = glfw.GetPrimaryMonitor()
		mode := monitor.GetVideoMode()
		a.config.windowInitialDimensions.width = mode.Width
		a.config.windowInitialDimensions.height = mode.Height
		glfw.WindowHint(glfw.RedBits, mode.RedBits)
		glfw.WindowHint(glfw.GreenBits, mode.GreenBits)
		glfw.WindowHint(glfw.BlueBits, mode.BlueBits)
		glfw.WindowHint(glfw.RefreshRate, mode.RefreshRate)
	default:
		return errors.Errorf("invalid window mode %T", a.config.windowMode)
	}

	a.window, err = glfw.CreateWindow(a.config.windowInitialDimensions.width, a.config.windowInitialDimensions.height, "Loading..", monitor, nil)
	if err != nil {
		return errors.Wrap(err, "creating glfw window")
	}
	glfw.DefaultWindowHints()
	defer a.window.Destroy()

	a.resourceWindow, err = createResourceWindow(a.window)
	if err != nil {
		fmt.Printf("go-flutter: WARNING %v\n", err)
	} else {
		defer a.resourceWindow.Destroy()
	}

	if a.config.windowIconProvider != nil {
		images, err := a.config.windowIconProvider()
		if err != nil {
			return errors.Wrap(err, "getting images from icon provider")
		}
		a.window.SetIcon(images)
	}

	if a.config.windowInitializerDeprecated != nil {
		err = a.config.windowInitializerDeprecated(a.window)
		if err != nil {
			return errors.Wrap(err, "executing window initializer")
		}
	}

	if a.config.windowDimensionLimits.minWidth != 0 {
		a.window.SetSizeLimits(
			a.config.windowDimensionLimits.minWidth,
			a.config.windowDimensionLimits.minHeight,
			a.config.windowDimensionLimits.maxWidth,
			a.config.windowDimensionLimits.maxHeight,
		)
	}

	a.engine = embedder.NewFlutterEngine()

	messenger := newMessenger(a.engine)
	for _, p := range a.config.plugins {
		err = p.InitPlugin(messenger)
		if err != nil {
			return errors.Wrap(err, "failed to initialize plugin "+fmt.Sprintf("%T", p))
		}

		// Extra init call for plugins that satisfy the PluginGLFW interface.
		if glfwPlugin, ok := p.(PluginGLFW); ok {
			err = glfwPlugin.InitPluginGLFW(a.window)
			if err != nil {
				return errors.Wrap(err, "failed to initialize glfw plugin"+fmt.Sprintf("%T", p))
			}
		}
	}

	if a.config.flutterAssetsPath != "" {
		a.engine.AssetsPath = a.config.flutterAssetsPath
	} else {
		execPath, err := execpath.ExecPath()
		if err != nil {
			return errors.Wrap(err, "failed to resolve path for executable")
		}
		a.engine.AssetsPath = filepath.Join(filepath.Dir(execPath), "flutter_assets")
	}

	if a.config.icuDataPath != "" {
		a.engine.IcuDataPath = a.config.icuDataPath
	} else {
		execPath, err := execpath.ExecPath()
		if err != nil {
			return errors.Wrap(err, "failed to resolve path for executable")
		}
		a.engine.IcuDataPath = filepath.Join(filepath.Dir(execPath), "icudtl.dat")
	}

	// Render callbacks
	a.engine.GLMakeCurrent = func() bool {
		a.window.MakeContextCurrent()
		return true
	}
	a.engine.GLClearCurrent = func() bool {
		glfw.DetachCurrentContext()
		return true
	}
	a.engine.GLPresent = func() bool {
		a.window.SwapBuffers()
		return true
	}
	a.engine.GLFboCallback = func() int32 {
		return 0
	}
	a.engine.GLMakeResourceCurrent = func() bool {
		if a.resourceWindow == nil {
			return false
		}
		a.resourceWindow.MakeContextCurrent()
		return true
	}
	a.engine.GLProcResolver = func(procName string) unsafe.Pointer {
		return glfw.GetProcAddress(procName)
	}

	a.engine.PlatfromMessage = messenger.handlePlatformMessage

	// Not very nice, but we can only really fix this when there's a pluggable
	// renderer.
	defaultTextinputPlugin.keyboardLayout = a.config.keyboardLayout

	flutterEnginePointer := uintptr(unsafe.Pointer(a.engine))
	defer func() {
		runtime.KeepAlive(flutterEnginePointer)
	}()
	a.window.SetUserPointer(unsafe.Pointer(&flutterEnginePointer))

	result := a.engine.Run(unsafe.Pointer(&flutterEnginePointer), a.config.vmArguments)
	if result != embedder.ResultSuccess {
		switch result {
		case embedder.ResultInvalidLibraryVersion:
			fmt.Printf("go-flutter: engine.Run() returned result code %d (invalid library version)\n", result)
		case embedder.ResultInvalidArguments:
			fmt.Printf("go-flutter: engine.Run() returned result code %d (invalid arguments)\n", result)
		default:
			fmt.Printf("go-flutter: engine.Run() returned result code %d (unknown result code)\n", result)
		}
		os.Exit(1)
	}

	defaultPlatformPlugin.glfwTasker = tasker.New()

	m := newWindowManager()
	m.forcedPixelRatio = a.config.forcePixelRatio

	m.glfwRefreshCallback(a.window)
	a.window.SetRefreshCallback(m.glfwRefreshCallback)
	a.window.SetPosCallback(m.glfwPosCallback)

	a.window.SetKeyCallback(
		func(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			defaultTextinputPlugin.glfwKeyCallback(window, key, scancode, action, mods)
			defaultKeyeventsPlugin.sendKeyEvent(window, key, scancode, action, mods)
		})
	a.window.SetCharCallback(defaultTextinputPlugin.glfwCharCallback)

	a.window.SetIconifyCallback(defaultLifecyclePlugin.glfwIconifyCallback)

	a.window.SetCursorEnterCallback(m.glfwCursorEnterCallback)
	a.window.SetCursorPosCallback(m.glfwCursorPosCallback)
	a.window.SetMouseButtonCallback(m.glfwMouseButtonCallback)
	a.window.SetScrollCallback(m.glfwScrollCallback)
	defer a.engine.Shutdown()

	for !a.window.ShouldClose() {
		glfw.WaitEventsTimeout(0.016) // timeout to get 60fps-ish iterations
		embedder.FlutterEngineFlushPendingTasksNow()
		defaultPlatformPlugin.glfwTasker.ExecuteTasks()
		messenger.engineTasker.ExecuteTasks()
	}

	fmt.Println("go-flutter: closing application")

	return nil
}

// TODO: app.Start(), app.Wait()?
