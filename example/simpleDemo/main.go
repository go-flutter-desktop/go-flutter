package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-flutter-desktop/go-flutter"
	"github.com/go-flutter-desktop/go-flutter/plugin"
	"github.com/pkg/errors"
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

	options := []flutter.Option{
		flutter.ProjectAssetsPath(dir + "/flutter_project/demo/build/flutter_assets"),

		// This path should not be changed. icudtl.dat is handled by engineDownloader.go
		flutter.ApplicationICUDataPath(dir + "/icudtl.dat"),

		flutter.WindowInitialDimensions(1280, 1024),
		flutter.WindowIcon(iconProvider),
		flutter.OptionVMArguments([]string{
			// "--disable-dart-asserts", // release mode flag
			// "--disable-observatory",
			"--observatory-port=50300",
		}),

		flutter.AddPlugin(&manualInputPlugin{}),

		// Default keyboard is Qwerty, if you want to change it, you can check keyboard.go in gutter package.
		// Otherwise you can create your own by usinng `KeyboardShortcuts` struct.
		// flutter.OptionKeyboardLayout(flutter.KeyboardAzertyLayout),
	}

	if err := flutter.Run(options...); err != nil {
		fmt.Printf("Failed running the Flutter app: %v\n", err)
		os.Exit(1)
	}

}

// Plugin that read the stdin and send the number to the dart side
type manualInputPlugin struct {
	channel *plugin.MethodChannel
}

func (p *manualInputPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	p.channel = plugin.NewMethodChannel(messenger, "plugin_demo", plugin.JSONMethodCodec{})
	p.channel.HandleFunc("getNumber", p.getNumber)
	p.channel.HandleFunc("print", p.print)
	go func() {
		_, err := p.channel.InvokeMethod("submit", "Message from the Go side, it's now: "+time.Now().String())
		if err != nil {
			fmt.Printf("error submitting time: %v", err)
		}
	}()
	return nil
}

func (p *manualInputPlugin) getNumber(arguments interface{}) (reply interface{}, err error) {
	time.Sleep(2 * time.Second)
	for {
		fmt.Printf("Please enter a number: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("failed to read from stdin: %v", err)
			return nil, errors.Wrap(err, "failed to read from stdin")
		}

		input = strings.TrimRight(input, "\r\n")
		number, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Failed to parse number: %v\n", err)
			fmt.Println("Try again")
			continue
		}

		return number, nil
	}
}

func (p *manualInputPlugin) print(arguments interface{}) (reply interface{}, err error) {
	args := struct {
		Textfield string
		Number    int
	}{}
	err = json.Unmarshal(arguments.(json.RawMessage), &args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode arguments")
	}
	fmt.Printf("Textfield: %s\n", args.Textfield)
	fmt.Printf("Number:    %d\n", args.Number)
	return nil, nil
}
