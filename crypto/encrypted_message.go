package crypto

import (
	"time"
)

type encryptedMessageCore struct {
	Id          int       `db:"id"`
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

func (em encryptedMessage) Sender() User {
	// TODO
	return nil
}

func (em *message) PublicKey() PublicKey {
	// TODO
	return nil
}

func newMessage(publicKeyId int, cipher, subject, sender string) (*encryptedMessage, error) {
	if publicKeyId == 0 || cipher == "" || sender == "" || key == "" {
		return nil, InvalidArgumentsForMessageError
	}
	currentTime := time.Now().UTC()
	return &encryptedMessage{&encryptedMessageCore{
		PublicKeyId: publicKeyId,
		Subject:     subject,
		Cipher:      cipher,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}}, nil
}
