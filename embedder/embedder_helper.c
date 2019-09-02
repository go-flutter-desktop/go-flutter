#include <stdio.h>
#include <stdlib.h>

#include "embedder.h"

// C proxy definitions. These are implemented in Go.
bool proxy_make_current(void *user_data);
bool proxy_clear_current(void *user_data);
bool proxy_present(void *user_data);
uint32_t proxy_fbo_callback(void *user_data);
bool proxy_make_resource_current(void *user_data);
void *proxy_gl_proc_resolver(void *user_data, const char *procname);
void proxy_platform_message_callback(const FlutterPlatformMessage *message,
                                     void *user_data);
bool proxy_gl_external_texture_frame_callback(void *user_data,
                                              int64_t texture_id, size_t width,
                                              size_t height,
                                              FlutterOpenGLTexture *texture);

bool proxy_runs_task_on_current_thread_callback(void *user_data);
void proxy_post_task_callback(FlutterTask task, uint64_t target_time_nanos,
                              void *user_data);

void proxy_desktop_binary_reply(const uint8_t *data, size_t data_size,
                                void *user_data);

// C helper
FlutterEngineResult runFlutter(void *user_data, FlutterEngine *engine,
                               FlutterProjectArgs *Args,
                               const char *const *vmArgs, int nVmAgrs) {
  FlutterRendererConfig config = {};
  config.type = kOpenGL;

  config.open_gl.struct_size = sizeof(FlutterOpenGLRendererConfig);
  config.open_gl.make_current = proxy_make_current;
  config.open_gl.clear_current = proxy_clear_current;
  config.open_gl.present = proxy_present;
  config.open_gl.fbo_callback = proxy_fbo_callback;
  config.open_gl.make_resource_current = proxy_make_resource_current;
  config.open_gl.gl_proc_resolver = proxy_gl_proc_resolver;
  config.open_gl.gl_external_texture_frame_callback =
      proxy_gl_external_texture_frame_callback;

  Args->command_line_argc = nVmAgrs;
  Args->command_line_argv = vmArgs;
  Args->platform_message_callback = proxy_platform_message_callback;

  // Configure task runner interop
  FlutterTaskRunnerDescription platform_task_runner = {};
  platform_task_runner.struct_size = sizeof(FlutterTaskRunnerDescription);
  platform_task_runner.user_data = user_data;
  platform_task_runner.runs_task_on_current_thread_callback =
      proxy_runs_task_on_current_thread_callback;
  platform_task_runner.post_task_callback = proxy_post_task_callback;

  FlutterCustomTaskRunners custom_task_runners = {};
  custom_task_runners.struct_size = sizeof(FlutterCustomTaskRunners);
  custom_task_runners.platform_task_runner = &platform_task_runner;
  Args->custom_task_runners = &custom_task_runners;

  return FlutterEngineRun(FLUTTER_ENGINE_VERSION, &config, Args, user_data,
                          engine);
}

char **makeCharArray(int size) { return calloc(sizeof(char *), size); }

void setArrayString(char **a, char *s, int n) { a[n] = s; }

FlutterEngineResult
createMessageResponseHandle(FlutterEngine engine, void *user_data,
                            FlutterPlatformMessageResponseHandle **reply) {
  return FlutterPlatformMessageCreateResponseHandle(
      engine, proxy_desktop_binary_reply, user_data, reply);
}
