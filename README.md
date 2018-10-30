<img src="./mascot.png" width="170" align="right">

# Go Flutter desktop embedder 

[![Join the chat at https://gitter.im/go-flutter-desktop-embedder/Lobby](https://badges.gitter.im/go-flutter-desktop-embedder/Lobby.svg)](https://gitter.im/go-flutter-desktop-embedder/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
<a href="https://stackoverflow.com/questions/tagged/flutter?sort=votes">
   <img alt="Awesome Flutter" src="https://img.shields.io/badge/Awesome-Flutter-blue.svg?longCache=true&style=flat-square" />
</a>

A Go (golang) [Custom Flutter Engine
Embedder](https://github.com/flutter/engine/wiki/Custom-Flutter-Engine-Embedders)
for desktop

# Purpose
This project doesn't compete with
[this](https://github.com/google/flutter-desktop-embedding) awesome one.
The purpose of this project is to support the 
[Flutter](https://github.com/flutter/flutter) framework on Windows, MacOS, and
Linux using a **SINGLE** code base.  

[**GLFW**](https://github.com/go-gl/glfw) fits the job because it
provides the right abstractions over the OpenGL's Buffer/mouse/keyboard for each platform.  

The choice of [Golang](https://github.com/golang/go) comes from the fact that it
has the same tooling on every platform.  
Plus golang is a great language because it keeps everything simple and readable,
which, I hope, will encourage people to contribute :grin:.

## How to install

For this Flutter version:
```
$ flutter --version
Flutter 0.7.3 • channel beta • https://github.com/flutter/flutter.git
Framework • revision 3b309bda07 (2 weeks ago) • 2018-08-28 12:39:24 -0700
Engine • revision af42b6dc95
Tools • Dart 2.1.0-dev.1.0.flutter-ccb16f7282
```

<details>
<summary> :package: :penguin: Linux</summary>
<h4>From binaries</h4>
Check out the <a href="https://github.com/Drakirus/go-flutter-desktop-embedder/releases">Release</a> page for prebuilt versions.

<h4>From source</h4>

Go read first: [go-gl/glfw](https://github.com/go-gl/glfw/)  


```bash
# Clone
git clone https://github.com/Drakirus/go-flutter-desktop-embedder.git
cd go-flutter-desktop-embedder

# Build the flutter simpleDemo project
cd example/simpleDemo/
cd flutter_project/demo/
flutter build bundle
cd ../..

# Download the share library (CORRESPONDING to the Flutter's version shown above)
wget https://storage.googleapis.com/flutter_infra/flutter/af42b6dc95bd9f719e43c4e9f29a00640f0f0bba/linux-x64/linux-x64-embedder -O .build/temp.zip

# Extract the share library
unzip .build/temp.zip -x flutter_embedder.h

# REQUIRED: When using `go build` or `go run main.go`, the go library need to know where to look for the share library
export CGO_LDFLAGS="-L${PWD}"

# If you `go build`, the share library must stay in the same path, relative to the go binary

# Get the libraries
go get -u -v github.com/Drakirus/go-flutter-desktop-embedder

# Make sure the path in "main.go" to the `icudtl.dat` is correct.
# Build the example project
go build
```

</details>

<details>
<summary> :package: :checkered_flag: Windows</summary>
<h4>From binaries</h4>
Check out the <a href="https://github.com/Drakirus/go-flutter-desktop-embedder/releases">Release</a> page for prebuilt versions.

<h4>From source</h4>

Go read first: [go-gl/glfw](https://github.com/go-gl/glfw/)  


```bash
# Clone
git clone https://github.com/Drakirus/go-flutter-desktop-embedder.git
cd go-flutter-desktop-embedder

# Build the flutter simpleDemo project
cd example/simpleDemo/
cd flutter_project/demo/
flutter build bundle
cd ../..

# Download the share library (CORRESPONDING to the Flutter's version shown above)
# => https://storage.googleapis.com/flutter_infra/flutter/af42b6dc95bd9f719e43c4e9f29a00640f0f0bba/windows-x64/windows-x64-embedder.zip

# Move the share library
# => "flutter_engine.dll" must be in the flutter example project (where the main.go is)

# REQUIRED: When using `go build` or `go run main.go`, the go library need to know where to look for the share library
set CGO_LDFLAGS=-L%cd%

# If you `go build`, the share library must stay in the same path, relative to the go binary

# Get the libraries
go get -u -v github.com/Drakirus/go-flutter-desktop-embedder

# Make sure the path in "main.go" to the `icudtl.dat` is correct.
# Build or Run the example project
go run main.go
```

</details>

<details>
<summary> :package: :apple: MacOS</summary>
<h4>From binaries</h4>
Check out the <a href="https://github.com/Drakirus/go-flutter-desktop-embedder/releases">Release</a> page for prebuilt versions.

<h4>From source</h4>

Go read first: [go-gl/glfw](https://github.com/go-gl/glfw/)  


```bash
# Clone
git clone https://github.com/Drakirus/go-flutter-desktop-embedder.git
cd go-flutter-desktop-embedder

# Build the flutter simpleDemo project
cd example/simpleDemo/
cd flutter_project/demo/
flutter build bundle
cd ../..

# Download the share library (CORRESPONDING to the Flutter's version shown above)
wget https://storage.googleapis.com/flutter_infra/flutter/af42b6dc95bd9f719e43c4e9f29a00640f0f0bba/darwin-x64/FlutterEmbedder.framework.zip -O .build/temp.zip

# Move the share library
unzip .build/temp.zip -d .build && unzip .build/FlutterEmbedder.framework.zip -d .build/FlutterEmbedder.framework
mv .build/FlutterEmbedder.framework .

# REQUIRED: When using `go build` or `go run main.go`, the go library need to know where to look for the share library
export CGO_LDFLAGS="-F${PWD} -Wl,-rpath,@executable_path"

# If you `go build`, the share library must stay in the same path, relative to the go binary

# Get the libraries
go get -u -v github.com/Drakirus/go-flutter-desktop-embedder

# Make sure the path in "main.go" to the `icudtl.dat` is correct.
# Build the example project
go build
```

</details>


## Flutter Demos Projects

The examples are available [here](./example/).

<img src="./stocks.jpg" width="900" align="center" alt="Screenshot of the Stocks demo app on macOS">

## Support

- [x] Linux :penguin:
- [x] Windows :checkered_flag:
- [x] MacOS :apple:
- [x] Text input
- [ ] Plugins
- [x] Importable go library
- [x] Clipboard (through shortcuts)
- [ ] Clipboard (through the click)
- [x] Keyboard shortcuts
   - [x] <kbd>ctrl-c</kbd>  <kbd>ctrl-v</kbd>  <kbd>ctrl-x</kbd>  <kbd>ctrl-a</kbd>
   - [x] <kbd>Home</kbd>  <kbd>End</kbd>  <kbd>shift-Home</kbd>  <kbd>shift-End</kbd>
   - [x] <kbd>Left</kbd>  <kbd>ctrl-Left</kbd>  <kbd>ctrl-shift-Left</kbd>
   - [x] <kbd>Right</kbd>  <kbd>ctrl-Right</kbd>  <kbd>ctrl-shift-Right</kbd>
   - [x] <kbd>Backspace</kbd>  <kbd>ctrl-Backspace</kbd> <kbd>Delete</kbd>
   - [ ] <kbd>ctrl-Delete</kbd>
- [ ] Key events
