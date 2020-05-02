// Package currentthread gives you access to the underlying thread id.
package currentthread

// //
// // Extracted from TinyCThread, a minimalist, portable, threading library for C
// //
//
// /* Platform specific includes */
// #if defined(_WIN32) || defined(__WIN32__) || defined(__WINDOWS__)
//   #include <windows.h>
//   typedef HANDLE thrd_t;
// #else
//   #include <pthread.h>
//   typedef pthread_t thrd_t;
// #endif
//
// thrd_t thrd_current(void) {
//   #if defined(_WIN32) || defined(__WIN32__) || defined(__WINDOWS__)
//     return GetCurrentThread();
//   #else
//     return pthread_self();
//   #endif
// }
// size_t getCurrentThreadID() { return (size_t)thrd_current(); }
import "C"
import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
)

// ID returns the id of the current thread
func ID() int64 {
	fmt.Println("Real Thread id: " + spew.Sdump(C.thrd_current()))
	return (int64)(C.getCurrentThreadID())
}
