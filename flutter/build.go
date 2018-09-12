package flutter

/*
// Linux Build Tags
// ----------------
#cgo linux CFLAGS: -I${SRCDIR}/library
#cgo linux LDFLAGS: -lflutter_engine -Wl,-rpath,$ORIGIN/flutter/library/linux

// Windows Build Tags
// ----------------
#cgo windows CFLAGS: -I${SRCDIR}/library
#cgo windows LDFLAGS: -L${SRCDIR}/library/windows -lflutter_engine

*/
import "C"

import (
	// prevents dep from stripping out the c source files in flutter library.
	_ "github.com/Drakirus/go-flutter-desktop-embedder/flutter/library"
)
