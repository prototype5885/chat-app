package websocket

import (
	"encoding/json"
	"proto-chat/modules/clients"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"

	// log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
)

func (c *WsClient) onServerMemberListRequest(packetJson []byte) []byte {
	type MemberListRequest struct {
		ServerID uint64
	}

	var req MemberListRequest

	if err := json.Unmarshal(packetJson, &req); err != nil {
		macros.ErrorDeserializing(err.Error(), "member list request", c.UserID)
	}

	members := database.GetServerMembersList(req.ServerID)

	for i := 0; i < len(members); i++ {
		members[i].Online = clients.CheckIfUserIsOnline(c.UserID)
	}

	membersJson, err := json.Marshal(members)
	if err != nil {
		log.FatalError(err.Error(), "Error serializing member list of server ID [%d] for user ID [%d] into json", req.ServerID, c.UserID)
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

// func (c *WsClient) onLeaveServerRequest(packetJson []byte) (BroadcastData, []byte) {
// 	const jsonType string = "server member deletion"

// 	type LeaveServerRequest struct {
// 		ServerID uint64
// 	}

// 	var leaveServerRequest LeaveServerRequest

// 	if err := json.Unmarshal(packetJson, &leaveServerRequest); err != nil {
// 		return BroadcastData{
// 			MessageBytes: macros.ErrorDeserializing(err.Error(), jsonType, c.UserID),
// 		}, nil
// 	}

// 	type ServerMemberDeletionResponse struct {
// 		ServerID uint64
// 		UserID   uint64
// 	}

// 	var serverMemberDeletionResponse = ServerMemberDeletionResponse{
// 		ServerID: leaveServerRequest.ServerID,
// 		UserID:   c.UserID,
// 	}

// 	responseBytes, err := json.Marshal(serverMemberDeletionResponse)
// 	if err != nil {
// 		macros.ErrorSerializing(err.Error(), jsonType, c.UserID)
// 	}

// 	// to make sure client won't receive messages after leaving
// 	success := clients.SetCurrentServerID(c.SessionID, 200)
// 	if !success {
// 		removeWsClient(c.SessionID)
// 	}
// 	success = clients.SetCurrentChannelID(c.SessionID, 0)
// 	if !success {
// 		removeWsClient(c.SessionID)
// 	}

// 	return BroadcastData{
// 		MessageBytes:    macros.PreparePacket(deleteServerMember, responseBytes),
// 		AffectedServers: []uint64{leaveServerRequest.ServerID},
// 		Type:            deleteServerMember,
// 	}, nil
// }
