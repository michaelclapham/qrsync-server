/* Do not change, this code is generated from Golang structs */


export interface Client {
    id: string;
    remoteAddr: string;
    name: string;
}
export interface Session {
    id: string;
    ownerId: string;
    clientIds: string[];
}
export interface ClientConnectMsg {
    type: string;
    client: Client;
}
export interface UpdateClientMsg {
    type: string;
    name: string;
}
export interface AddSessionClientMsg {
    type: string;
    sessionId: string;
    addClientId: string;
}
export interface AddedToSessionMsg {
    type: string;
    sessionId: string;
}
export interface BroadcastToSessionMsg {
    type: string;
    payload: any;
}
export interface BroadcastFromSessionMsg {
    type: string;
    fromSessionOwner: boolean;
    senderId: string;
    payload: any;
}
export interface ErrorMsg {
    type: string;
    message: string;
}
export interface InfoMsg {
    type: string;
    message: string;
}