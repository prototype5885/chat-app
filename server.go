package main

import (
	"encoding/json"
	"proto-chat/modules/snowflake"
)

type Server struct {
	ServerID uint64
	OwnerID  uint64
	Name     string
	Picture  string
}

type ServerResponse struct {
	ServerID uint64
	Name     string
	Picture  string
}

// when client is requesting to add a new server, type 21
func (c *Client) onAddServerRequest(packetJson []byte) []byte {
	const jsonType string = "add new server"

	type AddServerRequest struct {
		Name string
	}

	var addServerRequest = AddServerRequest{}

	if err := json.Unmarshal(packetJson, &addServerRequest); err != nil {
		return errorDeserializing(err.Error(), jsonType, c.userID)
	}

	var serverID uint64 = snowflake.Generate()
	var picture string = "profilepic2.jpg"

	database.AddServer(serverID, c.userID, addServerRequest.Name, picture)

	var serverForClient = ServerResponse{
		ServerID: serverID,
		Name:     addServerRequest.Name,
		Picture:  picture,
	}

	messagesBytes, err := json.Marshal(serverForClient)
	if err != nil {
		errorSerializing(err.Error(), jsonType, c.userID)
	}
	return preparePacket(21, messagesBytes)
}

// when client requests list of server they are in, type 22
func (c *Client) onServerListRequest() []byte {
	const jsonType string = "server list"

	var servers []ServerResponse = database.GetServerList(c.userID)

	messagesBytes, err := json.Marshal(servers)
	if err != nil {
		errorSerializing(err.Error(), jsonType, c.userID)
	}
	return preparePacket(22, messagesBytes)
}

// when client wants to delete a server, type 23
func (c *Client) onServerDeleteRequest(jsonBytes []byte) []byte {
	const jsonType string = "server deletion"

	type ServerToDelete struct {
		ServerID uint64
	}

	var serverDeleteRequest = ServerToDelete{}

	if err := json.Unmarshal(jsonBytes, &serverDeleteRequest); err != nil {
		return errorDeserializing(err.Error(), jsonType, c.userID)
	}

	success := database.DeleteServer(serverDeleteRequest.ServerID, c.userID)
	if !success {
		return respondFailureReason("Couldn't delete server")
	}

	messagesBytes, err := json.Marshal(serverDeleteRequest)
	if err != nil {
		errorSerializing(err.Error(), jsonType, c.userID)
	}
	return preparePacket(23, messagesBytes)
}
