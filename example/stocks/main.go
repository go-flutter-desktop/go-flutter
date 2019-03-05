package main

import (
	"image"
	_ "image/png"
	"log"
	"os"

	gutter "github.com/go-flutter-desktop/go-flutter"
)

func main() {
	var (
		err error
	)

	options := []gutter.Option{
		gutter.ProjectAssetPath("flutter_project/stocks/build/flutter_assets"),
		gutter.OptionICUDataPath("/opt/flutter/bin/cache/artifacts/engine/linux-x64/icudtl.dat"), // Linux (arch)
		// gutter.OptionICUDataPath("./FlutterEmbedder.framework/Resources/icudtl.dat"),             // OSX
		gutter.OptionWindowDimension(800, 600),
		gutter.WindowIcon(iconProvider),
		gutter.OptionPixelRatio(1.2),
		gutter.OptionVMArguments([]string{"--dart-non-checked-mode", "--observatory-port=50300"}),
	}

	if err = gutter.Run(options...); err != nil {
		log.Fatalln(err)
	}

}

func iconProvider() ([]image.Image, error) {
	imgFile, err := os.Open("assets/icon.png")
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, err
	}

	return []image.Image{img}, nil
}
