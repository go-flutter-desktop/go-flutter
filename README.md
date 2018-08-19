<img src="./assets/mascot.png" width="170" align="right">

# Go Flutter desktop embedder 

A Go (golang) [Custom Flutter Engine
Embedder](https://github.com/flutter/engine/wiki/Custom-Flutter-Engine-Embedders)
for desktop

# Purpose
This project doesn't compete with
[this](https://github.com/google/flutter-desktop-embedding) awesome one.
The purpose of this project is to support the 
[Flutter](https://github.com/flutter/flutter) framework on Windows, macOS, and
Linux using a **SINGLE** code base.  

[**GLFW**](https://github.com/go-gl/glfw) fits the job because it
provides the right abstractions over the OpenGL's Buffer/mouse/keyboard for each platform.  

The choice of [Golang](https://github.com/golang/go) comes from the fact that it
has the same tooling on every platform.  
Plus golang is a great language because it keeps everything simple and readable,
which, I hope, will encourage people to contribute :grin:.


## How to setup

Read first: [go-gl/glfw](https://github.com/go-gl/glfw/)  

<details>
<summary> :package: :penguin: Linux</summary>
<h4>From binaries</h4>
Check out the <a href="https://github.com/Drakirus/go-flutter-desktop-embedder/releases">Release</a> page for prebuilt versions.

<h4>From source</h4>

```bash
# Clone
git clone https://github.com/Drakirus/Go-Flutter-desktop-embedder.git
cd Go-Flutter-desktop-embedder

# Download the share library
wget https://storage.googleapis.com/flutter_infra/flutter/1ed25ca7b7e3e3e8047df050bba4174074c9b336/linux-x64/linux-x64-embedder \
  -O temp.zip; unzip temp.zip; 

# Move the share library
mv libflutter_engine.so ./flutter/library/linux/

# Clean-up
rm flutter_embedder.h; rm temp.zip

# build the Embedder
go get -u github.com/go-gl/glfw/v3.2/glfw
go build

# build the flutter project
cd flutter_project/stocks/
flutter build bundle
cd ../..

# Play
./Go-Flutter-desktop-embedder
```
</details>

## Flutter Project

The example project is available [here](./flutter_project/stocks/) _(from the official flutter repo)_

