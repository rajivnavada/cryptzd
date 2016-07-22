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

	Name() string
	Environment() string
	DefaultAccessLevel() string
	CreatedAt() time.Time
	UpdatedAt() time.Time

	Members(dbMap *DataMapper) ([]ProjectMember, error)
	AddMember(userId int, dbMap *DataMapper) (ProjectMember, error)
	RemoveMember(userId int, dbMap *DataMapper) error

	Credentials(dbMap *DataMapper) ([]ProjectCredentialKey, error)
	AddCredential(key, value string, dbMap *DataMapper) (ProjectCredentialKey, error)
	UpdateCredential(key, value string, dbMap *DataMapper) (ProjectCredentialKey, error)
	RemoveCredential(key string, dbMap *DataMapper) error
}

type ProjectMember interface {
	Identifiable
	Saveable

	ProjectId() int
	UserId() int
	AccessLevel() string
	CreatedAt() time.Time
	UpdatedAt() time.Time

	User(dbMap *DataMapper) (User, error)
}

type ProjectCredentialKey interface {
	Identifiable
	Saveable

	ProjectId() int
	Key() string
	CreatedAt() time.Time
	UpdatedAt() time.Time
}

type ProjectCredentialValue interface {
	Identifiable
	Saveable

	CredentialId() int
	MemberId() int
	PublicKeyId() int

	Cipher() []byte
	SetCipher([]byte)

	CreatedAt() time.Time
	UpdatedAt() time.Time
	ExpiresAt() time.Time
}

type UserCredential interface {
	Identifiable
	Saveable

	UserId() int
	Key() string
	Cipher() []byte
	CreatedAt() time.Time
	UpdatedAt() time.Time
}
