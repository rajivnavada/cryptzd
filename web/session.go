package web

import (
	"cryptz/crypto"
	"encoding/gob"
	"errors"
	"github.com/gorilla/sessions"
	"net/http"
	"time"
)

var (
	sessionStore    sessions.Store = sessions.NewFilesystemStore("", []byte("zillow-hackweek-11"))
	sessionName                    = "ZecureSessions"
	NilSessionError                = errors.New("SessionObject is nil")
)

const (
	SessionObjectKey = "sessionObject"
)

type SessionObject struct {
	UserId           string
	UserName         string
	UserEmail        string
	KeyFingerprint   string
	ActivationURL    string
	ActivationToken  []byte
	ActivationExpiry time.Time
	user             crypto.User
}

func (so *SessionObject) IsCurrentUser(userId string) bool {
	return so != nil && so.UserId == userId
}

func (so *SessionObject) IsEmpty() bool {
	return so == nil || so.ActivationExpiry.IsZero()
}

func (so *SessionObject) User() (crypto.User, error) {
	if so == nil {
		return nil, NilSessionError
	}
	if so.user != nil {
		return so.user, nil
	}
	user, err := crypto.FindUserWithEmail(so.UserEmail)
	if err != nil {
		return nil, err
	}
	so.user = user
	return so.user, err
}

func (so *SessionObject) Save(w http.ResponseWriter, r *http.Request) error {
	// Prepare the session
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		return err
	}

	// Add an expiry time
	so.ActivationExpiry = time.Now().Add(3 * 24 * time.Hour)

	// Add object to session
	session.Values[SessionObjectKey] = so

	// Save session
	return session.Save(r, w)
}

func (so *SessionObject) Destroy(w http.ResponseWriter, r *http.Request) error {
	// Prepare the session
	session, err := sessionStore.Get(r, sessionName)
	if err != nil || session.IsNew {
		return err
	}

	delete(session.Values, SessionObjectKey)
	return session.Save(r, w)
}

func CurrentSession(r *http.Request) (*SessionObject, error) {
	// Get the session from the request
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		return nil, err
	}

	// If the session is new, we don't need to do much more
	if session.IsNew {
		return &SessionObject{}, nil
	}

	saved, ok := session.Values[SessionObjectKey]
	if !ok || saved == nil {
		return &SessionObject{}, nil
	}

	so, ok := saved.(*SessionObject)
	if !ok {
		return &SessionObject{}, nil
	}
	return so, nil
}

func init() {
	gob.Register(&SessionObject{})
}
