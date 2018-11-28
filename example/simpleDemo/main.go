package main

import (
	"bufio"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	gutter "github.com/Drakirus/go-flutter-desktop-embedder"
	"github.com/Drakirus/go-flutter-desktop-embedder/flutter"

	"github.com/go-gl/glfw/v3.2/glfw"
)

func main() {
	var (
		err error
	)

	_, currentFilePath, _, _ := runtime.Caller(0)
	dir := path.Dir(currentFilePath)

	options := []gutter.Option{
		gutter.OptionAssetPath(dir + "/flutter_project/demo/build/flutter_assets"),
		/* Depending on your architecture you need to change the enginer
		 * Mac OS X : flutter/bin/cache/artifacts/engine/darwin-x64/icudtl.dat
		 * Linux    : flutter/bin/cache/artifacts/engine/linux-x64/icudtl.dat
		 * Windows  : flutter/bin/cache/artifacts/engine/windows-x64/icudtl.dat
		 */
		gutter.OptionICUDataPath(dir + "/icudtl.dat"),
		gutter.OptionWindowInitializer(setIcon),
		gutter.OptionWindowDimension(800, 600),
		gutter.OptionWindowInitializer(setIcon),
		gutter.OptionPixelRatio(1.2),
		gutter.OptionVMArguments([]string{"--dart-non-checked-mode", "--observatory-port=50300"}),
		gutter.OptionAddPluginReceiver(ownPlugin, "plugin_demo"),
	}

	if err = gutter.Run(options...); err != nil {
		log.Fatalln(err)
	}

}

func setIcon(window *glfw.Window) error {
	_, currentFilePath, _, _ := runtime.Caller(0)
	dir := path.Dir(currentFilePath)
	imgFile, err := os.Open(dir + "/assets/icon.png")
	if err != nil {
		return err
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return err
	}
	window.SetIcon([]image.Image{img})
	return nil
}

// Plugin that read the stdin and send the number to the dart side
func ownPlugin(
	platMessage *flutter.PlatformMessage,
	flutterEngine *flutter.EngineOpenGL,
	window *glfw.Window,
) bool {
	if platMessage.Message.Method != "getNumber" {
		log.Printf("Unhandled platform method: %#v from channel %#v\n",
			platMessage.Message.Method, platMessage.Channel)
		return false
	}

	time.Sleep(1 * time.Second)
	go func() {
		fmt.Printf("Reading (A number): ")
		reader := bufio.NewReader(os.Stdin)
		s, _ := reader.ReadString('\n')
		s = strings.Trim(s, " \n")
		if _, err := strconv.Atoi(s); err == nil {
			flutterEngine.SendPlatformMessageResponse(platMessage, []byte("[ "+s+" ]"))
		} else {
			fmt.Printf(" ,%q don't looks like a number.\n", s)
		}
	}()

	return true

}
