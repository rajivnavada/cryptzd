package crypto

// #cgo CFLAGS: -DISCGO=1 -Wall
// #cgo CPPFLAGS: -DISCGO=1
// #cgo LDFLAGS: -lgpg-error -lassuan -lgpgme
// #include <stdlib.h>
// #include "gpgme-bridge.h"
import "C"
import (
	"strings"
	"time"
	"unsafe"
)

func importPublicKey(s string, bk *baseKey, bu *baseUser) error {
	// Get a keyInfo object
	var keyInfo *C.struct_key_info = C.new_key_info()
	// Free keyData since we use CString to allocate it
	defer C.free_key_info(keyInfo)

	// Convert passed in public key to a C char array
	keyData := C.CString(s)
	defer C.free(unsafe.Pointer(keyData))

	// Now perform the import
	C.import_key(keyInfo, keyData, C.size_t(len(s)))

	// Handle key info

	// fingerprint is a character array
	fingerprint := C.GoStringN(&keyInfo.fingerprint[0], C.KEY_FINGERPRINT_LEN)
	if fingerprint == "" {
		return InvalidKeyError
	}
	bk.Fingerprint = fingerprint

	if keyInfo.expires > 0 {
		bk.ExpiresAt = time.Unix(int64(keyInfo.expires), 0)
	}

	// Now handle the user info

	emailLen := C.int(C.strlen(&keyInfo.user_email[0]))
	nameLen := C.int(C.strlen(&keyInfo.user_name[0]))
	commentLen := C.int(C.strlen(&keyInfo.user_comment[0]))

	if emailLen == 0 {
		return MissingEmailError
	}

	email := C.GoStringN(&keyInfo.user_email[0], emailLen)
	if email == "" {
		return MissingEmailError
	}
	bu.Email = email

	bu.Name = strings.TrimSpace(C.GoStringN(&keyInfo.user_name[0], nameLen))
	bu.Comment = strings.TrimSpace(C.GoStringN(&keyInfo.user_comment[0], commentLen))

	return nil
}
