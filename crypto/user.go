package crypto

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"
)

type userCore struct {
	Id        int       `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Comment   string    `db:"comment"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type user struct {
	*userCore
}

func (u user) Id() int {
	return u.userCore.Id
}

func (u user) Name() string {
	return u.userCore.Name
}

func (u *user) SetName(name string) {
	u.userCore.Name = name
}

func (u user) Email() string {
	return strings.TrimSpace(u.userCore.Email)
}

func (u user) Comment() string {
	return u.userCore.Comment
}

func (u *user) SetComment(comment string) {
	u.userCore.Comment = comment
}

func (u user) ImageURL() string {
	email := strings.ToLower(u.Email())
	h := md5.New()
	io.WriteString(h, email)
	return fmt.Sprintf("//www.gravatar.com/avatar/%x?s=64&d=wavatar", h.Sum(nil))
}

func (u user) PublicKeys(dbMap *DataMapper) ([]PublicKey, error) {
	return make([]PublicKey, 0), nil
}

func (u user) ActivePublicKeys(dbMap *DataMapper) ([]PublicKey, error) {
	var ret []PublicKey
	var keys []*publicKeyCore
	_, err := dbMap.Select(&keys, "SELECT * FROM public_keys WHERE user_id = ?", u.Id())
	if err != nil {
		return nil, err
	}
	for _, k := range keys {
		ret = append(ret, &publicKey{k})
	}
	return ret, nil
}

func (u user) EncryptAndSave(senderId int, message, subject string, dbMap *DataMapper) (map[string]EncryptedMessage, error) {
	kc, err := u.ActivePublicKeys(dbMap)
	if err != nil {
		return nil, err
	}
	ch := make(chan encryptionResult)

	// Loop over the keys and create go routines to encrypt messages per key
	for _, k := range kc {

		go func(senderId int, message, subject string, dbMap *DataMapper, k PublicKey) {

			er := encryptionResult{key: k.Fingerprint()}
			encrypted, err := k.EncryptAndSave(senderId, message, subject, dbMap)
			if err != nil {
				er.err = err
			} else {
				er.message = encrypted
			}
			ch <- er

		}(senderId, message, subject, dbMap, k)

	}

	var results encryptionResults
	ret := make(map[string]EncryptedMessage)

	for results.Size() < len(kc) {
		select {
		case encResult := <-ch:
			results.Add(encResult)
			if encResult.err == nil {
				ret[encResult.Key()] = encResult.Message()
			}
			break
		}
	}

	close(ch)

	if results.IsErr() {
		return nil, results
	}
	return ret, nil
}

func (u user) Save(dbMap *DataMapper) error {
	if u.Id() > 0 {
		_, err := dbMap.Update(u.userCore)
		return err
	}
	return dbMap.Insert(u.userCore)
}

func FindUserWithId(id int, dbMap *DataMapper) (User, error) {
	uc := &userCore{Id: id}
	err := dbMap.SelectOne(uc, "SELECT * FROM users WHERE id = ?", uc.Id)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &user{uc}, nil
}

func FindUserWithEmail(email string, dbMap *DataMapper) (User, error) {
	uc := &userCore{Email: email}
	err := dbMap.SelectOne(uc, "SELECT * FROM users WHERE email = ?", uc.Email)
	if err != nil {
		return nil, err
	}
	return &user{uc}, nil
}

func FindOrCreateUserWithEmail(email string, dbMap *DataMapper) (User, error) {
	uc := &userCore{Email: email}
	err := dbMap.SelectOne(uc, "SELECT * FROM users WHERE email = ?", uc.Email)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &user{uc}, nil
}

func FindAllUsers(dbMap *DataMapper) ([]User, error) {
	var users []*userCore
	_, err := dbMap.Select(&users, "SELECT * FROM users ORDER BY id ASC")
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	var ret []User
	for _, u := range users {
		ret = append(ret, &user{u})
	}
	return ret, nil
}
