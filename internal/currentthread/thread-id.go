// Package currentthread gives you access to the underlying C thread id.
package currentthread

// //
// // Extracted from TinyCThread, a minimalist, portable, threading library for C
// //
//
// /* Platform specific includes */
// #if defined(_WIN32) || defined(__WIN32__) || defined(__WINDOWS__)
//   #include <windows.h>
//   typedef DWORD thrd_t;
// #else
//   #include <pthread.h>
//   typedef pthread_t thrd_t;
// #endif
//
// int thrd_equal(thrd_t thr0, thrd_t thr1) {
// #if defined(_WIN32) || defined(__WIN32__) || defined(__WINDOWS__)
//   return thr0 == thr1;
// #else
//   return pthread_equal(thr0, thr1);
// #endif
// }
//
// thrd_t thrd_current(void) {
//   #if defined(_WIN32) || defined(__WIN32__) || defined(__WINDOWS__)
//     return GetCurrentThreadId();
//   #else
//     return pthread_self();
//   #endif
// }
import "C"

// ThreadID correspond to an opaque thread identifier
type ThreadID C.thrd_t

// ID returns the id of the current thread
func ID() ThreadID {
	return (ThreadID)(C.thrd_current())
}

// Equal compares two thread identifiers.
func Equal(t1, t2 ThreadID) bool {
	return C.thrd_equal((C.thrd_t)(t1), (C.thrd_t)(t2)) != 0
}
