package crypto

import (
	"errors"
	"gibberz/mongo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"sort"
	"strings"
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

type Looper interface {
	Next() error
}

type Key interface {
	Id() string

	Fingerprint() string

	Active() bool

	ActivatedAt() time.Time

	ExpiresAt() time.Time

	Encrypt(string) (EncryptedData, error)

	Messages() EncryptedDataCollection
}

type User interface {
	Id() string

	Name() string

	Email() string

	Comment() string

	Keys() []Key
}

type EncryptedData interface {
	io.Reader
}

type EncryptedDataCollection interface {
	Looper

	Data() EncryptedData

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
	if err := sess.FindDocument(bk, bson.M{"fingerprint": bk.Fingerprint}, KEY_COLLECTION_NAME); err != nil {
		return err
	}
	return nil
}

func (bk *baseKey) mergeWith(saved *baseKey) {
	bk.Id = saved.Id
	bk.IsActive = saved.IsActive
	bk.ActivatedAt = saved.ActivatedAt
}

func (bk *baseKey) ObjectId() bson.ObjectId {
	return bk.Id
}

func (bk *baseKey) Save(sess mongo.Session) error {
	if !bk.Id.Valid() {
		println(bk.Id)
		println("Not valid :(")
		bk.Id = bson.NewObjectId()
		return sess.SaveDocument(bk, KEY_COLLECTION_NAME)
	}
	return sess.UpdateDocument(bk, KEY_COLLECTION_NAME)
}

func (bk *baseKey) Encrypt(msg string) (EncryptedData, error) {
	s, err := encryptMessage(msg, bk.Fingerprint)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(s), nil
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

func (k *key) Messages() EncryptedDataCollection {
	return nil
}

//----------------------------------------
// A user implementation
//----------------------------------------

type baseUser struct {
	Id bson.ObjectId "_id"

	Name string

	Email string

	Comment string

	Keys []string
}

func (bu *baseUser) reload(sess mongo.Session) error {
	if err := sess.FindDocument(bu, bson.M{"email": bu.Email}, USER_COLLECTION_NAME); err != nil {
		return err
	}
	return nil
}

func (bu *baseUser) mergeWith(saved *baseUser) {
	bu.Id = saved.Id
	bu.Keys = saved.Keys
}

func (bu *baseUser) ObjectId() bson.ObjectId {
	return bu.Id
}

func (bu *baseUser) Save(sess mongo.Session) error {
	if !bu.Id.Valid() {
		bu.Id = bson.NewObjectId()
		return sess.SaveDocument(bu, USER_COLLECTION_NAME)
	}
	return sess.UpdateDocument(bu, USER_COLLECTION_NAME)
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

func (u *user) Keys() []Key {
	return nil
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
	} else if bk == nil || bu == nil {
		return nil, nil, errors.New("An unknown error has occured in ImportKeyAndUser()")
	}

	sess := mongo.NewSession(MONGO_HOST_NAME, MONGO_DB_NAME)
	defer sess.Close()

	savedKey := &baseKey{Fingerprint: bk.Fingerprint}
	if err := savedKey.reload(sess); err != nil && err != mgo.ErrNotFound {
		return nil, nil, err
	} else {
		// If we successfully loaded the key from database, update some data
		if err == nil {
			bk.mergeWith(savedKey)
		}
		if err := bk.Save(sess); err != nil {
			return nil, nil, err
		}
	}

	savedUser := &baseUser{Email: bu.Email}
	if err := savedUser.reload(sess); err != nil && err != mgo.ErrNotFound {
		return nil, nil, err
	} else {
		// Update the object
		if err == nil {
			bu.mergeWith(savedUser)
		}
		// We need to detect if the key already exists in the user's list of keys
		keyExists := false
		if bu.Keys == nil {
			bu.Keys = make([]string, 0)
		} else {
			for _, fpr := range bu.Keys {
				if fpr == bk.Fingerprint {
					break
				}
			}
		}
		// Add the key if necessary
		if !keyExists {
			bu.Keys = append(bu.Keys, bk.Fingerprint)
		}
		// Save and done
		if err := bu.Save(sess); err != nil {
			return nil, nil, err
		}
	}

	return &key{bk}, &user{bu}, nil
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
	if err := indexer.AddUniqueIndex(KEY_COLLECTION_NAME, "fingerprint"); err != nil {
		panic(err)
	}
}
