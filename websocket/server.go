package websocket

import (
	"bytes"
	"fmt"
	ws "golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
	"time"
)

// Echo the data received on the WebSocket.
func EchoServer(c *ws.Conn) {
	var stdout bytes.Buffer

	w := io.MultiWriter(&stdout, c)
	if _, err := io.Copy(w, c); err != nil {
		log.Fatal(err)
	}

	log.Println(stdout.String())
}

// Sends random strings to the client
func HelloServer(c *ws.Conn) {
	fmt.Println(fmt.Sprintf("Connection made to %s", c.RemoteAddr()))
	for {
		c.Write([]byte("Hello world!"))
		log.Println("Sent a message")
		time.Sleep(60 * time.Second)
	}
}

// This example demonstrates a trivial echo server.
func StartServer() {
	http.Handle("/", ws.Handler(EchoServer))
	http.Handle("/hw", ws.Handler(HelloServer))

	err := http.ListenAndServe("127.0.0.1:8888", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
