package crypto

import (
	"time"
)

type Identifiable interface {
	Id() int
}

type Saveable interface {
	Save(dbMap *DataMapper) error
}

type User interface {
	Identifiable
	Saveable

	Name() string
	SetName(string)

	Email() string

	Comment() string
	SetComment(string)

	ImageURL() string

	PublicKeys(dbMap *DataMapper) ([]PublicKey, error)
	ActivePublicKeys(dbMap *DataMapper) ([]PublicKey, error)
	EncryptAndSave(senderId int, message, subject string, dbMap *DataMapper) (map[string]EncryptedMessage, error)
}

type PublicKey interface {
	Identifiable
	Saveable

	UserId() int
	SetUserId(int)

	Fingerprint() string

	KeyData() []byte
	SetKeyData([]byte)

	CreatedAt() time.Time
	UpdatedAt() time.Time

	ExpiresAt() time.Time
	SetExpiresAt(time.Time)

	ActivatedAt() time.Time
	Active() bool

	Activate()
	User(dbMap *DataMapper) User
	Messages(dbMap *DataMapper) ([]EncryptedMessage, error)
	Encrypt(string) (string, error)
	EncryptAndSave(senderId int, message, subject string, dbMap *DataMapper) (EncryptedMessage, error)
}

type EncryptedMessage interface {
	Identifiable
	Saveable

	PublicKeyId() int
	Subject() string
	Cipher() []byte
	CreatedAt() time.Time
	UpdatedAt() time.Time

	Sender() User
}

type Project interface {
	Identifiable
	Saveable

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
	Identifiable
	Saveable

	ProjectId() int
	UserId() int
	CreatedAt() time.Time
	UpdatedAt() time.Time

	Project() Project
	User() User
}

type ProjectCredential interface {
	Identifiable
	Saveable

	ProjectId() int
	Key() string
	CreatedAt() time.Time
	UpdatedAt() time.Time
}

type EncryptedProjectCredential interface {
	Identifiable
	Saveable

	CredentialId() int
	PublicKeyId() int
	Credential() string
	CreatedAt() time.Time
	UpdatedAt() time.Time
	ExpiresAt() time.Time
}

type UserCredential interface {
	Identifiable
	Saveable

	UserId() int
	Key() string

	CreatedAt() time.Time
	UpdatedAt() time.Time
}
