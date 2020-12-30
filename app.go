package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"github.com/twinj/uuid"
)

// App Stores the state of our web server
type App struct {
	IDCounter  int
	Router     *mux.Router
	ClientMap  map[string]Client
	SessionMap map[string]Session
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Init - Initialises app
func (a *App) Init() {
	a.Router = mux.NewRouter()
	a.ClientMap = make(map[string]Client)
	a.SessionMap = make(map[string]Session)
	a.Router.HandleFunc("/api/v1/ws", a.serveWs)

	// @TODO Secure with an admin password
	a.Router.HandleFunc("/api/v1/clients", a.getClients)
	a.Router.HandleFunc("/api/v1/sessions", a.getSessions)

	a.ListenOnPort(4010, false)
}

func (a *App) getClients(w http.ResponseWriter, r *http.Request) {
	res, _ := json.MarshalIndent(a.ClientMap, "\n", "  ")
	w.Write(res)
}

func (a *App) getSessions(w http.ResponseWriter, r *http.Request) {
	res, _ := json.MarshalIndent(a.SessionMap, "\n", "  ")
	w.Write(res)
}

func (a *App) getSessionClients(sessionID string) []Client {
	if session, ok := a.SessionMap[sessionID]; ok {
		clientsSlice := make([]Client, len(session.ClientIDs))
		for i, clientID := range session.ClientIDs {
			clientsSlice[i] = a.ClientMap[clientID]
		}
		return clientsSlice
	}
	return []Client{}
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
		ID:         uuid.NewV4().String(),
		RemoteAddr: r.RemoteAddr,
		conn:       conn,
	}
	a.ClientMap[client.ID] = client
	conn.SetCloseHandler(func(_ int, _ string) error {
		fmt.Println("Connection closed ", r.RemoteAddr)
		fmt.Println("Informing session that client left, id ", client.ID)
		for _, otherClient := range a.getSessionClients(client.activeSessionID) {
			if otherClient.ID != client.ID {
				clientLeftMsg := ClientLeftSessionMsg{
					Type:      "ClientLeftSession",
					ClientID:  client.ID,
					SessionID: client.activeSessionID,
				}
				otherClient.conn.WriteJSON(clientLeftMsg)
			}
		}
		delete(a.ClientMap, client.ID)
		return nil
	})
	connectMsg := ClientConnectMsg{
		Type:   "ClientConnect",
		Client: client,
	}
	conn.WriteJSON(connectMsg)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println("We got a message!")
		fmt.Println(string(message))
		typeJSONValue := gjson.GetBytes(message, "type")
		if !typeJSONValue.Exists() {
			fmt.Println("No message type")
		} else {
			msgType := typeJSONValue.String()
			fmt.Println("Message type =", msgType)
			switch msgType {
			case "UpdateClient":
				msg := UpdateClientMsg{}
				json.Unmarshal(message, &msg)
				a.onUpdateClientMsg(client, msg)
			case "CreateSession":
				a.onCreateSessionMsg(client)
			case "AddSessionClient":
				msg := AddSessionClientMsg{}
				json.Unmarshal(message, &msg)
				a.onAddSessionClientMsg(client, msg)
			case "BroadcastToSession":
				msg := BroadcastToSessionMsg{}
				json.Unmarshal(message, &msg)
				a.onBroadcastToSessionMsg(client, msg)
			}
		}

	}
}

func (a *App) onUpdateClientMsg(senderClient Client, msg UpdateClientMsg) {
	senderClient.Name = msg.Name
}

func (a *App) onCreateSessionMsg(senderClient Client) {
	a.IDCounter = a.IDCounter + 1
	session := Session{
		ID:          fmt.Sprint(a.IDCounter),
		OwnerID:     senderClient.ID,
		ClientIDs:   make([]string, 0, 2),
		createdDate: time.Now(),
	}
	a.SessionMap[session.ID] = session
	fmt.Println("Created session", session)
	senderClient.activeSessionID = session.ID
	addedMsg := ClientJoinedSessionMsg{
		Type:      "ClientJoinedSession",
		ClientID:  senderClient.ID,
		SessionID: session.ID,
	}
	senderClient.conn.WriteJSON(addedMsg)
}

func (a *App) onAddSessionClientMsg(senderClient Client, msg AddSessionClientMsg) {
	session, sessionExists := a.SessionMap[msg.SessionID]
	if sessionExists {
		if client, ok := a.ClientMap[msg.AddClientID]; ok {
			session.ClientIDs = append(session.ClientIDs, msg.AddClientID)
			client.activeSessionID = session.ID
			joinMsg := ClientJoinedSessionMsg{
				Type:           "ClientJoinedSession",
				ClientID:       msg.AddClientID,
				SessionID:      session.ID,
				SessionOwnerID: session.OwnerID,
			}
			client.conn.WriteJSON(joinMsg)
		} else {
			errMsg := ErrorMsg{
				Type:    "error",
				Message: "No client with ID " + msg.AddClientID,
			}
			senderClient.conn.WriteJSON(errMsg)
		}
	} else {
		errMsg := ErrorMsg{
			Type:    "error",
			Message: "No session with ID " + msg.SessionID,
		}
		senderClient.conn.WriteJSON(errMsg)
	}
	fmt.Println("Created session", session)
}

// Map - Apply function to all elements of a slice
func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func (a *App) onBroadcastToSessionMsg(senderClient Client, inboundMsg BroadcastToSessionMsg) {
	session, sessionExists := a.SessionMap[senderClient.activeSessionID]
	if sessionExists {
		outboundMsg := BroadcastFromSessionMsg{
			Type:             "broadcast-from-session",
			FromSessionOwner: session.OwnerID == senderClient.ID,
			SenderID:         senderClient.ID,
			Payload:          inboundMsg.Payload,
		}
		for _, clientID := range session.ClientIDs {
			client := a.ClientMap[clientID]
			if clientID != senderClient.ID {
				client.conn.WriteJSON(outboundMsg)
			}
		}
	}
}
