package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"proto-chat/modules/snowflake"
	"proto-chat/modules/structs"
	"strconv"
)

// when client is requesting to add a new server, type 21
func (c *Client) onAddServerRequest(packetJson []byte) []byte {
	const jsonType string = "add new server"

	type AddServerRequest struct {
		Name string
	}

	var addServerRequest = AddServerRequest{}

	if err := json.Unmarshal(packetJson, &addServerRequest); err != nil {
		return macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
	}

	var server database.Server = database.AddNewServer(c.userID, addServerRequest.Name, "default_serverpic.webp")

	// this was made because javascript client doesn't like serialized
	var serverResponse = structs.ServerResponse{
		ServerID: strconv.FormatUint(server.ServerID, 10),
		OwnerID:  strconv.FormatUint(server.OwnerID, 10),
		Name:     server.Name,
		Picture:  server.Picture,
	}

	messagesBytes, err := json.Marshal(serverResponse)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}
	return macros.PreparePacket(21, messagesBytes)
}

// when client requests list of server they are in, type 22
func (c *Client) onServerListRequest() []byte {
	const jsonType string = "server list"

	var servers []structs.ServerResponse = database.GetServerList(c.userID)

	messagesBytes, err := json.Marshal(servers)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}
	return macros.PreparePacket(22, messagesBytes)
}

// when client wants to delete a server, type 23
func (c *Client) onServerDeleteRequest(jsonBytes []byte, packetType byte) BroadcastData {
	const jsonType string = "server deletion"

	type ServerToDelete struct {
		ServerID uint64
	}

	var serverDeleteRequest = ServerToDelete{}

	if err := json.Unmarshal(jsonBytes, &serverDeleteRequest); err != nil {
		return BroadcastData{
			MessageBytes: macros.ErrorDeserializing(err.Error(), jsonType, c.userID),
		}
	}

	var serverToDelete = database.ServerDeletion{
		ServerID: serverDeleteRequest.ServerID,
		UserID:   c.userID,
	}

	success := database.Delete(serverToDelete)
	if !success {
		return BroadcastData{
			MessageBytes: macros.RespondFailureReason("Couldn't delete server"),
		}
	}

	var serverDeletionResponse = structs.ServerDeletionResponse{
		ServerID: strconv.FormatUint(serverDeleteRequest.ServerID, 10),
		UserID:   strconv.FormatUint(c.userID, 10),
	}

	messagesBytes, err := json.Marshal(serverDeletionResponse)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	return BroadcastData{
		MessageBytes: macros.PreparePacket(23, messagesBytes),
		Type:         packetType,
		ID:           serverDeleteRequest.ServerID,
	}
}

func (c *Client) onServerInviteRequest(packetJson []byte) []byte {
	const jsonType string = "server invite"

	type ServerInviteRequest struct {
		ServerID   uint64
		SingleUse  bool
		Expiration uint32
	}

	var serverInviteRequest = ServerInviteRequest{}

	if err := json.Unmarshal(packetJson, &serverInviteRequest); err != nil {
		return macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
	}

	var inviteID uint64 = snowflake.Generate()

	var serverInvite = database.ServerInvite{
		InviteID:   inviteID,
		ServerID:   serverInviteRequest.ServerID,
		SingleUse:  serverInviteRequest.SingleUse,
		Expiration: uint64(serverInviteRequest.Expiration),
	}

	if !database.Insert(serverInvite) {
		log.Fatal("Error creating invite for server ID [%d] for user ID [%d]", serverInviteRequest.ServerID, c.userID)
	}

	messagesBytes, err := json.Marshal(strconv.FormatUint(inviteID, 10))
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}
	return macros.PreparePacket(24, messagesBytes)
}
