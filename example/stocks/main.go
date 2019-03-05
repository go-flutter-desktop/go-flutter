package main

import (
	"image"
	_ "image/png"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/go-flutter-desktop/go-flutter"
)

func iconProvider() ([]image.Image, error) {
	_, currentFilePath, _, _ := runtime.Caller(0)
	dir := path.Dir(currentFilePath)
	imgFile, err := os.Open(dir + "/assets/icon.png")
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, err
	}

	return []image.Image{img}, nil
}

func main() {

	_, currentFilePath, _, _ := runtime.Caller(0)
	dir := path.Dir(currentFilePath)

	initialApplicationHeight := 600
	initialApplicationWidth := 800

	options := []flutter.Option{
		flutter.ProjectAssetsPath(dir + "/flutter_project/stocks/build/flutter_assets"),

		// This path should not be changed. icudtl.dat is handled by engineDownloader.go
		flutter.ApplicationICUDataPath(dir + "/icudtl.dat"),

		flutter.ApplicationWindowDimension(initialApplicationWidth, initialApplicationHeight),
		flutter.WindowIcon(iconProvider),
		flutter.OptionVMArguments([]string{
			// "--disable-dart-asserts", // release mode flag
			// "--disable-observatory",
			"--observatory-port=50300",
		}),

		// Default keyboard is Qwerty, if you want to change it, you can check keyboard.go in gutter package.
		// Otherwise you can create your own by usinng `KeyboardShortcuts` struct.
		//flutter.OptionKeyboardLayout(flutter.KeyboardAzertyLayout),
	}

	if err := flutter.Run(options...); err != nil {
		log.Fatalln(err)
	}

}
