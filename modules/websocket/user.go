package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"
	"proto-chat/modules/macros"
	"strconv"
)

func (c *Client) onUpdateUserDataRequest(packetJson []byte) (BroadcastData, []byte) {
	const jsonType string = "change display name"

	type UpdateUserDataRequest struct {
		DisplayName string
		Pronouns    string
	}

	var updateUserDataRequest UpdateUserDataRequest

	if err := json.Unmarshal(packetJson, &updateUserDataRequest); err != nil {
		return BroadcastData{}, macros.ErrorDeserializing(err.Error(), jsonType, c.UserID)
	}

	// change name in database
	// if !setUserDisplayName(c.UserID, updateUserDataRequest.DisplayName) {
	// 	return BroadcastData{}, macros.RespondFailureReason("Failed changing display name")
	// }

	type NewDisplayName struct {
		UserID  string
		NewName string
	}

	var newDisplayName = NewDisplayName{
		UserID:  strconv.FormatUint(c.UserID, 10),
		NewName: updateUserDataRequest.DisplayName,
	}

	jsonBytes, err := json.Marshal(newDisplayName)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.UserID)
	}

	// get what servers are the user part of, so message will broadcast to members of these servers
	// this should make sure users who don't have visual on the user who changed display name won't get the message
	serverIDs := database.GetJoinedServersList(c.UserID)
	if len(serverIDs) == 0 {
		return BroadcastData{}, macros.PreparePacket(updateUserData, jsonBytes)
	}

	// deserialize the server ID list
	//var serverIDs []uint64
	//if err := json.Unmarshal(serverIDsJson, &serverIDs); err != nil {
	//	macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
	//}

	// prepare broadcast data that will be sent to affected users
	var broadcastData = BroadcastData{
		MessageBytes:    macros.PreparePacket(updateUserData, jsonBytes),
		Type:            updateUserData,
		AffectedServers: serverIDs,
	}

	// workaround so it also sends self if not in any server
	if c.currentServerID == 0 {
		c.WriteChan <- macros.PreparePacket(updateUserData, jsonBytes)
	}

	return broadcastData, nil
}

func (c *Client) onUpdateUserStatusValue(packetJson []byte) {
	const jsonType string = "change status value"

	type UpdateUserStatusRequest struct {
		Status byte
	}

	var updateUserStatusRequest = UpdateUserStatusRequest{}

	if err := json.Unmarshal(packetJson, &updateUserStatusRequest); err != nil {
		macros.ErrorDeserializing(err.Error(), jsonType, c.UserID)
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), jsonType, c.UserID)
	}
	setUserStatus(c.UserID, updateUserStatusRequest.Status)
}
