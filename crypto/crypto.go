package crypto

import (
	"errors"
	"fmt"
	"gibberz/mongo"
	"gopkg.in/mgo.v2"
	"io"
	"sort"
	"time"
)

const (
	MONGO_HOST_NAME      string = "127.0.0.1"
	MONGO_DB_NAME        string = "gibberz"
	KEY_COLLECTION_NAME  string = "keys"
	USER_COLLECTION_NAME string = "users"
)

var (
	NotImplementedError   = errors.New("Not implemented")
	InvalidKeyError       = errors.New("Provided public key is invalid. Please make sure the key has not expired or revoked.")
	MissingEmailError     = errors.New("Public key must contain a valid email address.")
	FailedEncryptionError = errors.New("Failed to encrypt message.")
)

//----------------------------------------
// PUBLIC INTERFACES
//----------------------------------------

type Saveable interface {
	Save(sess mongo.Session) error
}

type Key interface {
	Fingerprint() string

	Active() bool

	ActivatedAt() time.Time

	ExpiresAt() time.Time

	Encrypt(string) (string, error)
}

type SaveableKey interface {
	Key

	Saveable
}

type User interface {
	Name() string

	Email() string

	Comment() string
}

type SaveableUser interface {
	User

	Saveable

	AddKey(key Key)
}

type UserKeyCollection interface {
	User() User

	Next() Key

	sort.Interface
}

type EncryptedData interface {
	Data() []byte

	Key() Key

	io.Reader
}

type UserEncryptedDataCollection interface {
	User() User

	Next() EncryptedData

	sort.Interface
}

//----------------------------------------
// A Key Implementation
//----------------------------------------

type baseKey struct {
	Fingerprint string

	IsActive bool

	ActivatedAt time.Time

	ExpiresAt time.Time
}

func (bk *baseKey) Save(sess mongo.Session) error {
	return sess.SaveDocument(bk, KEY_COLLECTION_NAME)
}

func (bk *baseKey) Encrypt(msg string) (string, error) {
	return encryptMessage(msg, bk.Fingerprint)
}

type key struct {
	*baseKey
}

func (k *key) Fingerprint() string {
	return k.baseKey.Fingerprint
}

func (k *key) Active() bool {
	return k.baseKey.IsActive
}

func (k *key) ActivatedAt() time.Time {
	return k.baseKey.ActivatedAt
}

func (k *key) ExpiresAt() time.Time {
	return k.baseKey.ExpiresAt
}

//----------------------------------------
// A user implementation
//----------------------------------------

type baseUser struct {
	Name string

	Email string

	Comment string
}

func (bu *baseUser) Save(sess mongo.Session) error {
	return sess.SaveDocument(bu, USER_COLLECTION_NAME)
}

type user struct {
	*baseUser
}

func (u *user) Name() string {
	return u.baseUser.Name
}

func (u *user) Email() string {
	return u.baseUser.Email
}

func (u *user) Comment() string {
	return u.baseUser.Comment
}

//----------------------------------------
// PUBLIC FUNCTIONS
//----------------------------------------

func ImportKeyAndUser(publicKey string) (Key, User, error) {
	// This is where we need to do a C thang
	bk := &baseKey{}
	bu := &baseUser{}

	if err := importPublicKey(publicKey, bk, bu); err != nil {
		return nil, nil, err
	}

	sess := mongo.NewSession(MONGO_HOST_NAME, MONGO_DB_NAME)
	defer sess.Close()

	k := &key{bk}
	if err := k.Save(sess); err != nil && !mgo.IsDup(err) {
		return nil, nil, err
	}

	u := &user{bu}
	if err := u.Save(sess); err != nil && !mgo.IsDup(err) {
		return nil, nil, err
	}

	return k, u, nil
}

//----------------------------------------
// INIT
//----------------------------------------

func init() {
	// Make sure we have the right indexes on the data
	indexer := mongo.NewIndexer(MONGO_HOST_NAME, MONGO_DB_NAME)
	defer indexer.Close()

	if err := indexer.AddUniqueIndex(USER_COLLECTION_NAME, "email"); err != nil {
		panic(err)
	}
	fmt.Println("Unique index applied to", USER_COLLECTION_NAME)

	if err := indexer.AddUniqueIndex(KEY_COLLECTION_NAME, "fingerprint"); err != nil {
		panic(err)
	}
	fmt.Println("Unique index applied to", KEY_COLLECTION_NAME)
}
