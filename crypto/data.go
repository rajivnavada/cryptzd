package crypto

import (
	"time"
)

type Identifiable interface {
	Id() int
}

type Saveable interface {
	Identifiable
	Save(dbMap *DataMapper) error
}

type User interface {
	Saveable

	Name() string
	SetName(string)

	Email() string

	Comment() string
	SetComment(string)

	ImageURL() string

	PublicKeys(dbMap *DataMapper) ([]PublicKey, error)
	ActivePublicKeys(dbMap *DataMapper) ([]PublicKey, error)
	EncryptAndSave(sender User, message, subject string, dbMap *DataMapper) (map[string]EncryptedMessage, error)
}

type PublicKey interface {
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
	EncryptAndSave(sender User, message, subject string, dbMap *DataMapper) (EncryptedMessage, error)
}

type EncryptedMessage interface {
	Saveable

	PublicKeyId() int
	Subject() string
	Cipher() []byte
	CreatedAt() time.Time
	UpdatedAt() time.Time

	Sender() User
}

type Project interface {
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
	Saveable

	ProjectId() int
	UserId() int
	AccessLevel() string
	CreatedAt() time.Time
	UpdatedAt() time.Time

	User(dbMap *DataMapper) (User, error)
}

type ProjectCredentialKey interface {
	Saveable

	ProjectId() int
	Key() string
	CreatedAt() time.Time
	UpdatedAt() time.Time

	ValueForPublicKey(publicKeyId int, dbMap *DataMapper) (ProjectCredentialValue, error)
}

type ProjectCredentialValue interface {
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
	Saveable

	UserId() int
	Key() string
	Cipher() []byte
	CreatedAt() time.Time
	UpdatedAt() time.Time
}
