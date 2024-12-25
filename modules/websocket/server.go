package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"proto-chat/modules/snowflake"
	"strconv"
)

// when client is requesting to add a new server, type 21
func (c *WsClient) onAddServerRequest(packetJson []byte) []byte {
	const jsonType string = "add new server"

	type AddServerRequest struct {
		Name string
	}

	var addServerRequest = AddServerRequest{}

	if err := json.Unmarshal(packetJson, &addServerRequest); err != nil {
		return macros.ErrorDeserializing(err.Error(), jsonType, c.UserID)
	}

	const defaultPic = ""

	serverID := database.AddNewServer(c.UserID, addServerRequest.Name, defaultPic)

	type ServerResponse struct {
		ServerID uint64
		OwnerID  uint64
		Name     string
		Picture  string
	}

	var serverResponse = ServerResponse{
		ServerID: serverID,
		OwnerID:  c.UserID,
		Name:     addServerRequest.Name,
		Picture:  defaultPic,
	}

	messagesBytes, err := json.Marshal(serverResponse)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.UserID)
	}
	return macros.PreparePacket(21, messagesBytes)
}

// when client requests list of server they are in, type 22
// func (c *Client) onServerListRequest() []byte {
// 	return macros.PreparePacket(22, database.GetServerList(c.userID))
// }

// when client wants to delete a server, type 23
func (c *WsClient) onServerDeleteRequest(jsonBytes []byte, packetType byte) BroadcastData {
	const jsonType string = "server deletion"

	type ServerToDelete struct {
		ServerID uint64
	}

	var serverDeleteRequest = ServerToDelete{}

	if err := json.Unmarshal(jsonBytes, &serverDeleteRequest); err != nil {
		return BroadcastData{
			MessageBytes: macros.ErrorDeserializing(err.Error(), jsonType, c.UserID),
		}
	}

	type ServerDeletionResponse struct {
		ServerID uint64
		UserID   uint64
	}

	var serverDeletionResponse = ServerDeletionResponse{
		ServerID: serverDeleteRequest.ServerID,
		UserID:   c.UserID,
	}

	messagesBytes, err := json.Marshal(serverDeletionResponse)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.UserID)
	}

	return BroadcastData{
		MessageBytes:    macros.PreparePacket(23, messagesBytes),
		Type:            packetType,
		AffectedServers: []uint64{serverDeleteRequest.ServerID},
	}
}

func (c *WsClient) onServerInviteRequest(packetJson []byte) []byte {
	const jsonType string = "server invite"

	type ServerInviteRequest struct {
		ServerID   uint64
		SingleUse  bool
		Expiration uint32
	}

	var serverInviteRequest = ServerInviteRequest{}

	if err := json.Unmarshal(packetJson, &serverInviteRequest); err != nil {
		return macros.ErrorDeserializing(err.Error(), jsonType, c.UserID)
	}

	var inviteID uint64 = snowflake.Generate()

	var serverInvite = database.ServerInvite{
		InviteID:   inviteID,
		ServerID:   serverInviteRequest.ServerID,
		SingleUse:  serverInviteRequest.SingleUse,
		Expiration: uint64(serverInviteRequest.Expiration),
	}

	success := database.Insert(serverInvite)
	if !success {
		log.Fatal("Error creating invite for server ID [%d] for user ID [%d]", serverInviteRequest.ServerID, c.UserID)
	}

	messagesBytes, err := json.Marshal(strconv.FormatUint(inviteID, 10))
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.UserID)
	}
	return macros.PreparePacket(24, messagesBytes)
}
