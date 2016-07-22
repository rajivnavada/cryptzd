package crypto

import (
	"time"
)

type encryptedMessageCore struct {
	Id          int       `db:"id"`
	SenderId    int       `db:"sender_id"`
	PublicKeyId int       `db:"public_key_id"`
	Subject     string    `db:"subject"`
	Cipher      string    `db:"cipher"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type encryptedMessage struct {
	*encryptedMessageCore
}

func (em encryptedMessage) Id() int {
	return em.encryptedMessageCore.Id
}

func (em encryptedMessage) PublicKeyId() int {
	return em.encryptedMessageCore.PublicKeyId
}

func (em encryptedMessage) Subject() string {
	return em.encryptedMessageCore.Subject
}

func (em encryptedMessage) Cipher() string {
	return em.encryptedMessageCore.Cipher
}

func (em encryptedMessage) CreatedAt() time.Time {
	return em.encryptedMessageCore.CreatedAt
}

func (em encryptedMessage) UpdatedAt() time.Time {
	return em.encryptedMessageCore.UpdatedAt
}

func (em *encryptedMessage) Sender() User {
	if em.encryptedMessageCore.SenderId == 0 {
		return nil
	}
	dbMap, err := NewDataMapper()
	if err != nil {
		return nil
	}
	defer dbMap.Close()

	u, err := FindUserWithId(em.encryptedMessageCore.SenderId, dbMap)
	if err != nil {
		return nil
	}

	return u
}

func (em *encryptedMessage) Save(dbMap *DataMapper) error {
	if em.Id() > 0 {
		_, err := dbMap.Update(em.encryptedMessageCore)
		return err
	}
	return dbMap.Insert(em.encryptedMessageCore)
}

func newMessage(publicKeyId, senderId int, cipher, subject string) (*encryptedMessage, error) {
	if publicKeyId == 0 || senderId == 0 || cipher == "" {
		return nil, InvalidArgumentsForMessageError
	}
	currentTime := time.Now().UTC()
	return &encryptedMessage{&encryptedMessageCore{
		PublicKeyId: publicKeyId,
		SenderId:    senderId,
		Subject:     subject,
		Cipher:      cipher,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}}, nil
}
