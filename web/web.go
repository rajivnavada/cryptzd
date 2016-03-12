package web

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"
	"zecure/crypto"
	//"github.com/gorilla/websocket"
)

const (
	tokenLength = 16

	IndexURL             = "/"
	LoginURL             = "/login"
	PendingActivationURL = "/pendingactivation"
	ActivateURLBase      = "/activate/"

	PublicKeyFormFieldName = "public_key"
	UserIdFormFieldName    = "user_id"
	SubjectFormFieldName   = "subject"
	MessageFormFieldName   = "message"
)

var (
	MissingUserIdError  = errors.New("POST data does not contain a valid userId field")
	MissingMessageError = errors.New("POST data does not contain a message")
)

func GetLoginHandler(w http.ResponseWriter, r *http.Request) {
	// If user is already logged in, redirect them to index page
	if sess, err := CurrentSession(r); err == nil && !sess.IsEmpty() {
		http.Redirect(w, r, IndexURL, http.StatusSeeOther)
		return
	}
	// Else Render template
	templateDefs := struct {
		LoginURL               string
		PublicKeyFormFieldName string
	}{
		LoginURL:               LoginURL,
		PublicKeyFormFieldName: PublicKeyFormFieldName,
	}
	if err := loginTemplate.Execute(w, &templateDefs); err != nil {
		panic(err)
	}
}

func PostLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Handle the actual logging in
	// Get the public key information and process
	publicKey := r.FormValue(PublicKeyFormFieldName)
	if publicKey == "" {
		http.Redirect(w, r, fmt.Sprintf("%s?error=emptybody", LoginURL), http.StatusSeeOther)
		return
	}

	// Try to parse
	key, user, err := crypto.ImportKeyAndUser(publicKey)
	if err != nil {
		logError(err, fmt.Sprintf("Error handling %s", r.URL.String()))
		http.Redirect(w, r, fmt.Sprintf("%s?error=invalidpublickey", LoginURL), http.StatusSeeOther)
		return
	}

	// We have user and key
	// Create an activation token and build an activation url.
	tokenBytes := make([]byte, tokenLength)
	if numBytes, err := rand.Read(tokenBytes); err != nil || numBytes != tokenLength {
		logError(err, "Error generating random bytes for activation token")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	token := hex.EncodeToString(tokenBytes)

	// Encrypt activation URL
	activationURL := buildUrl(r, ActivateURLBase+token, "")

	so := &SessionObject{
		UserName:        user.Name(),
		UserEmail:       user.Email(),
		KeyFingerprint:  key.Fingerprint(),
		ActivationToken: tokenBytes,
		ActivationURL:   activationURL,
	}

	// Save session
	err = so.Save(w, r)
	if err != nil {
		logError(err, "Error saving session")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	activationEmailWriter := &bytes.Buffer{}
	activationEmailTemplate.Execute(activationEmailWriter, so)

	activationMessage, err := key.Encrypt(activationEmailWriter.String())
	if !assertErrorIsNil(w, err, "Error encrypting message") {
		return
	}

	// TODO: send email
	logIt(activationMessage)

	// Redirect to need activation message page
	http.Redirect(w, r, buildUrl(r, PendingActivationURL, ""), http.StatusSeeOther)
}

func NeedActivationMessageHandler(w http.ResponseWriter, r *http.Request) {
	// If the user is not authenticated return
	sess := mustBeAuthenticated(w, r)
	if sess == nil {
		return
	}

	// Get the session values
	userEmail := sess.UserEmail
	if userEmail == "" {
		logIt("ERROR: Could not get email from session")
	}

	keyFingerprint := sess.KeyFingerprint
	if keyFingerprint == "" {
		logIt("ERROR: Could not get fingerprint from session")
	}

	// Prepare the template definitions
	templateDefs := &struct {
		UserEmail      string
		KeyFingerprint string
	}{
		UserEmail:      userEmail,
		KeyFingerprint: keyFingerprint,
	}

	// Show informational message asking the user to check their email
	if err := activationTemplate.Execute(w, templateDefs); err != nil {
		panic(err)
	}
}

func ActivationHandler(w http.ResponseWriter, r *http.Request) {
	// Check that the token matches and redirects appropriately
	sess := mustBeAuthenticated(w, r)
	if sess == nil {
		return
	}

	// Extract token
	vars := mux.Vars(r)
	tokenStr := vars["token"]

	// Check activation token
	var token []byte
	var err error

	if token, err = hex.DecodeString(tokenStr); err != nil || len(token) != tokenLength {
		logError(err, "Error converting token to byte array", "Token was: "+tokenStr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !bytes.Equal(sess.ActivationToken, token) {
		logError(err, "Error comparing tokens", fmt.Sprintf("%s != %s", hex.EncodeToString(sess.ActivationToken), token))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	key, err := crypto.FindKeyWithFingerprint(sess.KeyFingerprint)
	if !assertErrorIsNil(w, err, "Error finding key with fingerprint"+sess.KeyFingerprint) {
		return
	}

	err = key.Activate()
	if !assertErrorIsNil(w, err, "Error activating key") {
		return
	}

	http.Redirect(w, r, IndexURL, http.StatusSeeOther)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Logout handler.
	sess, err := CurrentSession(r)
	if !assertErrorIsNil(w, err, "Error getting current session") {
		return
	}

	// Destroy the session
	err = sess.Destroy(w, r)
	if !assertErrorIsNil(w, err, "Error destroying session") {
		return
	}

	http.Redirect(w, r, LoginURL, http.StatusSeeOther)
}

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrades the connection to a websocket connection and registers the user in a users map
	// For new connections, it should send the initial list of messages
	// For existing connections, it should push new messages as they arrive for the user
}

func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is logged in or not. If not logged in redirect to login page
	sess := mustBeAuthenticated(w, r)
	if sess == nil {
		return
	}

	// This is the landing page after login. It should send the initial set of messages
	key, err := crypto.FindKeyWithFingerprint(sess.KeyFingerprint)
	if !assertErrorIsNil(w, err, "Error finding key with fingerprint"+sess.KeyFingerprint) {
		return
	}

	// Get the message collection and use it to render template of user messages
	kc, err := key.Messages().Slice()
	if !assertErrorIsNil(w, err, "Error extracting messages for a key") {
		return
	}

	uc, err := crypto.FindAllUsers().Slice()
	if !assertErrorIsNil(w, err, "Error extracting all users") {
		return
	}

	// Execute the template and return
	messagesTemplate.Execute(w, &struct {
		Messages []crypto.Message
		Users    []crypto.User
	}{
		Messages: kc,
		Users:    uc,
	})

}

func PostMessageHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is logged in or not. If not logged in redirect to login page
	sess := mustBeAuthenticated(w, r)
	if sess == nil {
		return
	}

	errs := make([]errors, 0)

	// Check userId
	userId := r.FormValue(UserIdFormFieldName)
	if userId == "" {
		logError(err, "No userId in request")
		errs = append(errs, MissingUserIdError)
	}

	// Check message
	message := r.FormValue(MessageFormFieldName)
	if message == "" {
		logError(err, "No message in request")
		errs = append(errs, MissingMessageError)
	}

	// Subject can be empty
	subject := r.FormValue(SubjectFormFieldName)

	toUser, err := FindUserWithId(userId)
	if err != nil {
		logError(err, "Could not find user with Id "+userId)
		errs = append(errs, err)
	}

	err = toUser.EncryptMessage(message, subject, toUser.Id())
	if err != nil {
		logError(err, "Error occured when encrypting message for user")
		errs = append(errs, err)
	}

	// If len(errors) == 0, it would mean things worked successfully
	err = json.NewEncoder(w).Encode(&struct {
		Errors []error `json:"errors"`
	}{
		Errors: errs,
	})

	if err != nil {
		logError(err, "An error occured when writing JSON to response writer")
	}
}

func Router() http.Handler {
	r := mux.NewRouter()

	// Add routes
	r.HandleFunc(IndexURL, GetMessagesHandler).Methods("GET")
	r.HandleFunc(IndexURL, PostMessageHandler).Methods("POST")
	r.HandleFunc(LoginURL, GetLoginHandler).Methods("GET")
	r.HandleFunc(LoginURL, PostLoginHandler).Methods("POST")
	r.HandleFunc(PendingActivationURL, NeedActivationMessageHandler).Methods("GET")
	r.HandleFunc("/activate/{token}", ActivationHandler).Methods("GET")
	r.HandleFunc("/logout", LogoutHandler).Methods("GET")
	r.HandleFunc("/wc", WebsocketHandler)

	return r
}

func init() {
	gob.Register(time.Now())
}
