package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"
	"proto-chat/modules/structs"
	"strconv"

	// log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
)

func (c *Client) onServerMemberListRequest(packetJson []byte) []byte {
	const jsonType string = "member list"

	type UserListRequest struct {
		ServerID uint64
	}
	var userListRequest UserListRequest

	if err := json.Unmarshal(packetJson, &userListRequest); err != nil {
		macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
	}
	var serverID uint64 = userListRequest.ServerID

	var userIDs []string = database.GetServerMembersList(serverID)

	jsonBytes, err := json.Marshal(userIDs)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	return macros.PreparePacket(42, jsonBytes)
}

func (c *Client) onServerMemberDeleteRequest(packetJson []byte, packetType byte) (BroadcastData, []byte) {
	const jsonType string = "server member deletion"

	type LeaveServerRequest struct {
		ServerID uint64
	}

	var leaveServerRequest LeaveServerRequest

	if err := json.Unmarshal(packetJson, &leaveServerRequest); err != nil {
		return BroadcastData{
			MessageBytes: macros.ErrorDeserializing(err.Error(), jsonType, c.userID),
		}, nil
	}

	var serverMember = database.ServerMember{
		ServerID: leaveServerRequest.ServerID,
		UserID:   c.userID,
	}

	if !database.Delete(serverMember) {
		return BroadcastData{}, macros.RespondFailureReason("Couldn't leave server")
	}

	var serverMemberDeletionResponse = structs.ServerMemberDeletionResponse{
		ServerID: strconv.FormatUint(leaveServerRequest.ServerID, 10),
		UserID:   strconv.FormatUint(c.userID, 10),
	}

	responseBytes, err := json.Marshal(serverMemberDeletionResponse)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	return BroadcastData{
		MessageBytes: macros.PreparePacket(packetType, responseBytes),
		ID:           leaveServerRequest.ServerID,
		Type:         packetType,
	}, nil
}
