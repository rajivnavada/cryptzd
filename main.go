package main

import (
	"cryptzd/crypto"
	"cryptzd/mail"
	"cryptzd/web"
	"flag"
	"fmt"
	"net/http"
	"os"
)

var (
	host                    = flag.String("host", "127.0.0.1", "HTTP service host")
	port                    = flag.String("port", "8000", "HTTP port at which the service will run")
	sqliteFilePath          = flag.String("db", "/usr/local/var/db/cryptz/cryptz.db", "Path to sqllite db file.")
	appEmail                = flag.String("appEmail", "zocmyworld@gmail.com", "Email address to use for sender for this app")
	appEmailPasswordEnvName = flag.String("appPasswordEnvName", "MAILPASS", "Name of the environment variable that contains the password for this app email sender")
	debug                   = flag.Bool("debug", false, "Turn on debug mode")
)

func main() {
	flag.Parse()

	// Init services
	crypto.InitService(*sqliteFilePath, *debug)
	mail.InitService(*appEmail, os.Getenv(*appEmailPasswordEnvName))

	// start the connection hub for websocket stuff
	go web.H.Run()
	defer web.H.Close()

	router := web.Router()
	addr := *host + ":" + *port

	if *debug {
		fmt.Printf("Starting http server at: https://%s\n", addr)
	}

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	if err := server.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
		panic(err)
	}
}
