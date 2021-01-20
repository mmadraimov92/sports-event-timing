package websocket

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var wsm WSManager

func TestMain(m *testing.M) {
	wsm = NewWSManager()
	code := m.Run()
	os.Exit(code)
}

var firstMsg = "welcome ws client"
var updateMsg1 = "update 1"
var updateMsg2 = "update 2"

func handler(w http.ResponseWriter, r *http.Request) {
	ws, err := wsm.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	clientID := wsm.AddClient(ws, logrus.New())
	go wsm.SendMessageToOne([]byte(firstMsg), clientID)
	wsm.StartClient(clientID)
}

func TestWebSocketManager(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(handler))
	defer s.Close()

	u, err := url.Parse(s.URL)
	if err != nil {
		log.Fatal(err.Error())
	}
	u.Scheme = "ws"

	client1, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer client1.Close()

	_, msg, err := client1.ReadMessage()
	assert.Equal(t, nil, err)
	assert.Equal(t, firstMsg, string(msg))

	wsm.SendMessageToAll([]byte(updateMsg1))

	_, msg, err = client1.ReadMessage()
	assert.Equal(t, nil, err)
	assert.Equal(t, updateMsg1, string(msg))

	client2, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer client2.Close()
	_, msg, err = client2.ReadMessage()
	assert.Equal(t, nil, err)
	assert.Equal(t, firstMsg, string(msg))

	wsm.SendMessageToAll([]byte(updateMsg2))
	_, msg, err = client1.ReadMessage()
	assert.Equal(t, nil, err)
	assert.Equal(t, updateMsg2, string(msg))
	_, msg, err = client2.ReadMessage()
	assert.Equal(t, nil, err)
	assert.Equal(t, updateMsg2, string(msg))
}
