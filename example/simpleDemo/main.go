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

	initialApplicationHeight := 600
	initialApplicationWidth := 800

	options := []gutter.Option{
		gutter.ProjectAssetPath(dir + "/flutter_project/demo/build/flutter_assets"),
		/* This path should not be changed. icudtl.dat is handled by engineDownloader.go */
		gutter.ApplicationICUDataPath(dir + "/icudtl.dat"),
		gutter.ApplicationWindowDimension(initialApplicationWidth, initialApplicationHeight),
		gutter.OptionWindowInitializer(setIcon),
		gutter.OptionVMArguments([]string{"--dart-non-checked-mode", "--observatory-port=50300"}),
		gutter.OptionAddPluginReceiver(ownPlugin, "plugin_demo"),
		// Default keyboard is Qwerty, if you want to change it, you can check keyboard.go in gutter package.
		// Otherwise you can create your own by usinng `KeyboardShortcuts` struct.
		//gutter.OptionKeyboardLayout(gutter.KeyboardAzertyLayout),
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
		s = strings.TrimRight(s, "\r\n")
		if _, err := strconv.Atoi(s); err == nil {
			flutterEngine.SendPlatformMessageResponse(platMessage, []byte("[ "+s+" ]"))
		} else {
			fmt.Printf(" ,%q don't looks like a number.\n", s)
		}
	}()

	return true

}
