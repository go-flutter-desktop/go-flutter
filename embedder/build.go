package embedder

// #cgo linux LDFLAGS: -lflutter_engine -Wl,-rpath,$ORIGIN
// #cgo windows LDFLAGS: -lflutter_engine
// #cgo darwin LDFLAGS: -framework FlutterEmbedder
import "C"
