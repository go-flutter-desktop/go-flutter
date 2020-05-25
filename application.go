package flutter

import (
	"fmt"
	"os"
	"runtime"
	"time"
	"unsafe"

	"github.com/Xuanwo/go-locale"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/pkg/errors"
	"golang.org/x/text/language"

	"github.com/go-flutter-desktop/go-flutter/embedder"
	"github.com/go-flutter-desktop/go-flutter/internal/debounce"
	"github.com/go-flutter-desktop/go-flutter/internal/opengl"
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
		config: newApplicationConfig(),
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
	opt = append(opt, AddPlugin(defaultIsolatePlugin))
	opt = append(opt, AddPlugin(defaultMousecursorPlugin))

	// apply all configs
	for _, o := range opt {
		o(&app.config)
	}

	return app
}

// createResourceWindow creates an invisible GLFW window that shares the 'view'
// window's resource context. This window is used to upload resources in the
// background. Must be call after the 'view' window is created.
//
// Though optional, it is recommended that all embedders set this callback as
// it will lead to better performance in texture handling.
func createResourceWindow(window *glfw.Window) (*glfw.Window, error) {
	glfw.WindowHint(glfw.Decorated, glfw.False)
	glfw.WindowHint(glfw.Visible, glfw.False)
	if runtime.GOOS == "linux" {
		// Skia expects an EGL context on linux (libglvnd)
		glfw.WindowHint(glfw.ContextCreationAPI, glfw.EGLContextAPI)
	}
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
	case WindowModeMaximize:
		glfw.WindowHint(glfw.Maximized, glfw.True)
	case WindowModeBorderlessMaximize:
		glfw.WindowHint(glfw.Maximized, glfw.True)
		glfw.WindowHint(glfw.Decorated, glfw.False)
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

	opengl.GLFWWindowHint()

	{
		// TODO(drakirus): Delete this when https://github.com/go-gl/glfw/issues/272 is resolved.
		// Post an empty event from the main thread before it can happen in a non-main thread,
		// to work around https://github.com/glfw/glfw/issues/1649.
		glfw.PostEmptyEvent()
	}

	if a.config.windowInitialLocation.xpos != 0 {
		// To create the window at a specific position, make it initially invisible
		// using the Visible window hint, set its position and then show it.
		glfw.WindowHint(glfw.Visible, glfw.False)
	}

	glfw.WindowHint(glfw.ScaleToMonitor, glfw.True)
	if a.config.windowAlwaysOnTop {
		glfw.WindowHint(glfw.Floating, glfw.True)
	}
	if a.config.windowTransparent {
		glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)
	}

	if runtime.GOOS == "linux" {
		// Skia expects an EGL context on linux (libglvnd)
		glfw.WindowHint(glfw.ContextCreationAPI, glfw.EGLContextAPI)
	}

	a.window, err = glfw.CreateWindow(a.config.windowInitialDimensions.width, a.config.windowInitialDimensions.height, "Loading..", monitor, nil)
	if err != nil {
		return errors.Wrap(err, "creating glfw window")
	}
	defer a.window.Destroy()
	glfw.DefaultWindowHints()

	if a.config.windowInitialLocation.xpos != 0 {
		a.window.SetPos(a.config.windowInitialLocation.xpos, a.config.windowInitialLocation.ypos)
		a.window.Show()
	}

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

	a.window.SetTitle(ProjectName)

	if a.config.windowDimensionLimits.minWidth != 0 {
		a.window.SetSizeLimits(
			a.config.windowDimensionLimits.minWidth,
			a.config.windowDimensionLimits.minHeight,
			a.config.windowDimensionLimits.maxWidth,
			a.config.windowDimensionLimits.maxHeight,
		)
	}

	// Create a empty FlutterEngine.
	a.engine = embedder.NewFlutterEngine()

	// Set configuration values to engine.
	a.engine.AssetsPath = a.config.flutterAssetsPath
	a.engine.IcuDataPath = a.config.icuDataPath
	a.engine.ElfSnapshotPath = a.config.elfSnapshotpath

	// Create a messenger and init plugins
	messenger := newMessenger(a.engine)
	// Attach PlatformMessage callback function onto the engine
	a.engine.PlatfromMessage = messenger.handlePlatformMessage

	// Create a TextureRegistry
	texturer := newTextureRegistry(a.engine, a.window)
	// Attach TextureRegistry callback function onto the engine
	a.engine.GLExternalTextureFrameCallback = texturer.handleExternalTexture

	// Create a new eventloop
	eventLoop := newEventLoop(
		glfw.PostEmptyEvent, // Wakeup GLFW
		a.engine.RunTask,    // Flush tasks
	)
	// Attach TaskRunner callback functions onto the engine
	a.engine.TaskRunnerRunOnCurrentThread = eventLoop.RunOnCurrentThread
	a.engine.TaskRunnerPostTask = eventLoop.PostTask

	// Attach GL callback functions onto the engine
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

	// Set the glfw window user pointer to point to the FlutterEngine so that
	// callback functions may obtain the FlutterEngine from the glfw window
	// user pointer.
	flutterEnginePointer := uintptr(unsafe.Pointer(a.engine))
	defer func() {
		runtime.KeepAlive(flutterEnginePointer)
	}()
	a.window.SetUserPointer(unsafe.Pointer(&flutterEnginePointer))

	// Start the engine
	err = a.engine.Run(unsafe.Pointer(&flutterEnginePointer), a.config.vmArguments)
	if err != nil {
		fmt.Printf("go-flutter: %v\n", err)
		os.Exit(1)
	}

	languageTag, err := locale.Detect()
	if err != nil {
		fmt.Printf("go-flutter: failed to detect locale code: %v\n", err)
		languageTag = language.English
	}
	base, _ := languageTag.Base()
	region, _ := languageTag.Region()
	scriptCode, _ := languageTag.Script()
	err = a.engine.UpdateSystemLocale(base.String(), region.String(), scriptCode.String())
	if err != nil {
		fmt.Printf("go-flutter: %v\n", err)
	}

	// Register plugins
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

		// Extra init call for plugins that satisfy the PluginTexture interface.
		if texturePlugin, ok := p.(PluginTexture); ok {
			err = texturePlugin.InitPluginTexture(texturer)
			if err != nil {
				return errors.Wrap(err, "failed to initialize texture plugin"+fmt.Sprintf("%T", p))
			}
		}
	}

	// Change the flutter initial route
	initialRoute := os.Getenv("GOFLUTTER_ROUTE")
	if initialRoute != "" {
		defaultPlatformPlugin.addFrameworkReadyCallback(func() {
			defaultNavigationPlugin.
				channel.InvokeMethod("pushRoute", initialRoute)
		})
	}

	// Setup a new windowManager to handle windows pixel ratio's and pointer
	// devices.
	windowManager := newWindowManager(a.config.forcePixelRatio)
	// force first refresh
	windowManager.glfwRefreshCallback(a.window)
	// Attach glfw window callbacks for refresh and position changes
	a.window.SetRefreshCallback(windowManager.glfwRefreshCallback)
	// Debounce the position callback.
	// This avoid making too much flutter redraw and potentially redundant
	// network calls.
	glfwDebouceTasker := tasker.New()
	debounced := debounce.New(50 * time.Millisecond)
	// SetPosCallback is called when the window is moved, this directly calls
	// glfwRefreshCallback in order to redraw and avoid transparent scene.
	a.window.SetPosCallback(func(window *glfw.Window, xpos int, ypos int) {
		debounced(func() {
			glfwDebouceTasker.Do(func() {
				windowManager.glfwRefreshCallback(window)
			})
		})
	})

	// Attach glfw window callbacks for text input
	a.window.SetKeyCallback(
		func(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			defaultTextinputPlugin.glfwKeyCallback(window, key, scancode, action, mods)
			defaultKeyeventsPlugin.sendKeyEvent(window, key, scancode, action, mods)
		})
	a.window.SetCharCallback(defaultTextinputPlugin.glfwCharCallback)

	// Attach glfw window callback for iconification
	a.window.SetIconifyCallback(defaultLifecyclePlugin.glfwIconifyCallback)

	// Attach glfw window callbacks for mouse input
	a.window.SetCursorEnterCallback(windowManager.glfwCursorEnterCallback)
	a.window.SetCursorPosCallback(windowManager.glfwCursorPosCallback)
	a.window.SetMouseButtonCallback(windowManager.glfwMouseButtonCallback)
	a.window.SetScrollCallback(
		func(window *glfw.Window, xoff float64, yoff float64) {
			windowManager.glfwScrollCallback(window, xoff, yoff, a.config.scrollAmount)
		})

	// Shutdown the engine if we return from this function (on purpose or panic)
	defer a.engine.Shutdown()

	// Handle events until the window indicates we should stop. An event may tell the window to stop, in which case
	// we'll exit on next iteration.
	for !a.window.ShouldClose() {
		eventLoop.WaitForEvents(func(duration float64) {
			glfw.WaitEventsTimeout(duration)
		})

		// Execute tasks that MUST be run in the engine thread (!blocks rendering!)
		glfwDebouceTasker.ExecuteTasks()
		messenger.engineTasker.ExecuteTasks()
		texturer.engineTasker.ExecuteTasks()
	}

	fmt.Println("go-flutter: closing application")

	return nil
}

// TODO: app.Start(), app.Wait()?
