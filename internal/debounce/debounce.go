// Copyright © 2019 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package debounce provides a debouncer func. The most typical use case would be
// the user typing a text into a form; the UI needs an update, but let's wait for
// a break.
//
// Copied from https://github.com/bep/debounce/blob/master/debounce.go
// modified to support an glfw.Window argument.
package debounce

import (
	"sync"
	"time"

	"github.com/go-flutter-desktop/go-flutter/internal/tasker"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// New returns a debounced function that takes another functions as its argument.
// This function will be called when the debounced function stops being called
// for the given duration.
// The debounced function can be invoked with different functions, if needed,
// the last one will win.
func New(after time.Duration, glfwTasker *tasker.Tasker) func(f func(window *glfw.Window), window *glfw.Window) {
	d := &debouncer{after: after, glfwTasker: glfwTasker}

	return func(f func(window *glfw.Window), window *glfw.Window) {
		d.add(f, window)
	}
}

type debouncer struct {
	mu         sync.Mutex
	after      time.Duration
	timer      *time.Timer
	glfwTasker *tasker.Tasker
}

func (d *debouncer) add(f func(window *glfw.Window), window *glfw.Window) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.after, func() {
		// needs to run on main thread
		d.glfwTasker.Do(func() {
			f(window)
		})
	})
}
