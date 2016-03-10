#pragma once

#include <stdlib.h>
#include <string.h>
#include <gpgme.h>
#include <gpg-error.h>

// Minimum version of GPGME we'll accept
static const char *GPGME_MIN_VERSION = "1.6.0";

static const int KEY_FINGERPRINT_LEN = 40;
static const int KEY_USERNAME_LEN = 255;
static const int KEY_USEREMAIL_LEN = 255;
static const int KEY_USERCOMMENT_LEN = 255;

// +1 for the terminating 0
typedef struct key_info {
    long int expires;
    char user_name[KEY_USERNAME_LEN+1];
    char user_email[KEY_USEREMAIL_LEN+1];
    char user_comment[KEY_USERCOMMENT_LEN+1];
    char fingerprint[KEY_FINGERPRINT_LEN+1];
    int is_new;
} *key_info_t;

#ifdef __cplusplus
extern "C" {
#endif

    // Returns an instance of key_info
    key_info_t new_key_info ();

    // Frees memory allocation to INFO
    void free_key_info (key_info_t info);

    // Tries to import KEY into the system keychain
    void import_key (key_info_t info, const char *key);

    void get_key_info (key_info_t info, const char *fingerprint, gpgme_ctx_t ctx);

    // Returns encrypted data that MUST be freed by the caller
    char *encrypt (const char *fingerprint, const char *message);

    void free_cipher_text (char *cipher_text);

#ifdef __cplusplus
}
#endif


