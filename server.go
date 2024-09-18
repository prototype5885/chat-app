package main

import (
	"encoding/json"
	"log"
	"proto-chat/modules/snowflake"
)

type Server struct {
	ServerID uint64
	OwnerID  uint64
	Name     string
	Picture  string
}

type ServerChatMessage struct {
	MessageID uint64
	ChannelID uint64
	UserID    uint64
	Username  string
	Message   string
}

type ServerResponse struct { // this is whats sent to the client when client requests server
	ServerID uint64
	Name     string
	Picture  string
}

// when client is requesting to add a new server
func onAddServerRequest(packetJson []byte, userID uint64) []byte {
	type AddServerRequest struct {
		Name string
	}

	var addServerRequest = AddServerRequest{}

	if err := json.Unmarshal(packetJson, &addServerRequest); err != nil {
		log.Printf("Error deserializing addServerRequest json of user ID [%d], reason: %s\n", userID, err.Error())
		return nil
	}

	var serverID uint64 = snowflake.Generate()
	var picture string = "profilepic2.jpg"

	database.AddServer(serverID, userID, addServerRequest.Name, picture)

	var serverForClient = ServerResponse{
		ServerID: serverID,
		Name:     addServerRequest.Name,
		Picture:  picture,
	}

	messagesBytes, err := json.Marshal(serverForClient)
	if err != nil {
		log.Panicf("Error serializing json at onAddServerRequest for user ID [%d], reason: %s\n:", userID, err.Error())
	}
	return preparePacket(21, messagesBytes)
}

// when client requests list of server they are in
func onServerListRequest(userID uint64) []byte {
	type ServerResponseList struct {
		Servers []ServerResponse
	}

	var serverListRequest = ServerResponseList{
		Servers: database.GetServerList(userID),
	}

	messagesBytes, err := json.Marshal(serverListRequest)
	if err != nil {
		log.Panicf("Error serializing json at onServerListRequest for user ID [%d], reason: %s\n:", userID, err.Error())
	}
	return preparePacket(22, messagesBytes)
}
