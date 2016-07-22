package crypto

import (
	"database/sql"
	"github.com/rajivnavada/gpgme"
	"log"
	"time"
)

type publicKeyCore struct {
	Id          int       `db:"id"`
	UserId      int       `db:"user_id"`
	Fingerprint string    `db:"fingerprint"`
	KeyData     []byte    `db:"key_data"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	ActivatedAt time.Time `db:"activated_at"`
	ExpiresAt   time.Time `db:"expires_at"`
}

type publicKey struct {
	*publicKeyCore
}

func (k publicKey) Id() int {
	return k.publicKeyCore.Id
}

func (k publicKey) UserId() int {
	return k.publicKeyCore.UserId
}

func (k *publicKey) SetUserId(uid int) {
	k.publicKeyCore.UserId = uid
}

func (k publicKey) Fingerprint() string {
	return k.publicKeyCore.Fingerprint
}

func (k publicKey) KeyData() []byte {
	return k.publicKeyCore.KeyData
}

func (k *publicKey) SetKeyData(d []byte) {
	k.publicKeyCore.KeyData = d
}

func (k publicKey) Active() bool {
	return !k.publicKeyCore.ActivatedAt.IsZero()
}

func (k publicKey) CreatedAt() time.Time {
	return k.publicKeyCore.CreatedAt
}

func (k publicKey) UpdatedAt() time.Time {
	return k.publicKeyCore.UpdatedAt
}

func (k publicKey) ActivatedAt() time.Time {
	return k.publicKeyCore.ActivatedAt
}

func (k publicKey) ExpiresAt() time.Time {
	return k.publicKeyCore.ExpiresAt
}

func (k *publicKey) SetExpiresAt(t time.Time) {
	k.publicKeyCore.ExpiresAt = t
}

func (k publicKey) User(dbMap *DataMapper) User {
	if k.publicKeyCore.UserId == 0 {
		return nil
	}
	u, err := FindUserWithId(k.publicKeyCore.UserId, dbMap)
	if err != nil {
		log.Println("An error occured when trying to retreive user")
		log.Println(err)
		return nil
	}
	return u
}

func (k publicKey) Encrypt(msg string) (string, error) {
	cipher, err := gpgme.EncryptMessage(msg, k.Fingerprint())
	if err != nil {
		return "", err
	}
	return cipher, nil
}

func (k publicKey) EncryptAndSave(senderId int, t, subject string, dbMap *DataMapper) (EncryptedMessage, error) {
	cipher, err := gpgme.EncryptMessage(t, k.Fingerprint())
	if err != nil {
		return nil, err
	}

	msg, err := newMessage(k.Id(), senderId, []byte(cipher), subject)
	if err != nil {
		return nil, err
	}

	err = msg.Save(dbMap)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (k *publicKey) Activate() {
	if k.publicKeyCore.ActivatedAt.IsZero() {
		k.publicKeyCore.ActivatedAt = time.Now().UTC()
	}
}

func (k *publicKey) Messages(dbMap *DataMapper) ([]EncryptedMessage, error) {
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

func (k publicKey) Save(dbMap *DataMapper) error {
	if k.Id() > 0 {
		_, err := dbMap.Update(k.publicKeyCore)
		return err
	}
	return dbMap.Insert(k.publicKeyCore)
}

func FindKeyWithId(id int, dbMap *DataMapper) (PublicKey, error) {
	kc := &publicKeyCore{Id: id}
	err := dbMap.SelectOne(kc, "SELECT * FROM public_keys WHERE id = ?", kc.Id)
	if err != nil {
		return nil, err
	}
	return &publicKey{kc}, nil
}

func FindPublicKeyWithFingerprint(fingerprint string, dbMap *DataMapper) (PublicKey, error) {
	kc := &publicKeyCore{Fingerprint: fingerprint}
	err := dbMap.SelectOne(kc, "SELECT * FROM public_keys WHERE fingerprint = ?", kc.Fingerprint)
	if err != nil {
		return nil, err
	}
	return &publicKey{kc}, nil
}

func FindOrCreatePublicKeyWithFingerprint(fingerprint string, dbMap *DataMapper) (PublicKey, error) {
	kc := &publicKeyCore{Fingerprint: fingerprint}
	err := dbMap.SelectOne(kc, "SELECT * FROM public_keys WHERE fingerprint = ?", kc.Fingerprint)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &publicKey{kc}, nil
}
