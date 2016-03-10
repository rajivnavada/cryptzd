package crypto

import (
	"errors"
	"fmt"
	"gibberz/mongo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	reload(sess mongo.Session) error

	Save(sess mongo.Session) error
}

type Key interface {
	Id() string

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
	Id() string

	Name() string

	Email() string

	Comment() string
}

type SaveableUser interface {
	User

	Saveable
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
	Id bson.ObjectId "_id"

	Fingerprint string

	IsActive bool

	ActivatedAt time.Time

	ExpiresAt time.Time
}

func (bk *baseKey) reload(sess mongo.Session) error {
	saved := &baseKey{}
	if err := sess.FindDocument(saved, bson.M{"fingerprint": bk.Fingerprint}, KEY_COLLECTION_NAME); err != nil {
		return err
	}

	bk.Id = saved.Id
	bk.IsActive = saved.IsActive
	bk.ActivatedAt = saved.ActivatedAt
	bk.ExpiresAt = saved.ExpiresAt

	return nil
}

func (bk *baseKey) Save(sess mongo.Session) error {
	bk.Id = bson.NewObjectId()
	return sess.SaveDocument(bk, KEY_COLLECTION_NAME)
}

func (bk *baseKey) Encrypt(msg string) (string, error) {
	return encryptMessage(msg, bk.Fingerprint)
}

type key struct {
	*baseKey
}

func (k *key) Id() string {
	return k.baseKey.Id.String()
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
	Id bson.ObjectId "_id"

	Name string

	Email string

	Comment string
}

func (bu *baseUser) reload(sess mongo.Session) error {
	saved := baseUser{}
	if err := sess.FindDocument(&saved, bson.M{"email": bu.Email}, USER_COLLECTION_NAME); err != nil {
		return nil
	}

	bu.Id = saved.Id
	bu.Name = saved.Name
	bu.Comment = saved.Comment
	return nil
}

func (bu *baseUser) Save(sess mongo.Session) error {
	bu.Id = bson.NewObjectId()
	return sess.SaveDocument(bu, USER_COLLECTION_NAME)
}

type user struct {
	*baseUser
}

func (u *user) Id() string {
	return u.baseUser.Id.String()
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
	} else {
		if err := k.reload(sess); err != nil {
			return nil, nil, err
		}
		println("Key id is", k.Id())
	}

	u := &user{bu}
	if err := u.Save(sess); err != nil && !mgo.IsDup(err) {
		return nil, nil, err
	} else {
		if err := u.reload(sess); err != nil {
			return nil, nil, err
		}
		println("User id is", u.Id())
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
