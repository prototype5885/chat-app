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
	type MemberListRequest struct {
		ServerID uint64
	}

	var memberListRequest MemberListRequest

	if err := json.Unmarshal(packetJson, &memberListRequest); err != nil {
		macros.ErrorDeserializing(err.Error(), "member list request", c.UserID)
	}

	members := database.GetServerMembersList(memberListRequest.ServerID)

	for i := 0; i < len(members); i++ {
		members[i].Online = c.CheckIfUserIsOnline()
	}

	membersJson, err := json.Marshal(members)
	if err != nil {
		log.FatalError(err.Error(), "Error serializing member list of server ID [%d] for user ID [%d] into json", memberListRequest.ServerID, c.UserID)
	}

	return macros.PreparePacket(42, membersJson)
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
			MessageBytes: macros.ErrorDeserializing(err.Error(), jsonType, c.UserID),
		}, nil
	}

	type ServerMemberDeletionResponse struct {
		ServerID string
		UserID   string
	}

	var serverMemberDeletionResponse = ServerMemberDeletionResponse{
		ServerID: strconv.FormatUint(leaveServerRequest.ServerID, 10),
		UserID:   strconv.FormatUint(c.UserID, 10),
	}

	responseBytes, err := json.Marshal(serverMemberDeletionResponse)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.UserID)
	}

	// to make sure client won't receive messages after leaving
	c.currentServerID = 200
	c.CurrentChannelID = 0

	return BroadcastData{
		MessageBytes:    macros.PreparePacket(deleteServerMember, responseBytes),
		AffectedServers: []uint64{leaveServerRequest.ServerID},
		Type:            deleteServerMember,
	}, nil
}
