package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"strconv"
)

func (c *Client) onUpdateUserDataRequest(packetJson []byte, packetType byte) (BroadcastData, []byte) {
	const jsonType string = "change display name"

	// deserialize request
	type UpdateUserDataRequest struct {
		DisplayName string
		Pronouns    string
	}
	var updateUserDataRequest = UpdateUserDataRequest{}

	if err := json.Unmarshal(packetJson, &updateUserDataRequest); err != nil {
		return BroadcastData{}, macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
	}

	// change name in database
	success := database.UpdateUserRow(c.userID, updateUserDataRequest.DisplayName, 0, "display_name")
	if !success {
		return BroadcastData{}, macros.RespondFailureReason("Failed changing display name")
	}

	// serialize response
	type NewDisplayName struct {
		UserID  string
		NewName string
	}
	var newDisplayName = NewDisplayName{
		UserID:  strconv.FormatUint(c.userID, 10),
		NewName: updateUserDataRequest.DisplayName,
	}

	jsonBytes, err := json.Marshal(newDisplayName)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	// get what servers are the user part of, so message will broadcast to members of these servers
	// this should make sure users who don't have visual on the user who changed display name won't get the message
	serverIDsJson, notInAnyServers := database.GetJoinedServersList(c.userID)
	if notInAnyServers {
		log.Debug("User ID [%d] is not in any servers", c.userID)
		return BroadcastData{}, macros.PreparePacket(packetType, jsonBytes)
	}

	// deserialize the server ID list
	var serverIDs []uint64
	if err := json.Unmarshal(serverIDsJson, &serverIDs); err != nil {
		log.FatalError(err.Error(), "Error deserializing userServers in onUpdateUserDataRequest for user ID [%d]", c.userID)
	}

	// prepare broadcast data that will be sent to affected users
	var broadcastData = BroadcastData{
		MessageBytes:    macros.PreparePacket(packetType, jsonBytes),
		Type:            packetType,
		AffectedServers: serverIDs,
	}

	return broadcastData, nil
}

func (c *Client) onUpdateUserStatusValue(packetJson []byte, packetType byte) {
	const jsonType string = "change status value"

	// deserialize request
	type UpdateUserStatusRequest struct {
		Status byte
	}
	var updateUserStatusRequest = UpdateUserStatusRequest{}

	if err := json.Unmarshal(packetJson, &updateUserStatusRequest); err != nil {
		macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
		c.writeChan <- macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
	}
	setUserStatus(c.userID, updateUserStatusRequest.Status)
}
