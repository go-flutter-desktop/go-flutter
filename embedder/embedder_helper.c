
#include "flutter_embedder.h"
#include <stdlib.h>

// C proxies def
bool proxy_make_current(void *v);
bool proxy_clear_current(void *v);
bool proxy_present(void *v);
uint32_t proxy_fbo_callback(void *v);
bool proxy_make_resource_current(void *v);
void *proxy_gl_proc_resolver(void *v, const char *procname);
void proxy_platform_message_callback(const FlutterPlatformMessage *message, void *window);

// C helper
FlutterEngineResult runFlutter(uintptr_t window, FlutterEngine *engine, FlutterProjectArgs *Args,
                               const char *const *vmArgs, int nVmAgrs)
{
        FlutterRendererConfig config = {};
        config.type = kOpenGL;

        config.open_gl.struct_size = sizeof(FlutterOpenGLRendererConfig);
        config.open_gl.make_current = proxy_make_current;
        config.open_gl.clear_current = proxy_clear_current;
        config.open_gl.present = proxy_present;
        config.open_gl.fbo_callback = proxy_fbo_callback;
        config.open_gl.make_resource_current = proxy_make_resource_current;
        config.open_gl.gl_proc_resolver = proxy_gl_proc_resolver;

        Args->command_line_argc = nVmAgrs;
        Args->command_line_argv = vmArgs;
        Args->platform_message_callback = proxy_platform_message_callback;

        return FlutterEngineRun(FLUTTER_ENGINE_VERSION, &config, Args, (void *)window, engine);
}

char **makeCharArray(int size)
{
        return calloc(sizeof(char *), size);
}

void setArrayString(char **a, char *s, int n)
{
        a[n] = s;
}
