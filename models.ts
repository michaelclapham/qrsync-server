export namespace ServerTypes {
    export type Msg = ClientConnectMsg | CreateSessionMsg | UpdateClientMsg | AddSessionClientMsg | ClientJoinedSessionMsg | ClientLeftSessionMsg | BroadcastToSessionMsg | BroadcastFromSessionMsg | ErrorMsg | InfoMsg

    export interface Client {
        id: string;
        name: string;
    }
    export interface Session {
        id: string;
        ownerId: string;
        clientIds: string[];
    }
    export interface ClientConnectMsg {
        type: "ClientConnect";
        client: Client;
    }
    export interface CreateSessionMsg {
        type: "CreateSession";
        addClientId: string;
    }
    export interface UpdateClientMsg {
        type: "UpdateClient";
        name: string;
    }
    export interface AddSessionClientMsg {
        type: "AddSessionClient";
        sessionId: string;
        addClientId: string;
    }
    export interface ClientJoinedSessionMsg {
        type: "ClientJoinedSession";
        clientId: string;
        sessionId: string;
        sessionOwnerId: string;
        clientMap: {[key: string]: Client};
    }
    export interface ClientLeftSessionMsg {
        type: "ClientLeftSession";
        clientId: string;
        sessionId: string;
        sessionOwnerId: string;
        clientMap: {[key: string]: Client};
    }
    export interface BroadcastToSessionMsg {
        type: "BroadcastToSession";
        payload: any;
    }
    export interface BroadcastFromSessionMsg {
        type: "BroadcastFromSession";
        fromSessionOwner: boolean;
        senderId: string;
        payload: any;
    }
    export interface ErrorMsg {
        type: "Error";
        message: string;
    }
    export interface InfoMsg {
        type: "Info";
        message: string;
    }
}