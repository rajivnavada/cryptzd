package main

import (
	"cryptz/web"
	"net/http"
	"os"
)

//var port = flag.String("addr", "8000", "HTTP port at which the service will run")
//var host = flag.String("host", "127.0.0.1", "HTTP service host")
//

func main() {
	// TODO: start the connection hub for websocket stuff
	go web.H.Run()
	defer web.H.Close()

	//flag.Parse()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	router := web.Router()
	addr := "127.0.0.1:" + port

	println("Will start http server at:", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	if err := server.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
		panic(err)
	}
}
