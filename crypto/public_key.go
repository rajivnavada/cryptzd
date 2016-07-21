package crypto

import (
	"github.com/rajivnavada/gpgme"
	"time"
)

type defaultPublicKeyCore struct {
	Id          int       `db:"id"`
	UserId      int       `db:"user_id"`
	Fingerprint string    `db:"fingerprint"`
	KeyData     []byte    `db:"key_data"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	ActivatedAt time.Time `db:"activated_at"`
	ExpiresAt   time.Time `db:"expires_at"`
}

func (dpkc *defaultPublicKeyCore) reloadFromDataStore(sess Session) error {
	return NotImplementedError
}

func (dpkc *defaultPublicKeyCore) mergeWith(saved *defaultPublicKeyCore) {
	return NotImplementedError
}

func (dpkc *defaultPublicKeyCore) Save(sess Session) error {
	return NotImplementedError
}

func (dpkc *defaultPublicKeyCore) Encrypt(msg string) (string, error) {
	cipher, err := gpgme.EncryptMessage(msg, dpkc.Fingerprint)
	if err != nil {
		return "", err
	}
	return cipher, nil
}

func (dpkc *defaultPublicKeyCore) EncryptAndSave(s, subject, sender string) (Message, error) {
	cipher, err := gpgme.EncryptMessage(s, dpkc.Fingerprint)
	if err != nil {
		return nil, err
	}

	// TODO: save message
	return nil, nil
}

func (dpkc *defaultPublicKeyCore) Activate() error {
	if dpkc.ActivatedAt.IsZero() {
		dpkc.ActivatedAt = time.Now().UTC()
	}
	// TODO: save the entity
	return nil
}

func (dpkc *defaultPublicKeyCore) Messages() MessageCollection {
	return nil
}

type defaultPublicKey struct {
	*defaultPublicKeyCore
}

func (k defaultPublicKey) Id() int {
	return k.defaultPublicKeyCore.Id
}

func (k defaultPublicKey) UserId() int {
	return k.defaultPublicKeyCore.UserId
}

func (k defaultPublicKey) Fingerprint() string {
	return k.defaultPublicKeyCore.Fingerprint
}

func (k defaultPublicKey) KeyData() string {
	return k.defaultPublicKeyCore.KeyData
}

func (k defaultPublicKey) Active() bool {
	return !k.defaultPublicKeyCore.ActivatedAt.IsZero()
}

func (k defaultPublicKey) CreatedAt() time.Time {
	return k.defaultPublicKeyCore.CreatedAt
}

func (k defaultPublicKey) UpdatedAt() time.Time {
	return k.defaultPublicKeyCore.UpdatedAt
}

func (k defaultPublicKey) ActivatedAt() time.Time {
	return k.defaultPublicKeyCore.ActivatedAt
}

func (k defaultPublicKey) ExpiresAt() time.Time {
	return k.defaultPublicKeyCore.ExpiresAt
}

func (k defaultPublicKey) User() User {
	u, err := FindUserWithId(k.defaultPublicKeyCore.UserId)
	if err != nil {
		log.Println("An error occured when trying to retreive user")
		log.Println(err)
		return nil
	}
	return u
}

func FindKeyWithId(id int) (Key, error) {
	k := &defaultPublicKey{&defaultPublicKeyCore{Id: id}}
	if err := find(k); err != nil {
		return nil, err
	}
	return k, nil
}

func FindKeyWithFingerprint(fingerprint string) (Key, error) {
	k := &defaultPublicKey{&defaultPublicKeyCore{Fingerprint: fingerprint}}
	if err := find(k); err != nil {
		return nil, err
	}
	return k, nil
}
