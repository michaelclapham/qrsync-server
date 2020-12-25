package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fatih/color"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"github.com/twinj/uuid"
)

// App Stores the state of our web server
type App struct {
	Router     *mux.Router
	ClientMap  map[string]Client
	SessionMap map[string]Session
}

// Client - Connected client
type Client struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	conn *websocket.Conn
}

// Session - Session for sharing content
type Session struct {
	ID        string   `json:"id"`
	OwnerID   string   `json:"ownerId"`
	ClientIDs []string `json:"clientIds"`
}

// Message - Websocket message
type Message struct {
	Type string `json:"type"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Init - Initialises app
func (a *App) Init() {
	a.Router = mux.NewRouter()
	a.ClientMap = make(map[string]Client)
	a.SessionMap = make(map[string]Session)
	a.Router.HandleFunc("/ws", a.serveWs)
	a.ListenOnPort(4001, false)
}

// ListenOnPort Starts the app listening on the provided port
func (a *App) ListenOnPort(port int, useSSL bool) error {
	fmt.Println("Starting server on port ", port, " use ssl ", useSSL)
	if useSSL {
		return http.ListenAndServeTLS(fmt.Sprint(":", port), "ssl/server.crt", "ssl/server.key", handlers.CORS()(a.Router))
	}
	return http.ListenAndServe(fmt.Sprint(":", port), handlers.CORS()(a.Router))
}

func (a *App) serveWs(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Connection from ", r.RemoteAddr)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := Client{
		ID:   r.RemoteAddr,
		conn: conn,
	}
	if _, ok := a.ClientMap[client.ID]; ok {
		color.Red("Remote address tried to join twice with same IP and port", r.RemoteAddr)
	}
	conn.SetCloseHandler(func(_ int, _ string) error {
		fmt.Println("Connection closed ", r.RemoteAddr)
		delete(a.ClientMap, client.ID)
		return nil
	})
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println("We got a message!")
		fmt.Println(string(message))
		typeJSONValue := gjson.GetBytes(message, "type")
		msgType := typeJSONValue.String()
		fmt.Println("Message type =", msgType)
		switch msgType {
		case "create-session":
			a.onCreateSessionMsg(client)
		}
	}
}

func (a *App) onCreateSessionMsg(client Client) {
	session := Session{
		ID:        uuid.NewV4().String(),
		OwnerID:   client.ID,
		ClientIDs: make([]string, 0, 2),
	}
	a.SessionMap[session.ID] = session
	fmt.Println("Created session", session)
}
