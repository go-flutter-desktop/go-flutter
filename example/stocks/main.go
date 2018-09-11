package main

import (
	"image"
	_ "image/png"
	"log"
	"os"
	"os/signal"

	gutter "github.com/Drakirus/go-flutter-desktop-embedder"

	"github.com/go-gl/glfw/v3.2/glfw"
)

// #cgo linux LDFLAGS: -L${SRCDIR}/flutter/library/linux
import "C"

func main() {
	var (
		err error
	)

	options := []gutter.Option{
		gutter.OptionAssetPath("flutter_project/stocks/build/flutter_assets"),
		gutter.OptionICUDataPath("../../flutter/library/icudtl.dat"),
		gutter.OptionWindowInitializer(setIcon),
	}

	if err = gutter.Run(options...); err != nil {
		log.Fatalln(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Kill, os.Interrupt)
	select {
	case <-signals:
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
