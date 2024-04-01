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
)

// App Stores the state of our web server
type App struct {
	QRIDCounter int
	Router      *mux.Router
	// TODO: Potentially move to external data store
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
	a.QRIDCounter = 0
	a.Router = mux.NewRouter()
	a.ClientMap = make(map[string]Client)
	a.SessionMap = make(map[string]Session)
	a.Router.HandleFunc("/api/v1/ws", a.serveWs)

	// @TODO Secure with an admin password
	a.Router.HandleFunc("/api/v1/clients", a.getClients)
	a.Router.HandleFunc("/api/v1/sessions", a.getSessions)

	log.Fatal(a.ListenOnPort(4010, false))
}

func (a *App) newClientId() int {
	a.QRIDCounter++
	if a.QRIDCounter > 100000 {
		a.QRIDCounter = 0
	}
	return a.QRIDCounter
}

func (a *App) createClient(r *http.Request, conn *websocket.Conn) Client {
	newClientID := fmt.Sprint(a.newClientId())

	/* Allow client to reconnect with old id */
	rejoinClientID := r.URL.Query().Get("clientId")
	if len(rejoinClientID) > 0 {
		if _, alreadyConnected := a.ClientMap[rejoinClientID]; !alreadyConnected {
			newClientID = rejoinClientID
		}
	}

	fmt.Println("Connection from ", r.RemoteAddr)

	client := Client{
		ID:           newClientID,
		conn:         conn,
		LastJoinTime: time.Now(),
	}
	return client
}

func (a *App) getClients(w http.ResponseWriter, r *http.Request) {
	res, _ := json.MarshalIndent(a.ClientMap, "\n", "  ")
	w.Write(res)
}

func (a *App) getSessions(w http.ResponseWriter, r *http.Request) {
	res, _ := json.MarshalIndent(a.SessionMap, "\n", "  ")
	w.Write(res)
}

func (a *App) getSessionClientMap(sessionID string) map[string]Client {
	if session, ok := a.SessionMap[sessionID]; ok {
		sessionClientMap := make(map[string]Client, len(session.ClientIDs))
		for _, clientID := range session.ClientIDs {
			sessionClientMap[clientID] = a.ClientMap[clientID]
		}
		return sessionClientMap
	}
	return map[string]Client{}
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

	a.removeOldClients()

	client := a.createClient(r, conn)

	a.ClientMap[client.ID] = client
	conn.SetCloseHandler(func(_ int, _ string) error {
		fmt.Println("Connection closed ", r.RemoteAddr)
		fmt.Println("Informing session that client left, id ", client.ID)
		for _, otherClient := range a.getSessionClientMap(client.activeSessionID) {
			if otherClient.ID != client.ID {
				sessionOwnerID := ""
				if session, ok := a.SessionMap[sessionOwnerID]; ok {
					sessionOwnerID = session.OwnerID
					session.ClientIDs = filter(session.ClientIDs, func(ID string) bool {
						return client.ID == ID
					})
					a.SessionMap[session.ID] = session
				}
				clientLeftMsg := ClientLeftSessionMsg{
					Type:           "ClientLeftSession",
					ClientID:       client.ID,
					SessionID:      client.activeSessionID,
					SessionOwnerID: sessionOwnerID,
					ClientMap:      a.getSessionClientMap(client.activeSessionID),
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
			senderClient := a.ClientMap[client.ID]
			msgType := typeJSONValue.String()
			fmt.Println("Message type =", msgType)
			switch msgType {
			case "UpdateClient":
				msg := UpdateClientMsg{}
				json.Unmarshal(message, &msg)
				a.onUpdateClientMsg(senderClient, msg)
			case "CreateSession":
				msg := CreateSessionMsg{}
				json.Unmarshal(message, &msg)
				a.onCreateSessionMsg(client, msg)
			case "AddClientToSession":
				msg := AddClientToSessionMsg{}
				json.Unmarshal(message, &msg)
				a.onAddClientToSessionMsg(senderClient, msg, true)
			case "BroadcastToSession":
				msg := BroadcastToSessionMsg{}
				json.Unmarshal(message, &msg)
				a.onBroadcastToSessionMsg(senderClient, msg)
			}
		}

	}
}

/*
Removes clients that connected over 2 hours ago
*/
func (a *App) removeOldClients() {
	maxClientDuration, err := time.ParseDuration("2h")
	if err != nil {
		expiryTime := time.Now().Add(maxClientDuration)
		for id, client := range a.ClientMap {
			if client.LastJoinTime.After(expiryTime) {
				delete(a.ClientMap, id)
				client.conn.Close()
			}
		}
	}

}

func (a *App) onUpdateClientMsg(senderClient Client, msg UpdateClientMsg) {
	senderClient.Name = msg.Name
	a.ClientMap[senderClient.ID] = senderClient
	for _, client := range a.ClientMap {
		client.conn.WriteJSON(msg)
	}
}

func (a *App) onCreateSessionMsg(senderClient Client, msg CreateSessionMsg) {
	a.QRIDCounter++
	session := Session{
		ID:          fmt.Sprint(a.QRIDCounter),
		OwnerID:     senderClient.ID,
		ClientIDs:   []string{},
		createdDate: time.Now(),
	}
	clientIDsToAdd := []string{senderClient.ID}
	if _, ok := a.ClientMap[msg.AddClientID]; ok {
		clientIDsToAdd = append(clientIDsToAdd, msg.AddClientID)
	}
	a.SessionMap[session.ID] = session
	fmt.Println("Created session", session)
	for _, clientID := range clientIDsToAdd {
		AddClientToSessionMsg := AddClientToSessionMsg{
			Type:        "AddClientToSession",
			SessionID:   session.ID,
			AddClientID: clientID,
		}
		a.onAddClientToSessionMsg(senderClient, AddClientToSessionMsg, false)
	}
}

func (a *App) onAddClientToSessionMsg(senderClient Client, msg AddClientToSessionMsg, replyToSender bool) {
	session, sessionExists := a.SessionMap[msg.SessionID]
	if sessionExists {
		if client, ok := a.ClientMap[msg.AddClientID]; ok {
			session.ClientIDs = append(session.ClientIDs, msg.AddClientID)
			a.SessionMap[session.ID] = session
			client.activeSessionID = session.ID
			a.ClientMap[client.ID] = client
			joinMsg := ClientJoinedSessionMsg{
				Type:           "ClientJoinedSession",
				ClientID:       msg.AddClientID,
				SessionID:      session.ID,
				SessionOwnerID: session.OwnerID,
				ClientMap:      a.getSessionClientMap(session.ID),
			}
			if replyToSender {
				senderClient.conn.WriteJSON(joinMsg)
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

func filter(ss []string, test func(string) bool) (ret []string) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func (a *App) onBroadcastToSessionMsg(senderClient Client, inboundMsg BroadcastToSessionMsg) {
	session, sessionExists := a.SessionMap[senderClient.activeSessionID]
	if sessionExists {
		outboundMsg := BroadcastFromSessionMsg{
			Type:             "BroadcastFromSession",
			FromSessionOwner: session.OwnerID == senderClient.ID,
			SenderID:         senderClient.ID,
			Payload:          inboundMsg.Payload,
		}
		for _, clientID := range session.ClientIDs {
			client := a.ClientMap[clientID]
			client.conn.WriteJSON(outboundMsg)
		}
	}
}
