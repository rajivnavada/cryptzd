package crypto

import (
	"time"
)

type encryptedMessageCore struct {
	Id          int       `db:"id"`
	SenderId    int       `db:"sender_id"`
	PublicKeyId int       `db:"public_key_id"`
	Subject     string    `db:"subject"`
	Cipher      []byte    `db:"cipher"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`

	sender User `db:"-"`
}

func (e *encryptedMessageCore) loadSender(dbMap *DataMapper) error {
	u, err := FindUserWithId(e.SenderId, dbMap)
	if err != nil {
		return err
	}

	e.sender = u
	return nil
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

func (em encryptedMessage) Cipher() []byte {
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
	return em.encryptedMessageCore.sender
}

func (em *encryptedMessage) Save(dbMap *DataMapper) error {
	if em.Id() > 0 {
		_, err := dbMap.Update(em.encryptedMessageCore)
		return err
	}
	return dbMap.Insert(em.encryptedMessageCore)
}

func newMessage(publicKeyId, senderId int, cipher []byte, subject string) (*encryptedMessage, error) {
	if len(cipher) == 0 || publicKeyId == 0 || senderId == 0 {
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
