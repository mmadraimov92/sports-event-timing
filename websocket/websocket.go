package websocket

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{}

// WSManager manages WebSocket connections for athletes.Service
//
// AddClient adds new client to the map of connected clients and returns unique clientID
//
// StartClient calls start method for the connectedWSClient.
//
// SendMessageToAll connected clients
//
// SendMessageToOne client specified by clientID
//
// Upgrade upgrades http connection to ws by calling websocket.Upgrader.Upgrade.
// Also does origin check, but currently all requests are permitted.
type WSManager interface {
	AddClient(*websocket.Conn, *logrus.Logger) string
	StartClient(string)
	SendMessageToAll([]byte)
	SendMessageToOne([]byte, string)
	Upgrade(http.ResponseWriter, *http.Request, http.Header) (*websocket.Conn, error)
}

// wsManager implements WSManager.
// Also holds map of connected WebSocket clients and closeWS channel
type wsManager struct {
	connectedWSClients map[string]connectedWSClient
	closeWS            chan string
}

func (wsm wsManager) AddClient(ws *websocket.Conn, logger *logrus.Logger) string {
	clientID := uuid.New().String()
	logger.Infof("Client %s: connected", clientID)
	client := connectedWSClient{
		UUID:               clientID,
		Conn:               ws,
		ReceivedDisconnect: make(chan bool),
		SendUpdates:        make(chan []byte),
		logger:             logger,
	}
	wsm.connectedWSClients[clientID] = client
	return client.UUID
}

// StartClient calls start method for the connectedWSClient. Once start method returned,
// sends clientID to wsManager.closeWS channel
func (wsm wsManager) StartClient(clientID string) {
	wsm.connectedWSClients[clientID].start()
	wsm.closeWS <- clientID
}

func (wsm wsManager) SendMessageToAll(msg []byte) {
	for _, client := range wsm.connectedWSClients {
		go func(client connectedWSClient) { client.SendUpdates <- msg }(client)
	}
}

func (wsm wsManager) SendMessageToOne(msg []byte, clientID string) {
	go func() { wsm.connectedWSClients[clientID].SendUpdates <- msg }()
}

func (wsm wsManager) Upgrade(w http.ResponseWriter, r *http.Request, h http.Header) (*websocket.Conn, error) {
	// Allow all requests
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, h)
	if err != nil {
		return nil, err
	}
	return ws, nil
}

// NewWSManager initializes WSManager object and
// starts goroutine for cleanup process.
//
// Cleanup goroutine listens to wsManager.closeWS channel and removing
// clients from wsManager.connectedWSClients map by received clientID
func NewWSManager() WSManager {
	wsm := wsManager{map[string]connectedWSClient{}, make(chan string)}
	// Start cleanup goroutine
	go func(ws wsManager) {
		for uuid := range ws.closeWS {
			client := ws.connectedWSClients[uuid]
			client.Conn.Close()
			close(client.ReceivedDisconnect)
			close(client.SendUpdates)
			delete(ws.connectedWSClients, uuid)
		}
	}(wsm)
	return wsm
}

type connectedWSClient struct {
	UUID               string
	Conn               *websocket.Conn
	SendUpdates        chan []byte
	ReceivedDisconnect chan bool
	logger             *logrus.Logger
}

// start starts sendMessage and readMessage goroutines.
// Waits for done signal from sendMessage
func (client connectedWSClient) start() {
	done := make(chan bool)
	go client.sendMessage(done)
	go client.readMessage()
	<-done
}

// readMessage starts loop to constantly read incoming message from client.
// All messages are discarded. In case of error, signals client.ReceivedDisconnect channel.
func (client connectedWSClient) readMessage() {
	for {
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			client.logger.Infof("Client %s: %v", client.UUID, err.Error())
			client.logger.Infof("Client %s: closing WS connection", client.UUID)
			client.ReceivedDisconnect <- true
			return
		}
	}
}

// sendMessage starts listening to client.ReceivedDisconnect and
// client.SendUpdates channels.
//
// Receiving from client.ReceivedDisconnect meaning connection is no longer present,
// signals done channel and returns
//
// Receiving from client.SendUpdates, sends ws message to the client. In case of an error,
// signal doen channel and return
func (client connectedWSClient) sendMessage(done chan bool) {
	for {
		select {
		case <-client.ReceivedDisconnect:
			done <- true
			return
		case msg := <-client.SendUpdates:
			client.logger.Infof("Client %s: sending message %s", client.UUID, string(msg))
			err := client.Conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				client.logger.Errorf("Client %s: %v", client.UUID, err.Error())
				done <- true
				return
			}
		}
	}

}
