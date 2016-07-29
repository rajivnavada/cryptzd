package web

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	pb "github.com/rajivnavada/cryptz_pb"
	"github.com/rajivnavada/cryptzd/crypto"
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
	ErrDuplicateFingerprint       = errors.New("New connection attempted with duplicate fingerprint. Selecting new connection over old.")
	ErrInvalidArgsForProjectOp    = errors.New("Project operation received invalid arguments. Please make sure all required arguments are provided.")
	ErrInvalidArgsForCredentialOp = errors.New("Credential operation received invalid arguments. Please make sure all required arguments are provided.")
	ErrNoAccess                   = errors.New("You do not have permission to perform this operation.")
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
		result := &pb.Response{
			Status: pb.Response_ERROR,
			Error:  "This operation is temporarily unsupported",
		}

		// Perform the operation requested in the message (possibly by spawning a goroutine)
		if projectOp != nil {

			core := &pb.ProjectOperationResponse{
				Command: projectOp.Command,
			}
			result.ProjectOpResponse = core

			switch projectOp.Command {
			case pb.ProjectOperation_LIST:
				projects, err := c.listProjects(projectOp)
				if err != nil {
					logError(err, "Error when listing projects")
					result.Status = pb.Response_ERROR
					result.Error = err.Error()
				} else {
					result.Status = pb.Response_SUCCESS
					label := "projects"
					if len(projects) == 1 {
						label = "project"
					}
					result.Info = fmt.Sprintf("Found %d %s", len(projects), label)
					result.Error = ""
					core.Projects = projects
				}

			case pb.ProjectOperation_CREATE:
				project, err := c.createProject(projectOp)
				if err != nil {
					logError(err, "Error while creating project")
					result.Status = pb.Response_ERROR
					result.Error = err.Error()
				} else {
					result.Status = pb.Response_SUCCESS
					result.Info = fmt.Sprintf("Successfully created project with ID = %d", project.Id)
					result.Error = ""
					core.Project = project
				}

			case pb.ProjectOperation_UPDATE:

			case pb.ProjectOperation_DELETE:

			case pb.ProjectOperation_ADD_MEMBER:
				memberId, err := c.addMember(projectOp)
				if err != nil {
					logError(err, "Error while adding member to project")
					result.Status = pb.Response_ERROR
					result.Error = err.Error()
				} else {
					result.Status = pb.Response_SUCCESS
					result.Info = fmt.Sprintf("Successfully added %s (member ID = %d) to project with ID = %d", projectOp.MemberEmail, memberId, projectOp.ProjectId)
					result.Error = ""
				}

			case pb.ProjectOperation_DELETE_MEMBER:
				err := c.deleteMember(projectOp)
				if err != nil {
					logError(err, "Error while deleting member from project")
					result.Status = pb.Response_ERROR
					result.Error = err.Error()
				} else {
					result.Status = pb.Response_SUCCESS
					result.Info = fmt.Sprintf("Successfully deleted member with ID = %d from project with ID = %d", projectOp.MemberId, projectOp.ProjectId)
					result.Error = ""
				}

			case pb.ProjectOperation_LIST_CREDENTIALS:
				creds, err := c.listCredentials(projectOp)
				if err != nil {
					logError(err, "Error while listing project credentials")
					result.Status = pb.Response_ERROR
					result.Error = err.Error()
				} else {
					result.Status = pb.Response_SUCCESS
					label := "credentials"
					if len(creds) == 1 {
						label = "credential"
					}
					result.Info = fmt.Sprintf("Found %d %s for project with ID = %d", len(creds), label, projectOp.ProjectId)
					result.Error = ""
					core.Credentials = creds
				}

			case pb.ProjectOperation_GET_CREDENTIAL:
				cred, err := c.getCredential(projectOp)
				if err != nil {
					logError(err, "Error while getting a credential")
					result.Status = pb.Response_ERROR
					result.Error = err.Error()
				} else {
					result.Status = pb.Response_SUCCESS
					result.Info = ""
					result.Error = ""
					core.Credential = cred
				}

			case pb.ProjectOperation_ADD_CREDENTIAL:
				cred, err := c.setCredential(projectOp)
				if err != nil {
					logError(err, fmt.Sprintf("Error while setting credential with key '%s'", projectOp.Key))
					result.Status = pb.Response_ERROR
					result.Error = err.Error()
				} else {
					result.Status = pb.Response_SUCCESS
					result.Info = fmt.Sprintf("Successfully set credential with key '%s'", projectOp.Key)
					result.Error = ""
					core.Credential = cred
				}

			case pb.ProjectOperation_DELETE_CREDENTIAL:
				err := c.deleteCredential(projectOp)
				if err != nil {
					logError(err, fmt.Sprintf("Error while deleting credential with key '%s'", projectOp.Key))
					result.Status = pb.Response_ERROR
					result.Error = err.Error()
				} else {
					result.Status = pb.Response_SUCCESS
					result.Info = fmt.Sprintf("Successfully deleted credential with key '%s'", projectOp.Key)
					result.Error = ""
				}
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

func (c *connection) listProjects(op *pb.ProjectOperation) ([]*pb.Project, error) {
	if !c.isCLI {
		return nil, ErrInvalidArgsForProjectOp
	}

	dbMap, err := crypto.NewDataMapper()
	if err != nil {
		return nil, err
	}
	defer dbMap.Close()

	projects, err := crypto.FindProjectsForUser(int(c.userId), dbMap)
	if err != nil {
		return nil, err
	}

	var ret []*pb.Project
	for _, p := range projects {
		ret = append(ret, &pb.Project{
			Id:          int32(p.Id()),
			Name:        fmt.Sprintf("[%s] %s", p.DefaultAccessLevel(), p.Name()),
			Environment: p.Environment(),
		})
	}
	return ret, nil
}

func (c *connection) createProject(op *pb.ProjectOperation) (*pb.Project, error) {
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
	if _, err = project.AddMember(int(c.userId), crypto.ACCESS_LEVEL_ADMIN, dbMap); err != nil {
		return nil, err
	}
	ret := pb.Project{
		Id:          int32(project.Id()),
		Name:        project.Name(),
		Environment: project.Environment(),
	}
	// Return the new project
	return &ret, nil
}

func (c *connection) addMember(op *pb.ProjectOperation) (int32, error) {
	if !c.isCLI {
		return 0, ErrInvalidArgsForProjectOp
	}

	// Validate important input
	projectId := int(op.ProjectId)
	memberEmail := op.MemberEmail
	// Make sure we have all the requirements to perform the operation
	if projectId == 0 || memberEmail == "" {
		return 0, ErrInvalidArgsForProjectOp
	}

	accessLevel := op.AccessLevel
	if accessLevel == "" {
		accessLevel = crypto.ACCESS_LEVEL_READ
	}

	// Get a mapper
	dbMap, err := crypto.NewDataMapper()
	if err != nil {
		return 0, err
	}
	defer dbMap.Close()

	// Create a project with name/environment.
	p, err := crypto.FindProjectWithId(projectId, dbMap)
	if err != nil {
		return 0, err
	}

	// Assert that the current user has admin access to the project
	if !p.HasAdminWithUserId(int(c.userId), dbMap) {
		return 0, ErrNoAccess
	}

	u, err := crypto.FindUserWithEmail(memberEmail, dbMap)
	if err != nil {
		return 0, err
	}

	// Add a member to the project by granting current userId admin access
	m, err := p.AddMember(int(u.Id()), accessLevel, dbMap)
	if err != nil {
		return 0, err
	}

	// Return the new project
	return int32(m.Id()), nil
}

func (c *connection) deleteMember(op *pb.ProjectOperation) error {
	if !c.isCLI {
		return ErrInvalidArgsForProjectOp
	}

	// Validate important input
	memberId := int(op.MemberId)

	// Make sure we have all the requirements to perform the operation
	if memberId == 0 {
		return ErrInvalidArgsForProjectOp
	}

	// Get a mapper
	dbMap, err := crypto.NewDataMapper()
	if err != nil {
		return err
	}
	defer dbMap.Close()

	// Find member
	m, err := crypto.FindProjectMemberWithId(memberId, dbMap)
	if err != nil {
		return err
	}

	// Make sure the user is an admin of the project
	p, err := crypto.FindProjectWithId(m.ProjectId(), dbMap)
	if err != nil {
		return ErrNoAccess
	}

	// Assert that the current user has admin access to the project
	if !p.HasAdminWithUserId(int(c.userId), dbMap) {
		return ErrNoAccess
	}

	// Return the new project
	return m.Delete(dbMap)
}

func (c *connection) getCredential(op *pb.ProjectOperation) (*pb.Credential, error) {
	if !c.isCLI {
		return nil, ErrInvalidArgsForCredentialOp
	}
	// Validate important input
	projectId := int(op.ProjectId)
	key := strings.TrimSpace(op.Key)
	// Make sure we have all the requirements to perform the operation
	if projectId == 0 || key == "" {
		return nil, ErrInvalidArgsForCredentialOp
	}

	// Get a mapper
	dbMap, err := crypto.NewDataMapper()
	if err != nil {
		return nil, err
	}
	defer dbMap.Close()

	// Create a project with name/environment.
	p, err := crypto.FindProjectWithId(projectId, dbMap)
	if err != nil {
		return nil, err
	}

	pv, err := p.GetCredential(key, int(c.keyId), dbMap)
	if err != nil {
		return nil, err
	}

	cred := pb.Credential{
		Id:     int32(pv.CredentialId()),
		Key:    key,
		Cipher: string(pv.Cipher()),
	}

	// Return the new project
	return &cred, nil
}

func (c *connection) listCredentials(op *pb.ProjectOperation) ([]*pb.Credential, error) {
	if !c.isCLI {
		return nil, ErrInvalidArgsForProjectOp
	}
	// Validate important input
	projectId := int(op.ProjectId)
	// Make sure we have all the requirements to perform the operation
	if projectId == 0 {
		return nil, ErrInvalidArgsForProjectOp
	}

	// Get a mapper
	dbMap, err := crypto.NewDataMapper()
	if err != nil {
		return nil, err
	}
	defer dbMap.Close()

	p, err := crypto.FindProjectWithId(projectId, dbMap)
	if err != nil {
		return nil, err
	}

	pcList, err := p.Credentials(dbMap)
	if err != nil {
		return nil, err
	}

	var ret []*pb.Credential
	for _, pc := range pcList {
		ret = append(ret, &pb.Credential{
			Id:  int32(pc.Id()),
			Key: pc.Key(),
		})
	}

	return ret, nil
}

func (c *connection) setCredential(op *pb.ProjectOperation) (*pb.Credential, error) {
	if !c.isCLI {
		return nil, ErrInvalidArgsForCredentialOp
	}
	// Validate important input
	projectId := int(op.ProjectId)
	key := strings.TrimSpace(op.Key)
	value := op.Value
	// Make sure we have all the requirements to perform the operation
	if projectId == 0 || key == "" || value == "" {
		return nil, ErrInvalidArgsForCredentialOp
	}

	// Get a mapper
	dbMap, err := crypto.NewDataMapper()
	if err != nil {
		return nil, err
	}
	defer dbMap.Close()

	// Create a project with name/environment.
	p, err := crypto.FindProjectWithId(projectId, dbMap)
	if err != nil {
		return nil, err
	}

	// Assert that the current user has admin access to the project
	if !p.HasAdminWithUserId(int(c.userId), dbMap) {
		return nil, ErrNoAccess
	}

	pc, err := p.SetCredential(key, value, dbMap)
	if err != nil {
		return nil, err
	}

	cred := pb.Credential{
		Id:  int32(pc.Id()),
		Key: key,
	}

	// Return the new project
	return &cred, nil
}

func (c *connection) deleteCredential(op *pb.ProjectOperation) error {
	if !c.isCLI {
		return ErrInvalidArgsForCredentialOp
	}
	// Validate important input
	projectId := int(op.ProjectId)
	key := strings.TrimSpace(op.Key)
	// Make sure we have all the requirements to perform the operation
	if projectId == 0 || key == "" {
		return ErrInvalidArgsForCredentialOp
	}

	// Get a mapper
	dbMap, err := crypto.NewDataMapper()
	if err != nil {
		return err
	}
	defer dbMap.Close()

	// Create a project with name/environment.
	p, err := crypto.FindProjectWithId(projectId, dbMap)
	if err != nil {
		return err
	}

	// Assert that the current user has admin access to the project
	if !p.HasAdminWithUserId(int(c.userId), dbMap) {
		return ErrNoAccess
	}

	err = p.RemoveCredential(key, dbMap)
	if err != nil {
		return err
	}
	return nil
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
