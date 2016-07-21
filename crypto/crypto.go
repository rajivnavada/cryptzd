package crypto

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/rajivnavada/gpgme"
	"io"
	"log"
	"strings"
	"time"
)

var (
	SqliteFilePath                  = ""
	NotImplementedError             = errors.New("Not implemented")
	InvalidArgumentsForMessageError = errors.New("Some or all of the arguments provided to message constructor are invalid.")
	StopIterationError              = errors.New("No more items to return")
)

func ImportKeyAndUser(publicKey string) (Key, User, error) {
	// Protect access to the C function
	ki, err := gpgme.ImportPublicKey(publicKey)
	if err != nil {
		return nil, nil, err
	}

	bk := &defaultPublicKeyCore{
		Fingerprint: ki.Fingerprint(),
		ExpiresAt:   ki.ExpiresAt(),
	}
	bu := &defaultUserCore{
		Name:    ki.Name(),
		Email:   ki.Email(),
		Comment: ki.Comment(),
	}

	savedUser := &defaultUserCore{Email: bu.Email}
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
		// Save and done
		if err := bu.Save(sess); err != nil {
			return nil, nil, err
		}
	}

	u := &defaultUser{bu}

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

	return &defaultPublicKey{bk}, &defaultUser{bu}, nil
}

//----------------------------------------
// INIT
//----------------------------------------

func InitService(sqliteFilePath string) {
	SqliteFilePath = sqliteFilePath

	dbMap, err := initDb(SqliteFilePath)
	if err != nil {
		panic(err)
	}
	defer dbMap.Close()

	// Add the tables
	dbMap.AddTableWithName(defaultUserCore{}, "users").SetKeys(true, "Id")
	dbMap.AddTableWithName(defaultPublicKeyCore{}, "public_keys").SetKeys(true, "Id")
	dbMap.AddTableWithName(defaultEncryptedMessageCore{}, "encrypted_messages").SetKeys(true, "Id")

	// In non-debug environments we'll use migrations to generate tables
	if *debug {
		err = dbMap.CreateTablesIfNotExists()
		if err != nil {
			panic(err)
		}
	}
}
