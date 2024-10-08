package main

import (
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tkrajina/typescriptify-golang-structs/typescriptify"
)

func convertToTS() {
	converter := typescriptify.New().
		Add(Client{}).
		Add(Session{}).
		Add(ClientConnectMsg{}).
		Add(CreateSessionMsg{}).
		Add(UpdateClientMsg{}).
		Add(AddClientToSessionMsg{}).
		Add(ClientJoinedSessionMsg{}).
		Add(ClientLeftSessionMsg{}).
		Add(BroadcastToSessionMsg{}).
		Add(BroadcastFromSessionMsg{}).
		Add(ErrorMsg{}).
		Add(InfoMsg{})

	converter.CreateInterface = true
	converter.ManageType(time.Time{}, typescriptify.TypeOptions{
		TSType: "string",
	})
	converter.BackupDir = ""
	tsString, err := converter.Convert(make(map[string]string))
	if err != nil {
		panic(err.Error())
	}
	regex, _ := regexp.Compile("([A-z]+)Msg {\n    type: string;")
	msgTypeNames := make([]string, 1)
	modifiedTs := regex.ReplaceAllStringFunc(tsString, func(str string) string {
		msgTypeName := regex.FindStringSubmatch(str)[1]
		msgTypeNames = append(msgTypeNames, msgTypeName+"Msg")
		return strings.Replace(str, "type: string;", "type: \""+msgTypeName+"\";", -1)
	})
	modifiedTs = strings.ReplaceAll(modifiedTs, "\n", "\n    ")
	unionTypeStr := "    export type Msg = " + msgTypeNames[0] + strings.Join(msgTypeNames[1:], " | ")
	modifiedTs = "export namespace ServerTypes {\n" + unionTypeStr + "\n" + modifiedTs + "\n}"
	ioutil.WriteFile("./models.ts", []byte(modifiedTs), 0644)

}

// Client - Connected client
type Client struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	conn            *websocket.Conn
	activeSessionID string
	LastJoinTime    time.Time `json:"lastJoinTime"`
}

// Session - Session for sharing content
type Session struct {
	ID          string   `json:"id"`
	OwnerID     string   `json:"ownerId"`
	ClientIDs   []string `json:"clientIds"`
	createdDate time.Time
}

// CreateSessionMsg - Sent from client to create session
type CreateSessionMsg struct {
	Type string `json:"type"`
}

// ClientConnectMsg - Sent to client on connecting
type ClientConnectMsg struct {
	Type   string `json:"type"`
	Client Client `json:"client"`
}

// UpdateClientMsg - Updates a client
type UpdateClientMsg struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// AddClientToSessionMsg - Websocket message
type AddClientToSessionMsg struct {
	Type        string `json:"type"`
	SessionID   string `json:"sessionId"`
	AddClientID string `json:"addClientId"`
}

// ClientJoinedSessionMsg -
type ClientJoinedSessionMsg struct {
	Type           string            `json:"type"`
	ClientID       string            `json:"clientId"`
	SessionID      string            `json:"sessionId"`
	SessionOwnerID string            `json:"sessionOwnerId"`
	ClientMap      map[string]Client `json:"clientMap"`
}

// ClientLeftSessionMsg -
type ClientLeftSessionMsg struct {
	Type           string            `json:"type"`
	ClientID       string            `json:"clientId"`
	SessionID      string            `json:"sessionId"`
	SessionOwnerID string            `json:"sessionOwnerId"`
	ClientMap      map[string]Client `json:"clientMap"`
}

// BroadcastToSessionMsg - Used by client to send content to all clients in session
type BroadcastToSessionMsg struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// BroadcastFromSessionMsg - Used by server to send content all clients in a session
type BroadcastFromSessionMsg struct {
	Type             string `json:"type"`
	FromSessionOwner bool   `json:"fromSessionOwner"`
	SenderID         string `json:"senderId"`
	Payload          string `json:"payload"`
}

// ErrorMsg - Websocket error message
type ErrorMsg struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// InfoMsg - Websocket info message
type InfoMsg struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
