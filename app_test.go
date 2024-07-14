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

func SetupWsServer(t *testing.T) (*httptest.Server, string) {
	app := App{}
	app.Init()
	testServer := httptest.NewServer(app.MainHandler())

	// Convert http://127.0.0.1 to ws://127.0.0/api/v1/ws
	wsUrl := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/api/v1/ws"

	return testServer, wsUrl
}

func TestServerStartsAndWebsocketCanConnect(t *testing.T) {
	testServer, wsUrl := SetupWsServer(t)
	defer testServer.Close()

	ws, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		t.Fatalf("Failed to connect to websocket server: %v", err)
	}
	defer ws.Close()

	// Send message to server, read response and check to see if it's what we expect.
	_, _, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func TestFirstMessageIsClientConnect(t *testing.T) {
	testServer, wsUrl := SetupWsServer(t)
	defer testServer.Close()

	ws, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		t.Fatalf("Failed to connect to websocket server: %v", err)
	}
	defer ws.Close()

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

func TestFirstMessageHasClient(t *testing.T) {
	testServer, wsUrl := SetupWsServer(t)
	defer testServer.Close()

	ws, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		t.Fatalf("Failed to connect to websocket server: %v", err)
	}
	defer ws.Close()

	_, msgBytes, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var clientConnectMsg ClientConnectMsg
	err = json.Unmarshal(msgBytes, &clientConnectMsg)
	if err != nil {
		t.Fatalf("Error parsing message json as ClientConnectMsg: %v", err)
	}
}

func TestTwoClientsCanConnectAndHaveDifferentIds(t *testing.T) {
	testServer, wsUrl := SetupWsServer(t)
	defer testServer.Close()

	// Client 1 connect
	ws, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		t.Fatalf("Failed to connect to websocket server: %v", err)
	}
	defer ws.Close()

	_, msgBytes, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var client1ConnectMsg ClientConnectMsg
	err = json.Unmarshal(msgBytes, &client1ConnectMsg)
	if err != nil {
		t.Fatalf("Error parsing client1 message json as ClientConnectMsg: %v", err)
	}

	// Client 2 connect
	ws2, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		t.Fatalf("Failed to connect client 1 to websocket server: %v", err)
	}
	defer ws2.Close()

	_, msgBytes, err = ws2.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to connect client 2 to websocket server: %v", err)
	}

	var client2ConnectMsg ClientConnectMsg
	err = json.Unmarshal(msgBytes, &client2ConnectMsg)
	if err != nil {
		t.Fatalf("Error parsing client2 message json as ClientConnectMsg: %v", err)
	}

	if client1ConnectMsg.Client.ID == client2ConnectMsg.Client.ID {
		t.Fatalf("Expected clients to each have unique id but both were %s", client2ConnectMsg.Client.ID)
	}
}
