package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"

	// log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
)

func (c *Client) onMemberListRequest(packetJson []byte) []byte {
	const jsonType string = "member list"

	type UserListRequest struct {
		ServerID uint64
	}
	var userListRequest UserListRequest

	if err := json.Unmarshal(packetJson, &userListRequest); err != nil {
		macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
	}
	var serverID uint64 = userListRequest.ServerID

	var userIDs []uint64 = database.ServerMembersTable.GetServerMembersList(serverID)

	jsonBytes, err := json.Marshal(userIDs)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	return macros.PreparePacket(42, jsonBytes)
}
