
 #include "library/flutter_embedder.h"

// C proxies def
bool proxy_make_current(void *v);
bool proxy_clear_current(void *v);
bool proxy_present(void *v);
uint32_t proxy_fbo_callback(void *v);
bool proxy_make_resource_current(void *v);

bool proxy_on_platform_message(FlutterPlatformMessage *message,
                             void *window);

// C helper
FlutterResult runFlutter(uintptr_t window, FlutterEngine *engine, FlutterProjectArgs * Args){

	FlutterRendererConfig config = {};
	config.type = kOpenGL;

	config.open_gl.struct_size = sizeof(FlutterOpenGLRendererConfig);

	config.open_gl.make_current = proxy_make_current;
	config.open_gl.clear_current = proxy_clear_current;
	config.open_gl.present = proxy_present;
	config.open_gl.fbo_callback = proxy_fbo_callback;
	config.open_gl.make_resource_current = proxy_make_resource_current;

	const char *args_arr[] = {
		"",
		"--dart-non-checked-mode",
		NULL,
	};

	Args->command_line_argc = 2;
	Args->command_line_argv = args_arr;
	Args->platform_message_callback = (FlutterPlatformMessageCallback)proxy_on_platform_message;

	return FlutterEngineRun(FLUTTER_ENGINE_VERSION, &config, Args, (void*)window, engine);
}

