package crypto

import (
	"errors"
	"github.com/rajivnavada/gpgme"
)

var (
	DebugMode                       = false
	SqliteFilePath                  = ""
	NotImplementedError             = errors.New("Not implemented")
	InvalidArgumentsForMessageError = errors.New("Some or all of the arguments provided to message constructor are invalid.")
	StopIterationError              = errors.New("No more items to return")
	MisconfiguredKeyError           = errors.New("email address in key does not match email address of user in database.")
)

func ImportKeyAndUser(publicKey string) (PublicKey, User, error) {
	ki, err := gpgme.ImportPublicKey(publicKey)
	if err != nil {
		return nil, nil, err
	}

	var k PublicKey
	var u User

	dbMap, err := NewDataMapper()
	if err != nil {
		return nil, nil, err
	}
	defer dbMap.Close()

	k, err = FindOrCreatePublicKeyWithFingerprint(ki.Fingerprint(), dbMap)
	if err != nil {
		return nil, nil, err
	} else {
		// Try to find the user attached to this key
		u = k.User(dbMap)
		if u == nil {
			u, err = FindOrCreateUserWithEmail(ki.Email(), dbMap)
			if err != nil {
				return nil, nil, err
			}
			u.SetName(ki.Name())
			u.SetComment(ki.Comment())
			err = u.Save(dbMap)
			if err != nil {
				return nil, nil, err
			}
		} else if u.Email() != ki.Email() {
			// If the key already belongs to a user, the email addresses must match
			return nil, nil, MisconfiguredKeyError
		}

		// Now we can update some key info
		k.SetExpiresAt(ki.ExpiresAt())
		k.SetUserId(u.Id())
		err = k.Save(dbMap)
		if err != nil {
			return nil, nil, err
		}
	}

	return k, u, nil
}

//----------------------------------------
// INIT
//----------------------------------------

func InitService(sqliteFilePath string, debugMode bool) {
	SqliteFilePath = sqliteFilePath

	dbMap, err := NewDataMapper()
	if err != nil {
		panic(err)
	}
	defer dbMap.Close()

	DebugMode = debugMode

	// In non-debug environments we'll use migrations to generate tables
	if DebugMode {
		err = dbMap.CreateTablesIfNotExists()
		if err != nil {
			panic(err)
		}
	}
}
