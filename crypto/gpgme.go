package crypto

// #cgo CFLAGS: -DISCGO=1 -I/usr/local/include -I/usr/include -Wall
// #cgo CPPFLAGS: -DISCGO=1 -I/usr/local/include -I/usr/include
// #cgo LDFLAGS: -L/usr/local/lib -L/usr/lib -lgpg-error -lassuan -lgpgme
// #include <stdlib.h>
// #include "gpgme-bridge.h"
import "C"
import (
	"errors"
	"sync"
	"time"
	"unsafe"
)

var (
	FailedEncryptionError = errors.New("Failed to encrypt message.")
	FailedDecryptionError = errors.New("Failed to decrypt message.")

	importPublicKeyLock = &sync.Mutex{}
	encryptLock         = &sync.Mutex{}
	decryptLock         = &sync.Mutex{}
)

func importPublicKey(s string, bk *baseKey, bu *baseUser) error {
	// Get a keyInfo object
	var keyInfo *C.struct_key_info = C.new_key_info()
	// Free keyData since we use CString to allocate it
	defer C.free_key_info(keyInfo)

	// Convert passed in public key to a C char array
	keyData := C.CString(s)
	defer C.free(unsafe.Pointer(keyData))

	// Now perform the import by protecting with a lock
	importPublicKeyLock.Lock()
	C.import_key(keyInfo, keyData)
	importPublicKeyLock.Unlock()

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

	bu.Name = C.GoStringN(&keyInfo.user_name[0], nameLen)
	bu.Comment = C.GoStringN(&keyInfo.user_comment[0], commentLen)

	return nil
}

func encryptMessage(message, fingerprint string) (string, error) {
	// Get the fingerprint as a C string
	fpr := C.CString(fingerprint)
	defer C.free(unsafe.Pointer(fpr))

	// Get message as a C string
	msg := C.CString(message)
	defer C.free(unsafe.Pointer(msg))

	// Call into C to encrypt
	encryptLock.Lock()
	cipher := C.encrypt(fpr, msg)
	encryptLock.Unlock()
	if cipher == nil {
		return "", FailedEncryptionError
	}
	defer C.free_cipher_text(cipher)

	output := C.GoString(cipher)
	if output == "" {
		return "", FailedEncryptionError
	}

	return output, nil
}

func DecryptMessage(encryptedMessage string) (string, error) {
	// Get message as a C string
	msg := C.CString(encryptedMessage)
	defer C.free(unsafe.Pointer(msg))

	// Call into C to encrypt
	decryptLock.Lock()
	decryptedMessage := C.decrypt(msg)
	decryptLock.Unlock()
	if decryptedMessage == nil {
		return "", FailedDecryptionError
	}
	defer C.free_cipher_text(decryptedMessage)

	output := C.GoString(decryptedMessage)
	if output == "" {
		return "", FailedDecryptionError
	}

	return output, nil
}
