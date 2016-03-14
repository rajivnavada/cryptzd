package main

import (
	"cryptz/web"
	"flag"
	"net/http"
)

var host = flag.String("host", "127.0.0.1", "HTTP service host")
var port = flag.String("port", "8000", "HTTP port at which the service will run")

func main() {
	// TODO: start the connection hub for websocket stuff
	go web.H.Run()
	defer web.H.Close()

	flag.Parse()

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
