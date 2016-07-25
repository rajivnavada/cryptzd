package web

import (
	"bytes"
	"cryptzd/crypto"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	pb "github.com/rajivnavada/cryptz_pb"
	"strings"
	"sync"
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
	maxMessageSize = 4096
)

var (
	ErrDuplicateFingerprint    = errors.New("New connection attempted with duplicate fingerprint. Selecting new connection over old.")
	ErrInvalidArgsForProjectOp = errors.New("Project operation received invalid arguments. Please make sure all required arguments are provided.")
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  maxMessageSize * 2,
	WriteBufferSize: maxMessageSize * 2,
}

type fingerprint string
type publicKeyId int
type userId int

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type Hub struct {
	// Registered connections.
	connections map[fingerprint]*connection

	// Channel to broadcast messages to connected users
	broadcastMessage chan map[string]crypto.EncryptedMessage

	// Channel to broadcast new user activations
	broadcastUser chan messagesTemplateExtensions

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

var H = Hub{
	broadcastMessage: make(chan map[string]crypto.EncryptedMessage),
	broadcastUser:    make(chan messagesTemplateExtensions),
	register:         make(chan *connection),
	unregister:       make(chan *connection),
	connections:      make(map[fingerprint]*connection),
}

// Run makes the hub ready to receive / broadcast connections
func (h *Hub) Run() {
	// NOTE: the reason the delete's don't need to be guarded by a mutex here is because each
	//       'case' is handled synchronously.
	for {
		select {
		case c := <-h.register:
			// If we are trying to register a connection for an existing fingerprint,
			// close that connection first
			if oldC, ok := h.connections[c.fingerprint]; ok {
				logError(ErrDuplicateFingerprint, "Error maintaining connection with duplicate key")
				delete(h.connections, c.fingerprint)
				oldC.closeChan()
			}
			h.connections[c.fingerprint] = c

		case c := <-h.unregister:
			if _, ok := h.connections[c.fingerprint]; ok {
				delete(h.connections, c.fingerprint)
				c.closeChan()
			}

		case messages := <-h.broadcastMessage:
			// m is a map of fingerprint to message
			for k, m := range messages {
				// For each key, find if we have an active connection
				if c, ok := h.connections[fingerprint(k)]; ok {
					// Prepare a bytes buffer to collect the output
					buf := &bytes.Buffer{}
					var err error
					if c.isCLI {
						err = messageTextTemplate.Execute(buf, m)
					} else {
						err = messageTemplate.Execute(buf, m)
					}
					// If there is an active connection, send message
					if err != nil {
						logError(err, "Error constructing message HTML")
					} else {
						select {
						case c.send <- buf.Bytes():
						default:
							delete(h.connections, fingerprint(k))
							c.closeChan()
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
				for k, c := range h.connections {
					select {
					case c.send <- buf.Bytes():
					default:
						delete(h.connections, fingerprint(k))
						c.closeChan()
					}
				}
			}
		}
	}
}

// Close closes all open connections and destroys the hub
func (h *Hub) Close() {
	// closes all open connections
	// Loops over all connenctions and closes connections
	for _, c := range h.connections {
		h.unregister <- c
	}
	// Also closes the broadcast channels
	close(h.broadcastMessage)
	close(h.broadcastUser)
	close(h.register)
	close(h.unregister)
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Protects the send channel
	lock sync.Locker
	// Buffered channel of outbound messages.
	send chan []byte
	// Records if this connection is closed
	closed bool

	// userId of the user this connection belongs to
	userId userId

	keyId publicKeyId

	// fingerprint of the key used in this connection
	fingerprint fingerprint

	isCLI bool
}

func (c *connection) closeChan() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return
	}
	c.ws.Close()
	close(c.send)
	c.closed = true
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump() {
	defer func() {
		// unregister should also close the channel
		// no need to call closeChan here
		H.unregister <- c
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		messageType, messageBody, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				logError(err, "Error in websocket readPump")
			}
			break
		}

		if messageType != websocket.BinaryMessage {
			continue
		}

		// Unmarshal the message
		opQuery := &pb.Operation{}
		err = proto.Unmarshal(messageBody, opQuery)
		if err != nil {
			logError(err, "Error unmarshaling operation query in readPump")
			continue
		}

		projectOp := opQuery.GetProjectOp()
		credOp := opQuery.GetCredentialOp()
		result := &pb.Response{}

		// Perform the operation requested in the message (possibly by spawning a goroutine)
		if projectOp != nil {

			core := &pb.ProjectOperationResponse{
				Command: projectOp.Command,
			}

			result.ProjectOrCredentialResponse = &pb.Response_ProjectOpResponse{
				ProjectOpResponse: core,
			}

			switch projectOp.Command {
			case pb.ProjectOperation_LIST:

			case pb.ProjectOperation_CREATE:
				project, err := c.createProject(projectOp)
				if err != nil {
					logError(err, "Error while creating project")
					result.Status = pb.Response_ERROR
					result.Error = err.Error()
				} else {
					result.Status = pb.Response_SUCCESS
					result.Info = fmt.Sprintf("Successfully created project with ID = %d", project.Id())
					core.Project = &pb.Project{
						Id:          int32(project.Id()),
						Name:        project.Name(),
						Environment: project.Environment(),
					}
				}

			case pb.ProjectOperation_UPDATE:

			case pb.ProjectOperation_DELETE:

			case pb.ProjectOperation_ADD_MEMBER:

			case pb.ProjectOperation_DELETE_MEMBER:

			case pb.ProjectOperation_LIST_CREDENTIALS:
			}

		} else if credOp != nil {

			switch credOp.Command {
			case pb.CredentialOperation_GET:
				// SELECT the credential using the project_id and key_id

			case pb.CredentialOperation_SET:

			case pb.CredentialOperation_DELETE:
			}
		}

		// Send back the response by calling c.send
		msg, err := proto.Marshal(result)
		if err != nil {
			logError(err, "Error while marshaling operation result")
			continue
		}
		c.send <- msg
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
		// We just close the websocket connection. Then in the readPump, the close is
		// detected and that exits as well
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			messageType := websocket.TextMessage
			if c.isCLI {
				messageType = websocket.BinaryMessage
			}
			if err := c.write(messageType, message); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *connection) createProject(op *pb.ProjectOperation) (crypto.Project, error) {
	if !c.isCLI {
		return nil, ErrInvalidArgsForProjectOp
	}
	name := strings.TrimSpace(op.Name)
	environ := strings.TrimSpace(op.Environment)
	// Make sure we have all the requirements to perform the operation
	if name == "" {
		return nil, ErrInvalidArgsForProjectOp
	}
	// Get a mapper
	dbMap, err := crypto.NewDataMapper()
	if err != nil {
		return nil, err
	}
	defer dbMap.Close()

	// Create a project with name/environment.
	project := crypto.NewProject(name, environ, "")
	if err = project.Save(dbMap); err != nil {
		return nil, err
	}
	// Add a member to the project by granting current userId admin access
	if _, err = project.AddMember(int(c.userId), dbMap); err != nil {
		return nil, err
	}
	// Return the new project
	return project, nil
}

func newConnection(wsConn *websocket.Conn, uid userId, keyId publicKeyId, fpr fingerprint, isCLI bool) *connection {
	return &connection{
		lock:        &sync.Mutex{},
		send:        make(chan []byte, 256),
		ws:          wsConn,
		userId:      uid,
		keyId:       keyId,
		fingerprint: fpr,
		isCLI:       isCLI,
	}
}
