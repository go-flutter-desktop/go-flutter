package flutter

/*
// Linux Build Tags
// ----------------
#cgo linux CFLAGS: -I${SRCDIR}/library
#cgo linux LDFLAGS: -lflutter_engine -Wl,-rpath,flutter/library/linux

*/
import "C"

import (
	// prevents dep from stripping out the c source files in flutter library.
	_ "github.com/Drakirus/go-flutter-desktop-embedder/flutter/library"
)
