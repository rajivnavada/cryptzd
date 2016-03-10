package testingcgo

// #include <stdlib.h>
// #include "messenger.h"
import "C"
import "unsafe"

// GetGreeting returns a greeting by leveraging a C++ library
func GetGreeting(name string) string {
	rawMessage := C.prepareMessage(C.CString(name))
	msg := C.GoString(rawMessage)
	C.free(unsafe.Pointer(rawMessage))
	return msg
}
