package main

import (
	"cryptz/crypto"
	"cryptz/mail"
	"cryptz/web"
	"flag"
	"net/http"
	"os"
)

var (
	host                    = flag.String("host", "127.0.0.1", "HTTP service host")
	port                    = flag.String("port", "8000", "HTTP port at which the service will run")
	mongoHost               = flag.String("mongoHost", "127.0.0.1", "MongoDB host")
	mongoDbName             = flag.String("mongoDbName", "cryptz", "MongoDB database name")
	appEmail                = flag.String("appEmail", "zocmyworld@gmail.com", "Email address to use for sender for this app")
	appEmailPasswordEnvName = flag.String("appPasswordEnvName", "MAILPASS", "Name of the environment variable that contains the password for this app email sender")
)

func main() {
	flag.Parse()

	// Check mongo service
	crypto.InitService(*mongoHost, *mongoDbName)
	mail.InitService(*appEmail, os.Getenv(*appEmailPasswordEnvName))

	// TODO: start the connection hub for websocket stuff
	go web.H.Run()
	defer web.H.Close()

	router := web.Router()
	addr := *host + ":" + *port

	println("Will start http server at:", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	if err := server.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
		panic(err)
	}
}
