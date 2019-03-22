package main

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-flutter-desktop/go-flutter"
)

// VMArguments may be set by hover at compile-time
var VMArguments string

func main() {
	// DO NOT EDIT, add options in options.go
	mainOptions := []flutter.Option{
		flutter.OptionVMArguments(strings.Split(VMArguments, ";")),
		flutter.WindowIcon(iconProvider),
	}
	err := flutter.Run(append(options, mainOptions...)...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func iconProvider() ([]image.Image, error) {
	imgFile, err := os.Open(filepath.Join("assets", "logo.png"))
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, err
	}
	return []image.Image{img}, nil
}
