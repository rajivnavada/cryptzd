package crypto

import (
	"errors"
	"gibberz/mongo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"log"
	"strings"
	"time"
)

const (
	MONGO_HOST_NAME         string = "127.0.0.1"
	MONGO_DB_NAME           string = "gibberz"
	KEY_COLLECTION_NAME     string = "keys"
	USER_COLLECTION_NAME    string = "users"
	MESSAGE_COLLECTION_NAME string = "messages"
)

var (
	NotImplementedError             = errors.New("Not implemented")
	InvalidKeyError                 = errors.New("Provided public key is invalid. Please make sure the key has not expired or revoked.")
	MissingEmailError               = errors.New("Public key must contain a valid email address.")
	FailedEncryptionError           = errors.New("Failed to encrypt message.")
	InvalidArgumentsForMessageError = errors.New("Some or all of the arguments provided to message constructor are invalid.")
)

//----------------------------------------
// PUBLIC INTERFACES
//----------------------------------------

type Key interface {
	Id() string

	Fingerprint() string

	Active() bool

	// NOTE: this is not the time of creation of the key. It refers to the time the key was added to the store
	CreatedAt() time.Time

	ActivatedAt() time.Time

	ExpiresAt() time.Time

	User() User

	// This is a transient encrypt. Nothing is saved to DB
	Encrypt(string) (io.Reader, error)

	// This will save the message to DB
	EncryptMessage(messageToEncrypt, subject string, sender User) (Message, error)

	Messages() MessageCollection
}

type User interface {
	Id() string

	Name() string

	Email() string

	Comment() string

	CreatedAt() time.Time

	UpdatedAt() time.Time

	Keys() []Key
}

type Message interface {
	Id() string

	Subject() string

	Sender() User

	Key() Key

	Text() string

	CreatedAt() time.Time
}

type MessageCollection interface {
	SetPageLength(len int)

	Next() Message
}

//----------------------------------------
// GENERIC FUNCTIONS
//----------------------------------------

type reloadable interface {
	reloadFromDataStore(mongo.Session) error
}

func find(selector reloadable) error {
	sess := mongo.NewSession(MONGO_HOST_NAME, MONGO_DB_NAME)
	defer sess.Close()
	return selector.reloadFromDataStore(sess)
}

//----------------------------------------
// A Key Implementation
//----------------------------------------

type baseKey struct {
	Id bson.ObjectId "_id"

	Fingerprint string

	IsActive bool

	CreatedAt time.Time

	ActivatedAt time.Time

	ExpiresAt time.Time

	User string
}

// NOTE: this will change the receiver. Use carfully.
func (bk *baseKey) reloadFromDataStore(sess mongo.Session) error {
	// We can select with Fingerprint or Id. One is required.
	var selector bson.M
	if bk.Fingerprint != "" {
		selector = bson.M{"fingerprint": bk.Fingerprint}
	} else {
		selector = bson.M{"_id": bk.Id}
	}
	if err := sess.Find(bk, selector, KEY_COLLECTION_NAME); err != nil {
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
		bk.Id = bson.NewObjectId()
		return sess.Save(bk, KEY_COLLECTION_NAME)
	}
	return sess.Update(bk, KEY_COLLECTION_NAME)
}

func (bk *baseKey) Encrypt(msg string) (io.Reader, error) {
	cipher, err := encryptMessage(msg, bk.Fingerprint)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(cipher), nil
}

func (bk *baseKey) EncryptMessage(s, subject string, sender User) (Message, error) {
	cipher, err := encryptMessage(s, bk.Fingerprint)
	if err != nil {
		return nil, err
	}
	m, err := newMessage(cipher, subject, sender, &key{bk})
	if err != nil {
		return nil, err
	}

	sess := mongo.NewSession(MONGO_HOST_NAME, MONGO_DB_NAME)
	defer sess.Close()

	if err := m.Save(sess); err != nil {
		return nil, err
	}
	return m, err
}

type key struct {
	*baseKey
}

func (k *key) Id() string {
	return k.baseKey.Id.Hex()
}

func (k *key) Fingerprint() string {
	return k.baseKey.Fingerprint
}

func (k *key) Active() bool {
	return k.baseKey.IsActive
}

func (k *key) CreatedAt() time.Time {
	return k.baseKey.CreatedAt
}

func (k *key) ActivatedAt() time.Time {
	return k.baseKey.ActivatedAt
}

func (k *key) ExpiresAt() time.Time {
	return k.baseKey.ExpiresAt
}

func (k *key) Messages() MessageCollection {
	return nil
}

func (k *key) User() User {
	u, err := FindUserWithId(k.baseKey.User)
	if err != nil {
		log.Println("An error occured when trying to retreive user")
		log.Println(err)
		return nil
	}
	return u
}

func FindKeyWithId(id string) (Key, error) {
	k := &key{&baseKey{Id: bson.ObjectId(id)}}
	if err := find(k); err != nil {
		return nil, err
	}
	return k, nil
}

func FindKeyWithFingerprint(fingerprint string) (Key, error) {
	k := &key{&baseKey{Fingerprint: fingerprint}}
	if err := find(k); err != nil {
		return nil, err
	}
	return k, nil
}

//----------------------------------------
// A user implementation
//----------------------------------------

type baseUser struct {
	Id bson.ObjectId "_id"

	Name string

	Email string

	Comment string

	CreatedAt time.Time

	UpdatedAt time.Time
}

// NOTE: this will change the receiver. Use carfully.
func (bu *baseUser) reloadFromDataStore(sess mongo.Session) error {
	// We can select with Id or Email. One of them is required
	var selector bson.M
	if bu.Email != "" {
		selector = bson.M{"email": bu.Email}
	} else {
		selector = bson.M{"_id": bu.Id}
	}
	if err := sess.Find(bu, selector, USER_COLLECTION_NAME); err != nil {
		return err
	}
	return nil
}

func (bu *baseUser) mergeWith(saved *baseUser) {
	bu.Id = saved.Id
}

func (bu *baseUser) ObjectId() bson.ObjectId {
	return bu.Id
}

func (bu *baseUser) Save(sess mongo.Session) error {
	if !bu.Id.Valid() {
		bu.Id = bson.NewObjectId()
		return sess.Save(bu, USER_COLLECTION_NAME)
	}
	return sess.Update(bu, USER_COLLECTION_NAME)
}

type user struct {
	*baseUser
}

func (u *user) Id() string {
	return u.baseUser.Id.Hex()
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

func (u *user) CreatedAt() time.Time {
	return u.baseUser.CreatedAt
}

func (u *user) UpdatedAt() time.Time {
	return u.baseUser.UpdatedAt
}

func (u *user) Keys() []Key {
	return nil
}

func FindUserWithId(id string) (User, error) {
	u := &user{&baseUser{Id: bson.ObjectId(id)}}
	if err := find(u); err != nil {
		return nil, err
	}
	return u, nil
}

func FindUserWithEmail(email string) (User, error) {
	u := &user{&baseUser{Email: email}}
	if err := find(u); err != nil {
		return nil, err
	}
	return u, nil
}

//----------------------------------------
// MESSAGE
//----------------------------------------

type baseMessage struct {
	Id bson.ObjectId "_id"

	Subject string

	Cipher string

	Sender string

	Key string

	CreatedAt time.Time
}

func (m *baseMessage) Save(sess mongo.Session) error {
	if !m.Id.Valid() {
		m.Id = bson.NewObjectId()
		return sess.Save(m, MESSAGE_COLLECTION_NAME)
	}
	return sess.Update(m, MESSAGE_COLLECTION_NAME)
}

func (m *baseMessage) ObjectId() bson.ObjectId {
	return m.Id
}

type message struct {
	*baseMessage
}

func (m *message) Id() string {
	return m.baseMessage.Id.Hex()
}

func (m *message) Subject() string {
	return m.baseMessage.Subject
}

func (m *message) Text() string {
	return m.baseMessage.Cipher
}

func (m *message) CreatedAt() time.Time {
	return m.baseMessage.CreatedAt
}

func (m *message) Sender() User {
	senderId := m.baseMessage.Sender
	sender := &baseUser{Id: bson.ObjectId(senderId)}

	if err := find(sender); err != nil {
		log.Println("An error occured trying to find user with id = ", senderId)
		log.Println(err)
		return nil
	}
	return &user{sender}
}

func (m *message) Key() Key {
	keyId := m.baseMessage.Key
	k := &baseKey{Id: bson.ObjectId(keyId)}

	if err := find(k); err != nil {
		log.Println("An error occured trying to find key with id = ", keyId)
		log.Println(err)
		return nil
	}
	return &key{k}
}

func newMessage(cipher, subject string, sender User, key Key) (*message, error) {
	if cipher == "" || sender.Id() == "" || key.Id() == "" {
		return nil, InvalidArgumentsForMessageError
	}
	return &message{&baseMessage{
		Subject:   subject,
		Cipher:    cipher,
		Sender:    sender.Id(),
		Key:       key.Id(),
		CreatedAt: time.Now().UTC(),
	}}, nil
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

	savedUser := &baseUser{Email: bu.Email}
	if err := savedUser.reloadFromDataStore(sess); err != nil && err != mgo.ErrNotFound {
		return nil, nil, err
	} else {
		// Update the object
		if err == nil {
			bu.mergeWith(savedUser)
		} else {
			bu.CreatedAt = time.Now().UTC()
		}
		bu.UpdatedAt = time.Now().UTC()
		//		// We need to detect if the key already exists in the user's list of keys
		//		keyExists := false
		//		if bu.Keys == nil {
		//			bu.Keys = make([]string, 0)
		//		} else {
		//			for _, fpr := range bu.Keys {
		//				if fpr == bk.Fingerprint {
		//					keyExists = true
		//					break
		//				}
		//			}
		//		}
		//		// Add the key if necessary
		//		if !keyExists {
		//			bu.Keys = append(bu.Keys, bk.Fingerprint)
		//		}
		// Save and done
		if err := bu.Save(sess); err != nil {
			return nil, nil, err
		}
	}

	u := &user{bu}

	savedKey := &baseKey{Fingerprint: bk.Fingerprint}
	if err := savedKey.reloadFromDataStore(sess); err != nil && err != mgo.ErrNotFound {
		return nil, nil, err
	} else {
		// If we successfully loaded the key from database, update some data
		if err == nil {
			bk.mergeWith(savedKey)
		} else {
			bk.CreatedAt = time.Now().UTC()
		}
		// TODO: Users should never really change once set.
		// Look into it and decide if this needs to do some error checking around that.
		bk.User = u.Id()
		// Save the key
		if err := bk.Save(sess); err != nil {
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
