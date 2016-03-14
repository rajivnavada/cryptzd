package web

import (
	"bytes"
	"cryptz/crypto"
	"github.com/gorilla/websocket"
	"time"
)

// NOTE: Most code is borrowed from github.com/gorilla/websocket/examples/chat

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type fingerprint string
type userId string

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type Hub struct {
	// Registered connections.
	connections map[fingerprint]*connection

	// Channel to broadcast messages to connected users
	broadcastMessage chan map[string]crypto.Message

	// Channel to broadcast new user activations
	broadcastUser chan messagesTemplateExtensions

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

var H = Hub{
	broadcastMessage: make(chan map[string]crypto.Message),
	broadcastUser:    make(chan messagesTemplateExtensions),
	register:         make(chan *connection),
	unregister:       make(chan *connection),
	connections:      make(map[fingerprint]*connection),
}

// Run makes the hub ready to receive / broadcast connections
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			// If we are trying to register a connection for an existing fingerprint,
			// close that connection first
			if oldC, ok := h.connections[c.fingerprint]; ok {
				delete(h.connections, c.fingerprint)
				close(oldC.send)
			}
			h.connections[c.fingerprint] = c

		case c := <-h.unregister:
			if _, ok := h.connections[c.fingerprint]; ok {
				delete(h.connections, c.fingerprint)
				close(c.send)
			}

		case messages := <-h.broadcastMessage:
			// m is a map of fingerprint to message
			for k, m := range messages {
				// For each key, find if we have an active connection
				if c, ok := h.connections[fingerprint(k)]; ok {
					// Prepare a bytes buffer to collect the output
					buf := &bytes.Buffer{}
					// If there is an active connection, send message
					if err := messageTemplate.Execute(buf, m); err != nil {
						logError(err, "Error constructing message HTML")
					} else {
						select {
						case c.send <- buf.Bytes():
						default:
							delete(h.connections, fingerprint(k))
							close(c.send)
						}
					}
				}
			}

		case user := <-h.broadcastUser:
			// Prepare a bytes buffer to collect the output
			buf := &bytes.Buffer{}
			// If there is an active connection, send message
			if err := userTemplate.Execute(buf, user); err != nil {
				logError(err, "Error constructing user HTML")
			} else {
				for _, c := range h.connections {
					c.send <- buf.Bytes()
				}
			}
		}
	}
}

// Close closes all open connections and destroys the hub
func (h *Hub) Close() {
	// TODO: closes all open connections
	// Loops over all connenctions and writes a close message to them
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// userId of the user this connection belongs to
	userId userId

	// fingerprint of the key used in this connection
	fingerprint fingerprint
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump() {
	defer func() {
		H.unregister <- c
		c.ws.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, _, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				logError(err, "Error in websocket readPump")
			}
			break
		}
	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
