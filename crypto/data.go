package crypto

import (
	"time"
)

type User interface {
	Id() int
	Name() string
	Email() string
	Comment() string
	ImageURL() string

	PublicKeys() []PublicKey
	ActivePublicKeys() []PublicKey
	EncryptAndSave(message, subject, sender string) ([]EncryptedMessage, error)
}

type PublicKey interface {
	Id() int
	UserId() int
	Fingerprint() string
	KeyData() string
	CreatedAt() time.Time
	UpdatedAt() time.Time
	ExpiresAt() time.Time
	ActivatedAt() time.Time
	Active() bool

	Activate() error
	User() User
	Messages() []EncryptedMessage
	Encrypt(string) (string, error)
	EncryptAndSave(message, subject, sender string) (EncryptedMessage, error)
}

type EncryptedMessage interface {
	Id() int
	PublicKeyId() int
	Subject() string
	Message() string
	CreatedAt() time.Time
	UpdatedAt() time.Time

	PublicKey() PublicKey
	Sender() User
}

type Project interface {
	Id() int
	AdminId() int
	Name() string
	Environment() string
	CreatedAt() time.Time
	UpdatedAt() time.Time

	Members() []ProjectMember
	AddMember(userId int) (ProjectMember, error)
	RemoveMember(userId int) error

	Credentials() []ProjectCredential
	AddCredential(key string) (ProjectCredential, error)
	UpdateCredential(key string) (ProjectCredential, error)
	RemoveCredential(key string) error
}

type ProjectMember interface {
	ProjectId() int
	UserId() int
	CreatedAt() time.Time
	UpdatedAt() time.Time

	Project() Project
	User() User
}

type ProjectCredential interface {
	Id() int
	ProjectId() int
	Key() string
	CreatedAt() time.Time
	UpdatedAt() time.Time
}

type EncryptedProjectCredential interface {
	CredentialId() int
	PublicKeyId() int
	Credential() string
	CreatedAt() time.Time
	UpdatedAt() time.Time
	ExpiresAt() time.Time
}
