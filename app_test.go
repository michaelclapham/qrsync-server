package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestServerStarts(t *testing.T) {
	app := App{}
	app.Init()
	testServer := httptest.NewServer(app.MainHandler())
	defer testServer.Close()

	// Connect to the server
	clientsUrl := fmt.Sprintf("%s%s", testServer.URL, "")
	_, err := http.Get(clientsUrl)
	if err != nil {
		t.Fail()
	}
	log.Printf("%s", clientsUrl)
}

func SetupServerAndConnect(t *testing.T) (*websocket.Conn, *httptest.Server) {
	app := App{}
	app.Init()
	testServer := httptest.NewServer(app.MainHandler())

	// Convert http://127.0.0.1 to ws://127.0.0/api/v1/ws
	u := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/api/v1/ws"

	log.Printf("Url %s", u)

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	return ws, testServer
}

func TestServerStartsAndWebsocketCanConnect(t *testing.T) {
	ws, testServer := SetupServerAndConnect(t)
	defer ws.Close()
	defer testServer.Close()

	// Send message to server, read response and check to see if it's what we expect.
	_, _, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func TestFirstMessageIsClientConnect(t *testing.T) {
	ws, testServer := SetupServerAndConnect(t)
	defer ws.Close()
	defer testServer.Close()

	_, msgBytes, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var connectMsgJson map[string]interface{}
	json.Unmarshal(msgBytes, &connectMsgJson)
	if connectMsgJson["type"] != "ClientConnect" {
		t.Fatalf("Expected type to be ClientConnect but was %s", connectMsgJson["type"])
	}
}
