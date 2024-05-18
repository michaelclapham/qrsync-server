package main

import (
	"fmt"
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
}

func TestServerStarts2(t *testing.T) {
	app := App{}
	app.Init()
	testServer := httptest.NewServer(app.MainHandler())
	defer testServer.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(testServer.URL, "http")

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// Send message to server, read response and check to see if it's what we expect.
	for i := 0; i < 10; i++ {
		if err := ws.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
			t.Fatalf("%v", err)
		}
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(p) != "hello" {
			t.Fatalf("bad message")
		}
	}
}
