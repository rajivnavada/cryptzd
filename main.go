package main

import (
	"bytes"
	"crypto/x509"
	"cryptz/crypto"
	"cryptz/mail"
	"cryptz/web"
	"flag"
	"fmt"
	"io"
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
	debug                   = flag.Bool("debug", false, "Turn on debug mode")
	receiver                = flag.Bool("receiver", false, "Start in receiver mode")
	receiverKey             = flag.String("receiverKey", "", "Fingerprint of key to use in receiver mode")
)

func startReceiver() {
	// Read cert.pem and key.pem into a buffer
	buf := &bytes.Buffer{}
	for _, fname := range []string{"cert.pem", "key.pem"} {
		if f, err := os.Open(fname); err != nil {
			panic(err)
		} else {
			io.Copy(buf, f)
		}
	}

	// Create a cert pool
	certs := x509.NewCertPool()
	if !certs.AppendCertsFromPEM(buf.Bytes()) {
		println("Could not parse cert from PEM")
		return
	}

	wssurl := fmt.Sprintf("wss://%s:%s/ws/%s", *host, *port, *receiverKey)
	origin := fmt.Sprintf("https://%s:%s", *host, *port)

	// Start a websocket client and hopefully it will receive messages
	client := web.NewWSClient(wssurl, origin, certs)
	if err := client.Run(); err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	if *receiver {
		startReceiver()
		return
	}

	// Check mongo service
	crypto.InitService(*mongoHost, *mongoDbName)
	mail.InitService(*appEmail, os.Getenv(*appEmailPasswordEnvName))

	// start the connection hub for websocket stuff
	go web.H.Run()
	defer web.H.Close()

	router := web.Router()
	addr := *host + ":" + *port

	if *debug {
		println("Will start http server at:", addr)
	}

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	if err := server.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
		panic(err)
	}
}
