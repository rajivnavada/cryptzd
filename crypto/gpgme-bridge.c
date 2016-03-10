#include "gpgme-bridge.h"


static const int COPY = 1;


int init_gpgme ()
{
    static int initialized = 0;

    // Initialize GPGME if not initialized yet
    if (!initialized)
    {
        const char *version = gpgme_check_version (GPGME_MIN_VERSION);
        initialized = !version ? 0 : 1;
    }

    return initialized;
}


key_info_t new_key_info ()
{
    key_info_t info = (key_info_t) malloc (sizeof (struct key_info));
    (void) memset (info, 0, sizeof (struct key_info));
    return info;
}


void free_key_info (key_info_t info)
{
    free (info);
    info = NULL;
}


void import_key (key_info_t info, const char *key, size_t key_size)
{
    // Variables that MUST be freed before returning
    gpgme_ctx_t ctx = NULL;
    gpgme_data_t key_data = NULL;

    // These should be variables managed by the context
    gpgme_import_result_t import_result = NULL;
    gpgme_import_status_t status = NULL;

    gpgme_error_t err;

    // First thing to do is INIT
    init_gpgme ();

    // Let's setup a gpgme context
    err = gpgme_new (&ctx);
    if (gpg_err_code (err) != GPG_ERR_NO_ERROR)
        goto free_resources_and_return;

    // Construct a gpgme_data_t instance from passed in key data
    // NOTE: we ask to copy since we don't want to mess up Go's memory manager
    err = gpgme_data_new_from_mem (&key_data, key, key_size, COPY);
    if (gpg_err_code (err) != GPG_ERR_NO_ERROR)
        goto free_resources_and_return;

    // Now we get key info
    err = gpgme_op_import (ctx, key_data);
    if (gpg_err_code (err) != GPG_ERR_NO_ERROR)
        goto free_resources_and_return;

    import_result = gpgme_op_import_result (ctx);
    if (!import_result || !import_result->imports)
        goto free_resources_and_return;

    // We'll only consider the first result
    status = import_result->imports;

    // Now pull the fingerprint from status and get full description of the key
    get_key_info (info, status->fpr, ctx);

    // Adding this here so that memory can be zeroed in get_key_info
    if (status->status&GPGME_IMPORT_NEW)
        info->is_new = 1;

free_resources_and_return:
    // Release all resources
    if (key_data)
        gpgme_data_release (key_data);
    if (ctx)
        gpgme_release (ctx);
}


// Get information about a key and inserts data into the KEY_INFO.
// If ctx is NULL, a new ctx will be created.
void get_key_info (key_info_t info, const char *fingerprint, gpgme_ctx_t ctx)
{
    if (!fingerprint || !info)
        return;

    gpgme_error_t err;
    int created_ctx = 0;

    if (!ctx)
    {
        init_gpgme ();
        err = gpgme_new (&ctx);
        if (gpg_err_code (err) != GPG_ERR_NO_ERROR)
            return;
        created_ctx = 1;
    }

    gpgme_key_t key = NULL;
    err = gpgme_get_key (ctx, fingerprint, &key, 0);
    if (!key || !key->subkeys || !key->uids)
        goto free_resources_and_return;

    // NOTE: assuming that the key_info will always be zeroed out
    // Copy the strings
    (void) strncpy (info->fingerprint, key->subkeys->fpr, KEY_FINGERPRINT_LEN);
    (void) strncpy (info->user_name, key->uids->name, KEY_USERNAME_LEN);
    (void) strncpy (info->user_email, key->uids->email, KEY_USEREMAIL_LEN);
    (void) strncpy (info->user_comment, key->uids->comment, KEY_USERCOMMENT_LEN);

    // Copy the expires timestamp
    info->expires = key->subkeys->expires;

    // In this function, is_new will always be set to false
    info->is_new = 0;

free_resources_and_return:
    if (created_ctx)
    {
        gpgme_release (ctx);
        ctx = NULL;
    }
}

