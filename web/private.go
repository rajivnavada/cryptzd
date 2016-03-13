package web

import (
	"log"
	"net/http"
	"net/url"
)

type templateArgs struct {
	Title      string
	ShowHeader bool
	Extensions interface{}
}

func newTemplateArgs() *templateArgs {
	return &templateArgs{
		Title:      "CRYPTZ | A messaging platform to securely communicate with peers",
		ShowHeader: true,
	}
}

func logError(err error, otherInfo ...string) {
	log.Println("")
	log.Println("----------------------------------------")
	for _, s := range otherInfo {
		log.Println(s)
	}
	log.Println(err)
	log.Println("----------------------------------------")
	log.Println("")
}

func logIt(messages ...string) {
	log.Println("")
	log.Println("----------------------------------------")
	for _, s := range messages {
		log.Println(s)
	}
	log.Println("----------------------------------------")
	log.Println("")
}

func buildUrl(r *http.Request, path, query string) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	// Encrypt activation URL
	u := &url.URL{
		Scheme:   scheme,
		Host:     r.Host,
		Path:     path,
		RawQuery: query,
	}
	return u.String()
}

func mustBeAuthenticated(w http.ResponseWriter, r *http.Request) *SessionObject {
	// Get session
	session, err := CurrentSession(r)
	if !assertErrorIsNil(w, err, "Error creating session") {
		return nil
	} else if session.IsEmpty() {
		http.Redirect(w, r, LoginURL, http.StatusSeeOther)
		return nil
	}
	return session
}

func assertErrorIsNil(w http.ResponseWriter, err error, logMessages ...string) bool {
	if err != nil {
		logError(err, logMessages...)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return false
	}
	return true
}
