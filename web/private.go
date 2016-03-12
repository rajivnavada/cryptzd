package web

import (
	"log"
	"net/http"
	"net/url"
)

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
