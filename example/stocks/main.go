package main

import (
	"flag"
	"image"
	_ "image/png"
	"log"
	"os"

	gutter "github.com/Drakirus/go-flutter-desktop-embedder"

	"github.com/go-gl/glfw/v3.2/glfw"
)

func main() {
	var (
		err         error
		flutter_prj string
	)

	flag.StringVar(&flutter_prj, "fp", "stocks", "flutter project's filename")
	flag.Parse()

	options := []gutter.Option{
		gutter.OptionAssetPath("flutter_project" + flutter_prj + "build/flutter_assets"),
		gutter.OptionICUDataPath("icudtl.dat"),
		gutter.OptionWindowInitializer(setIcon),
		gutter.OptionPixelRatio(1.9),
		gutter.OptionVmArguments([]string{"--dart-non-checked-mode", "--observatory-port=50300"}),
	}

	if err = gutter.Run(options...); err != nil {
		log.Fatalln(err)
	}

}

func setIcon(window *glfw.Window) error {
	imgFile, err := os.Open("assets/icon.png")
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
