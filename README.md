<img src="./mascot.png" width="170" align="right">

# Go Flutter desktop embedder 

[![Join the chat at https://gitter.im/go-flutter-desktop-embedder/Lobby](https://badges.gitter.im/go-flutter-desktop-embedder/Lobby.svg)](https://gitter.im/go-flutter-desktop-embedder/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
<a href="https://stackoverflow.com/questions/tagged/flutter?sort=votes">
   <img alt="Awesome Flutter" src="https://img.shields.io/badge/Awesome-Flutter-blue.svg?longCache=true&style=flat-square" />
</a>

A Go (golang) [Custom Flutter Engine
Embedder](https://github.com/flutter/flutter/wiki/Custom-Flutter-Engine-Embedders)
for desktop

# Purpose

> Disclaimer: Despite being very similar in the end goal, This project doesn't compete with
> [this](https://github.com/google/flutter-desktop-embedding) awesome one.

The flutter engine itself doesn't know how to deal with desktop platforms _(eg handling mouse/keyboard input)_. Instead, it exposes an abstraction layer for whatever platform to implement. The purpose of this project is to implement [Flutter's Embedding API](https://github.com/flutter/engine/blob/master/shell/platform/embedder/embedder.h) using a **SINGLE** code base that runs on Windows, MacOS, and Linux.  [**GLFW**](https://github.com/go-gl/glfw) fits the job because it provides the right abstractions over the OpenGL's Buffer/Mouse/Keyboard for each platform. 

The choice of [Golang](https://github.com/golang/go) comes from the fact that it has the same tooling on every platform. Plus golang is a great language because it keeps everything simple and readable, which, I hope, will encourage people to contribute :grin:.

## How to install

<details>
<summary> :package: :penguin: Linux</summary>
<h4>From binaries</h4>
Check out the <a href="https://github.com/go-flutter-desktop/go-flutter/releases">Release</a> page for prebuilt versions.

<h4>From source</h4>

Go read first: [go-gl/glfw](https://github.com/go-gl/glfw/)  


```bash
# Clone
git clone https://github.com/go-flutter-desktop/go-flutter.git
cd go-flutter-desktop-embedder

# Build the flutter simpleDemo project
cd example/simpleDemo/
cd flutter_project/demo/
flutter build bundle
cd ../..

# Download the share library, the one corresponding to your flutter version.
go run engineDownloader.go

# REQUIRED before every `go build`. The CGO compiler need to know where to look for the share library
export CGO_LDFLAGS="-L${PWD}"
# The share library must stay next to the generated binary.

# Get the libraries
go get -u -v github.com/go-flutter-desktop/go-flutter

# Build the example project
go build main.go

# `go run main.go` is not working ATM.
```

</details>

<details>
<summary> :package: :checkered_flag: Windows</summary>
<h4>From binaries</h4>
Check out the <a href="https://github.com/go-flutter-desktop/go-flutter/releases">Release</a> page for prebuilt versions.

<h4>From source</h4>

Go read first: [go-gl/glfw](https://github.com/go-gl/glfw/)  


```bash
# Clone
git clone https://github.com/go-flutter-desktop/go-flutter.git
cd go-flutter-desktop-embedder

# Build the flutter simpleDemo project
cd example/simpleDemo/
cd flutter_project/demo/
flutter build bundle
cd ../..

# Download the share library, the one corresponding to your flutter version.
go run engineDownloader.go

# REQUIRED before every `go build`. The CGO compiler need to know where to look for the share library
set CGO_LDFLAGS=-L%cd%
# The share library must stay next to the generated binary.
# If you ran into a MinGW ld error, checkout: https://github.com/go-flutter-desktop/go-flutter/issues/34

# Get the libraries
go get -u -v github.com/go-flutter-desktop/go-flutter

# Build the example project
go build main.go

# `go run main.go` is not working ATM.
```

</details>

<details>
<summary> :package: :apple: MacOS</summary>
<h4>From binaries</h4>
Check out the <a href="https://github.com/go-flutter-desktop/go-flutter/releases">Release</a> page for prebuilt versions.

<h4>From source</h4>

Go read first: [go-gl/glfw](https://github.com/go-gl/glfw/)  


```bash
# Clone
git clone https://github.com/go-flutter-desktop/go-flutter.git
cd go-flutter-desktop-embedder

# Build the flutter simpleDemo project
cd example/simpleDemo/
cd flutter_project/demo/
flutter build bundle
cd ../..

# Download the share library, the one corresponding to your flutter version.
go run engineDownloader.go

# REQUIRED before every `go build`. The CGO compiler need to know where to look for the share library
export CGO_LDFLAGS="-F${PWD} -Wl,-rpath,@executable_path"
# The share library must stay next to the generated binary.

# Get the libraries
go get -u -v github.com/go-flutter-desktop/go-flutter

# Build the example project
go build main.go

# `go run main.go` is not working ATM.
```

</details>



## Flutter Demos Projects

The examples are available [here](./example/).

<img src="./stocks.jpg" width="900" align="center" alt="Screenshot of the Stocks demo app on macOS">

## Support

- [x] Linux :penguin:
- [x] Windows :checkered_flag:
- [x] MacOS :apple:
- [x] Importable go library
- [ ] Plugins [Medium article on how the the Flutter's messaging works](https://medium.com/flutter-io/flutter-platform-channels-ce7f540a104e)
   - [x] JSON MethodChannel
   - [ ] StandardMethodCodec, ...
- [ ] System plugins [Platform channels used by the Flutter system](https://github.com/flutter/flutter/blob/master/packages/flutter/lib/src/services/system_channels.dart)
  - [x] Window Title
  - [x] Text input
  - [x] Clipboard (through shortcuts and UI)
  - [x] Keyboard shortcuts
    - [x] <kbd>ctrl-c</kbd>  <kbd>ctrl-v</kbd>  <kbd>ctrl-x</kbd>  <kbd>ctrl-a</kbd>
    - [x] <kbd>Home</kbd>  <kbd>End</kbd>  <kbd>shift-Home</kbd>  <kbd>shift-End</kbd>
    - [x] <kbd>Left</kbd>  <kbd>ctrl-Left</kbd>  <kbd>ctrl-shift-Left</kbd>
    - [x] <kbd>Right</kbd>  <kbd>ctrl-Right</kbd>  <kbd>ctrl-shift-Right</kbd>
    - [x] <kbd>Backspace</kbd>  <kbd>ctrl-Backspace</kbd> <kbd>Delete</kbd>
    - [ ] <kbd>ctrl-Delete</kbd>
  - [ ] Key events
