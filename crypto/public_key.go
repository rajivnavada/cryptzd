package crypto

import (
	"database/sql"
	"github.com/rajivnavada/gpgme"
	"log"
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

type defaultPublicKey struct {
	*defaultPublicKeyCore
}

func (k defaultPublicKey) Id() int {
	return k.defaultPublicKeyCore.Id
}

func (k defaultPublicKey) UserId() int {
	return k.defaultPublicKeyCore.UserId
}

func (k *defaultPublicKey) SetUserId(uid int) {
	k.defaultPublicKeyCore.UserId = uid
}

func (k defaultPublicKey) Fingerprint() string {
	return k.defaultPublicKeyCore.Fingerprint
}

func (k defaultPublicKey) KeyData() []byte {
	return k.defaultPublicKeyCore.KeyData
}

func (k *defaultPublicKey) SetKeyData(d []byte) {
	k.defaultPublicKeyCore.KeyData = d
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

func (k *defaultPublicKey) SetExpiresAt(t time.Time) {
	k.defaultPublicKeyCore.ExpiresAt = t
}

func (k defaultPublicKey) User(dbMap *DataMapper) User {
	if k.defaultPublicKeyCore.UserId == 0 {
		return nil
	}
	u, err := FindUserWithId(k.defaultPublicKeyCore.UserId, dbMap)
	if err != nil {
		log.Println("An error occured when trying to retreive user")
		log.Println(err)
		return nil
	}
	return u
}

func (k defaultPublicKey) Encrypt(msg string) (string, error) {
	cipher, err := gpgme.EncryptMessage(msg, k.Fingerprint())
	if err != nil {
		return "", err
	}
	return cipher, nil
}

func (k defaultPublicKey) EncryptAndSave(senderId int, t, subject string, dbMap *DataMapper) (EncryptedMessage, error) {
	cipher, err := gpgme.EncryptMessage(t, k.Fingerprint())
	if err != nil {
		return nil, err
	}

	msg, err := newMessage(k.Id(), senderId, cipher, subject)
	if err != nil {
		return nil, err
	}

	err = msg.Save(dbMap)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (k *defaultPublicKey) Activate() {
	if k.defaultPublicKeyCore.ActivatedAt.IsZero() {
		k.defaultPublicKeyCore.ActivatedAt = time.Now().UTC()
	}
}

func (k *defaultPublicKey) Messages(dbMap *DataMapper) ([]EncryptedMessage, error) {
	var ret []EncryptedMessage
	var messages []*encryptedMessageCore
	_, err := dbMap.Select(&messages, "SELECT * FROM encrypted_messages WHERE public_key_id = ?", k.Id())
	if err != nil {
		return nil, err
	}
	for _, m := range messages {
		ret = append(ret, &encryptedMessage{m})
	}
	return ret, nil
}

func (k defaultPublicKey) Save(dbMap *DataMapper) error {
	if k.Id() > 0 {
		_, err := dbMap.Update(k.defaultPublicKeyCore)
		return err
	}
	return dbMap.Insert(k.defaultPublicKeyCore)
}

func FindKeyWithId(id int, dbMap *DataMapper) (PublicKey, error) {
	kc := &defaultPublicKeyCore{Id: id}
	err := dbMap.SelectOne(kc, "SELECT * FROM public_keys WHERE id = ?", kc.Id)
	if err != nil {
		return nil, err
	}
	return &defaultPublicKey{kc}, nil
}

func FindPublicKeyWithFingerprint(fingerprint string, dbMap *DataMapper) (PublicKey, error) {
	kc := &defaultPublicKeyCore{Fingerprint: fingerprint}
	err := dbMap.SelectOne(kc, "SELECT * FROM public_keys WHERE fingerprint = ?", kc.Fingerprint)
	if err != nil {
		return nil, err
	}
	return &defaultPublicKey{kc}, nil
}

func FindOrCreatePublicKeyWithFingerprint(fingerprint string, dbMap *DataMapper) (PublicKey, error) {
	kc := &defaultPublicKeyCore{Fingerprint: fingerprint}
	err := dbMap.SelectOne(kc, "SELECT * FROM public_keys WHERE fingerprint = ?", kc.Fingerprint)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &defaultPublicKey{kc}, nil
}
