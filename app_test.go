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

func CloseWithCloseMessage(conn *websocket.Conn) {
	conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Normal close"),
	)
	conn.Close()
}

func TestServerStartsAndWebsocketCanConnect(t *testing.T) {
	testServer, wsUrl := SetupWsServer(t)
	defer testServer.Close()

	ws, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		t.Fatalf("Failed to connect to websocket server: %v", err)
	}
	defer CloseWithCloseMessage(ws)

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
	defer CloseWithCloseMessage(ws)

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
	defer CloseWithCloseMessage(ws)

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
	defer CloseWithCloseMessage(ws)

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
	defer CloseWithCloseMessage(ws)

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

func TestClientIsAddedToSessionWhenCreatingSession(t *testing.T) {
	testServer, wsUrl := SetupWsServer(t)
	defer testServer.Close()

	// Client 1 connect
	ws, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		t.Fatalf("Failed to connect to websocket server: %v", err)
	}
	defer CloseWithCloseMessage(ws)

	// Received client 1 connect message
	_, msgBytes, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Parse client 1 connect message json
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
	defer CloseWithCloseMessage(ws)

	// Received client 2 connect message
	_, msgBytes, err = ws2.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to connect client 2 to websocket server: %v", err)
	}

	// Parse client 2 connect message json
	var client2ConnectMsg ClientConnectMsg
	err = json.Unmarshal(msgBytes, &client2ConnectMsg)
	if err != nil {
		t.Fatalf("Error parsing client2 message json as ClientConnectMsg: %v", err)
	}

	if client1ConnectMsg.Client.ID == client2ConnectMsg.Client.ID {
		t.Fatalf("Expected clients to each have unique id but both were %s", client2ConnectMsg.Client.ID)
	}

	// Client 1 creates session
	createSessionMsg := CreateSessionMsg{
		Type: "CreateSession",
	}
	err = ws.WriteJSON(createSessionMsg)
	if err != nil {
		t.Fatalf("Failed to send CreateSessionMsg %v", err)
	}

	// Client 1 should receive a message that it was added to the new session
	_, msgBytes, _ = ws.ReadMessage()
	msgString := string(msgBytes)
	log.Printf("msg string %s", msgString)
	var client1AddedToSessionMsg ClientJoinedSessionMsg
	err = json.Unmarshal(msgBytes, &client1AddedToSessionMsg)
	if err != nil {
		t.Fatalf("Error parsing ClientJoinedSessionMsg")
	}
	if client1AddedToSessionMsg.SessionOwnerID != client1ConnectMsg.Client.ID {
		t.Fatalf(
			"Expected session owner client to be first client that connected with id %s",
			client1ConnectMsg.Client.ID,
		)
	}
}

func Test_two_clients_can_join_session(t *testing.T) {
	testServer, wsUrl := SetupWsServer(t)
	defer testServer.Close()

	// Client 1 connect
	ws, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		t.Fatalf("Failed to connect to websocket server: %v", err)
	}
	defer CloseWithCloseMessage(ws)

	// Received client 1 connect message
	_, msgBytes, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Parse client 1 connect message json
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
	defer CloseWithCloseMessage(ws)

	// Received client 2 connect message
	_, msgBytes, err = ws2.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to connect client 2 to websocket server: %v", err)
	}

	// Parse client 2 connect message json
	var client2ConnectMsg ClientConnectMsg
	err = json.Unmarshal(msgBytes, &client2ConnectMsg)
	if err != nil {
		t.Fatalf("Error parsing client2 message json as ClientConnectMsg: %v", err)
	}

	if client1ConnectMsg.Client.ID == client2ConnectMsg.Client.ID {
		t.Fatalf("Expected clients to each have unique id but both were %s", client2ConnectMsg.Client.ID)
	}

	// Client 1 creates session
	createSessionMsg := CreateSessionMsg{
		Type: "CreateSession",
	}
	err = ws.WriteJSON(createSessionMsg)
	if err != nil {
		t.Fatalf("Failed to send CreateSessionMsg %v", err)
	}

	// Client 1 should receive a message that it was added to the new session
	_, msgBytes, _ = ws.ReadMessage()
	msgString := string(msgBytes)
	log.Printf("msg string %s", msgString)
	var client1AddedToSessionMsg ClientJoinedSessionMsg
	err = json.Unmarshal(msgBytes, &client1AddedToSessionMsg)
	if err != nil {
		t.Fatalf("Error parsing ClientJoinedSessionMsg")
	}
	if client1AddedToSessionMsg.SessionOwnerID != client1ConnectMsg.Client.ID {
		t.Fatalf(
			"Expected session owner client to be first client that connected with id %s",
			client1ConnectMsg.Client.ID,
		)
	}

	sessionId := client1AddedToSessionMsg.SessionID

	t.Logf("Client 1 created and added itself to session with id %s", sessionId)

	addClient2ToSessionMsg := AddClientToSessionMsg{
		Type:        "AddClientToSession",
		SessionID:   sessionId,
		AddClientID: client2ConnectMsg.Client.ID,
	}
	err = ws.WriteJSON(addClient2ToSessionMsg)
	if err != nil {
		t.Fatal("Failed to send AddClientToSessionMsg for client 2")
	}

	// Client 2 should receive a message that it was added to the new session
	_, msgBytes, _ = ws2.ReadMessage()
	msgString = string(msgBytes)
	log.Printf("msg string %s", msgString)
	var client2AddedToSessionMsg ClientJoinedSessionMsg
	err = json.Unmarshal(msgBytes, &client2AddedToSessionMsg)
	if err != nil {
		t.Fatalf("Error parsing ClientJoinedSessionMsg")
	}
	if client2AddedToSessionMsg.SessionID != sessionId {
		t.Fatalf("Expected client 2 to be added to session with id %s but was %s", sessionId, client2AddedToSessionMsg.SessionID)
	}
}

// Example payload for session broadcasts
type ExamplePayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func Test_session_owner_can_broadcast_to_another_client(t *testing.T) {
	testServer, wsUrl := SetupWsServer(t)
	defer testServer.Close()

	// Client 1 connect
	ws, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		t.Fatalf("Failed to connect to websocket server: %v", err)
	}
	defer CloseWithCloseMessage(ws)

	// Received client 1 connect message
	_, msgBytes, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Parse client 1 connect message json
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
	defer CloseWithCloseMessage(ws)

	// Received client 2 connect message
	_, msgBytes, err = ws2.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to connect client 2 to websocket server: %v", err)
	}

	// Parse client 2 connect message json
	var client2ConnectMsg ClientConnectMsg
	err = json.Unmarshal(msgBytes, &client2ConnectMsg)
	if err != nil {
		t.Fatalf("Error parsing client2 message json as ClientConnectMsg: %v", err)
	}

	if client1ConnectMsg.Client.ID == client2ConnectMsg.Client.ID {
		t.Fatalf("Expected clients to each have unique id but both were %s", client2ConnectMsg.Client.ID)
	}

	// Client 1 creates session
	createSessionMsg := CreateSessionMsg{
		Type: "CreateSession",
	}
	err = ws.WriteJSON(createSessionMsg)
	if err != nil {
		t.Fatalf("Failed to send CreateSessionMsg %v", err)
	}

	// Client 1 should receive a message that it was added to the new session
	_, msgBytes, _ = ws.ReadMessage()
	msgString := string(msgBytes)
	log.Printf("msg string %s", msgString)
	var client1AddedToSessionMsg ClientJoinedSessionMsg
	err = json.Unmarshal(msgBytes, &client1AddedToSessionMsg)
	if err != nil {
		t.Fatalf("Error parsing ClientJoinedSessionMsg")
	}
	if client1AddedToSessionMsg.SessionOwnerID != client1ConnectMsg.Client.ID {
		t.Fatalf(
			"Expected session owner client to be first client that connected with id %s",
			client1ConnectMsg.Client.ID,
		)
	}

	sessionId := client1AddedToSessionMsg.SessionID

	t.Log("Client 1 created and added itself to session with id ")

	addClient2ToSessionMsg := AddClientToSessionMsg{
		Type:        "AddClientToSession",
		SessionID:   sessionId,
		AddClientID: client2ConnectMsg.Client.ID,
	}
	err = ws.WriteJSON(addClient2ToSessionMsg)
	if err != nil {
		t.Fatal("Failed to send AddClientToSessionMsg for client 2")
	}

	// Client 2 should receive a message that it was added to the new session
	_, msgBytes, _ = ws2.ReadMessage()
	msgString = string(msgBytes)
	log.Printf("msg string %s", msgString)
	var client2AddedToSessionMsg ClientJoinedSessionMsg
	err = json.Unmarshal(msgBytes, &client2AddedToSessionMsg)
	if err != nil {
		t.Fatalf("Error parsing ClientJoinedSessionMsg")
	}
	if client2AddedToSessionMsg.SessionID != sessionId {
		t.Fatalf("Expected client 2 to be added to session with id %s but was %s", sessionId, client2AddedToSessionMsg.SessionID)
	}

	// Client 1 (session owner) broadcast to other clients in session
	payload := ExamplePayload{
		Title: "test title",
		Body:  "test body",
	}
	broadcastToSessionMsg := BroadcastToSessionMsg{
		Type:    "BroadcastToSession",
		Payload: payload,
	}
	err = ws.WriteJSON(broadcastToSessionMsg)
	if err != nil {
		t.Fatalf("Failed to write JSON for broadcast message from client 1 to session %v", err)
	}

	// Expect client 2 to receive broadcast
	_, msgBytes, _ = ws2.ReadMessage()
	msgString = string(msgBytes)
	log.Printf("msg string %s", msgString)
	var broadcastFromSessionMsg BroadcastFromSessionMsg
	err = json.Unmarshal(msgBytes, &broadcastFromSessionMsg)
	if err != nil {
		t.Fatalf("Error parsing BroadcastFromSessionMsg")
	}

	if broadcastFromSessionMsg.Payload != payload {
		t.Fatalf("Expected payload received to be same as one sent received %v", broadcastFromSessionMsg.Payload)
	}

}
