# embedder

This package wraps the [Flutter embedder API](https://raw.githubusercontent.com/flutter/engine/master/shell/platform/embedder/embedder.h) in Go and adds some helper functions to work with it. This package does NOT contain any platform specific code (glfw, etc.) and may eventualy be used for platforms other than the ones targeted by go-flutter right now. Note that this package does not promise a stable API: types, functions, methods may all change in a breaking way.

## Build

To build this package set the `CGO_LDFLAGS` and run `go build`. For example:

```bash
export CGO_LDFLAGS="-L/home/${HOME}/.cache/hover/engine/linux/"
go build
```

This works if [hover](https://github.com/go-flutter-desktop/hover) has cached the flutter engine for the local user.
