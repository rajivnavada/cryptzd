package crypto

import (
	"crypto/md5"
	"fmt"
	"io"
	"strings"
	"time"
)

type defaultUserCore struct {
	Id        string    `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Comment   string    `db:"comment"`
	ImageURL  string    `db:"image_url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type defaultUser struct {
	*defaultUserCore
}

func (du defaultUser) Id() string {
	return du.defaultUserCore.Id
}

func (du defaultUser) Name() string {
	return du.defaultUserCore.Name
}

func (du defaultUser) Email() string {
	return strings.TrimSpace(du.defaultUserCore.Email)
}

func (du defaultUser) Comment() string {
	return du.defaultUserCore.Comment
}

func (du defaultUser) ImageURL() string {
	email := strings.ToLower(du.Email())
	h := md5.New()
	io.WriteString(h, email)
	return fmt.Sprintf("//www.gravatar.com/avatar/%x?s=64&d=wavatar", h.Sum(nil))
}

func (du defaultUser) PublicKeys() []PublicKey {
	return make([]PublicKey, 0)
}

func (du defaultUser) ActivePublicKeys() []PublicKey {
	return make([]PublicKey, 0)
}

func (du defaultUser) EncryptAndSave(message, subject, sender string) ([]EncryptedMessage, error) {
	kc := du.ActivePublicKeys()
	ch := make(chan encryptionResult)

	// Loop over the keys and create go routines to encrypt messages per key
	for _, k := range kc {

		go func(message, subject, sender string, k PublicKey) {

			er := encryptionResult{key: k.Fingerprint()}
			encrypted, err := k.EncryptAndSave(message, subject, sender)
			if err != nil {
				er.err = err
			} else {
				er.message = encrypted
			}
			ch <- er

		}(message, subject, sender, k)

	}

	var ret encryptionResults

	for ret.Size() < len(kc) {
		select {
		case encResult := <-ch:
			ret.Add(encResult)
			break
		}
	}

	close(ch)

	if ret.IsErr() {
		return nil, ret
	}

	return ret, nil
}

func FindUserWithId(id int) (User, error) {
	u := &defaultUser{&defaultUserCore{Id: id}}
	if err := find(u); err != nil {
		return nil, err
	}
	return u, nil
}

func FindUserWithEmail(email string) (User, error) {
	u := &defaultUser{&defaultUserCore{Email: email}}
	if err := find(u); err != nil {
		return nil, err
	}
	return u, nil
}

func FindAllUsers() []User {
	return make([]User, 0)
}

////----------------------------------------
//// A user implementation
////----------------------------------------
//
//type baseUser struct {
//	Id bson.ObjectId "_id"
//
//	Name string
//
//	Email string
//
//	Comment string
//
//	IsActive bool
//
//	CreatedAt time.Time
//
//	UpdatedAt time.Time
//
//	ActivatedAt time.Time
//}
//
//// NOTE: this will change the receiver. Use carfully.
//func (bu *baseUser) reloadFromDataStore(sess Session) error {
//	// We can select with Id or Email. One of them is required
//	var selector bson.M
//	if bu.Email != "" {
//		selector = bson.M{"email": bu.Email}
//	} else {
//		selector = bson.M{"_id": bu.Id}
//	}
//	if err := sess.Find(bu, selector, USER_COLLECTION_NAME); err != nil {
//		return err
//	}
//	return nil
//}
//
//func (bu *baseUser) mergeWith(saved *baseUser) {
//	bu.Id = saved.Id
//}
//
//func (bu *baseUser) ObjectId() bson.ObjectId {
//	return bu.Id
//}
//
//func (bu *baseUser) Keys() KeyCollection {
//	return newKeyCollection(uint8(20), bu.Id.Hex(), 0)
//}
//
//func (bu *baseUser) Save(sess Session) error {
//	if !bu.Id.Valid() {
//		bu.Id = bson.NewObjectId()
//		return sess.Save(bu, USER_COLLECTION_NAME)
//	}
//	return sess.Update(bu, USER_COLLECTION_NAME)
//}
//
//func (bu *baseUser) Activate() error {
//	bu.IsActive = true
//	if bu.ActivatedAt.IsZero() {
//		bu.ActivatedAt = time.Now().UTC()
//	}
//
//	sess := newSession()
//	defer sess.Close()
//
//	return bu.Save(sess)
//}
//
//type user struct {
//	*baseUser
//}
//
//func (u *user) Id() string {
//	return u.baseUser.Id.Hex()
//}
//
//func (u *user) Name() string {
//	return u.baseUser.Name
//}
//
//func (u *user) Email() string {
//	return u.baseUser.Email
//}
//
//func (u *user) Comment() string {
//	return u.baseUser.Comment
//}
//
//func (u *user) Active() bool {
//	return u.baseUser.IsActive
//}
//
//func (u *user) ImageURL() string {
//	email := strings.ToLower(strings.TrimSpace(u.Email()))
//	h := md5.New()
//	io.WriteString(h, email)
//	return fmt.Sprintf("//www.gravatar.com/avatar/%x?s=64&d=wavatar", h.Sum(nil))
//}
//
//func (u *user) CreatedAt() time.Time {
//	return u.baseUser.CreatedAt
//}
//
//func (u *user) UpdatedAt() time.Time {
//	return u.baseUser.UpdatedAt
//}
//
//func (u *user) ActivatedAt() time.Time {
//	return u.baseUser.ActivatedAt
//}
//
//func (u *user) EncryptMessage(message, subject, sender string) (map[string]Message, error) {
//	// First get the keys for this user
//	kc := u.Keys()
//
//	okCh := make(chan encryptionResult)
//	errCh := make(chan error)
//	total := 0
//
//	// Loop over the keys and create go routines to encrypt messages per key
//	for k, err := kc.Next(); k != nil && err == nil; k, err = kc.Next() {
//
//		total++
//
//		go func(message, subject, sender string, k Key) {
//
//			encrypted, err := k.EncryptMessage(message, subject, sender)
//			if err != nil {
//				errCh <- err
//				return
//			}
//			okCh <- encryptionResult{
//				key:     k.Fingerprint(),
//				message: encrypted,
//			}
//
//		}(message, subject, sender, k)
//
//	}
//
//	numOk := 0
//	numErr := 0
//	ret := make(map[string]Message)
//	errors := make([]error, 0)
//
//	for numOk+numErr < total {
//		select {
//		case encResult := <-okCh:
//			numOk++
//			ret[encResult.key] = encResult.message
//			break
//
//		case err := <-errCh:
//			numErr++
//			log.Println("An error occured when trying to encrypt message")
//			log.Println(err)
//			errors = append(errors, err)
//			break
//		}
//	}
//
//	// TODO: Return an `EncryptionResult` object that can contain errors and result data
//	//       This way the caller can decide what is to be treated as an error.
//	if numErr > 0 {
//		return nil, errors[0]
//	}
//	return ret, nil
//}
