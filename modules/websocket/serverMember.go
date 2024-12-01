package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
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

	var memberInfos = database.GetServerMembersList(userListRequest.ServerID, c.userID)

	//responseBytes := database.GetServerMembersList(userListRequest.ServerID, c.userID)

	for i := range memberInfos {
		// parse user ID string into uint64
		foundUserID, err := strconv.ParseUint(memberInfos[i].UserID, 10, 64)
		if err != nil {
			log.FatalError(err.Error(), "Error parsing user ID string [%s] as uint64", memberInfos[i].UserID)
		}

		// check if this user is online currently
		memberInfos[i].Online = checkIfUserIsOnline(foundUserID)
	}

	responseBytes, err := json.Marshal(memberInfos)
	if err != nil {
		macros.ErrorSerializing(err.Error(), "server member list", c.userID)
	}

	return macros.PreparePacket(42, responseBytes)
}

//func (c *Client) onMemberOnlineStatusesRequest(packetJson []byte) []byte {
//	const jsonType string = "member statuses"
//	type OnlineStatusRequest struct {
//		ServerID uint64
//	}
//
//	var onlineStatusRequest OnlineStatusRequest
//
//	if err := json.Unmarshal(packetJson, &onlineStatusRequest); err != nil {
//		macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
//	}
//
//	var onlineMembers = make([]uint64, 0, len(Clients))
//
//	for i, client := range Clients {
//		onlineMembers[i] = client.userID
//	}
//
//	responseBytes, err := json.Marshal(onlineMembers)
//	if err != nil {
//		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
//	}
//
//	return responseBytes
//}

func (c *Client) onLeaveServerRequest(packetJson []byte) (BroadcastData, []byte) {
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

	type ServerMemberDeletionResponse struct {
		ServerID string
		UserID   string
	}

	var serverMemberDeletionResponse = ServerMemberDeletionResponse{
		ServerID: strconv.FormatUint(leaveServerRequest.ServerID, 10),
		UserID:   strconv.FormatUint(c.userID, 10),
	}

	responseBytes, err := json.Marshal(serverMemberDeletionResponse)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	// to make sure client won't receive messages after leaving
	c.currentServerID = 200
	c.currentChannelID = 0

	return BroadcastData{
		MessageBytes:    macros.PreparePacket(deleteServerMember, responseBytes),
		AffectedServers: []uint64{leaveServerRequest.ServerID},
		Type:            deleteServerMember,
	}, nil
}
