package crypto

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"
)

type defaultUserCore struct {
	Id        int       `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Comment   string    `db:"comment"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type defaultUser struct {
	*defaultUserCore
}

func (du defaultUser) Id() int {
	return du.defaultUserCore.Id
}

func (du defaultUser) Name() string {
	return du.defaultUserCore.Name
}

func (du *defaultUser) SetName(name string) {
	du.defaultUserCore.Name = name
}

func (du defaultUser) Email() string {
	return strings.TrimSpace(du.defaultUserCore.Email)
}

func (du defaultUser) Comment() string {
	return du.defaultUserCore.Comment
}

func (du *defaultUser) SetComment(comment string) {
	du.defaultUserCore.Comment = comment
}

func (du defaultUser) ImageURL() string {
	email := strings.ToLower(du.Email())
	h := md5.New()
	io.WriteString(h, email)
	return fmt.Sprintf("//www.gravatar.com/avatar/%x?s=64&d=wavatar", h.Sum(nil))
}

func (du defaultUser) PublicKeys(dbMap *DataMapper) ([]PublicKey, error) {
	return make([]PublicKey, 0), nil
}

func (du defaultUser) ActivePublicKeys(dbMap *DataMapper) ([]PublicKey, error) {
	var ret []PublicKey
	var keys []*defaultPublicKeyCore
	_, err := dbMap.Select(&keys, "SELECT * FROM public_keys WHERE user_id = ?", du.Id())
	if err != nil {
		return nil, err
	}
	for _, k := range keys {
		ret = append(ret, &defaultPublicKey{k})
	}
	return ret, nil
}

func (du defaultUser) EncryptAndSave(senderId int, message, subject string, dbMap *DataMapper) (map[string]EncryptedMessage, error) {
	kc, err := du.ActivePublicKeys(dbMap)
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

func (du defaultUser) Save(dbMap *DataMapper) error {
	if du.Id() > 0 {
		_, err := dbMap.Update(du.defaultUserCore)
		return err
	}
	return dbMap.Insert(du.defaultUserCore)
}

func FindUserWithId(id int, dbMap *DataMapper) (User, error) {
	duc := &defaultUserCore{Id: id}
	err := dbMap.SelectOne(duc, "SELECT * FROM users WHERE id = ?", duc.Id)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &defaultUser{duc}, nil
}

func FindUserWithEmail(email string, dbMap *DataMapper) (User, error) {
	duc := &defaultUserCore{Email: email}
	err := dbMap.SelectOne(duc, "SELECT * FROM users WHERE email = ?", duc.Email)
	if err != nil {
		return nil, err
	}
	return &defaultUser{duc}, nil
}

func FindOrCreateUserWithEmail(email string, dbMap *DataMapper) (User, error) {
	duc := &defaultUserCore{Email: email}
	err := dbMap.SelectOne(duc, "SELECT * FROM users WHERE email = ?", duc.Email)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &defaultUser{duc}, nil
}

func FindAllUsers(dbMap *DataMapper) ([]User, error) {
	var users []*defaultUserCore
	_, err := dbMap.Select(&users, "SELECT * FROM users ORDER BY id ASC")
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	var ret []User
	for _, u := range users {
		ret = append(ret, &defaultUser{u})
	}
	return ret, nil
}
