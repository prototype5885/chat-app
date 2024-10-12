package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"strconv"
)

func (c *Client) onChangeDisplayNameRequest(packetJson []byte, packetType byte) (BroadcastData, []byte) {
	const jsonType string = "change display name"

	// deserialize request
	type ChangeDisplayNameRequest struct {
		NewName string
	}
	var changeDisplayNameRequest = ChangeDisplayNameRequest{}

	if err := json.Unmarshal(packetJson, &changeDisplayNameRequest); err != nil {
		return BroadcastData{}, macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
	}

	// change name in database
	success := database.ChangeDisplayName(c.userID, changeDisplayNameRequest.NewName)
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
		NewName: changeDisplayNameRequest.NewName,
	}

	jsonBytes, err := json.Marshal(newDisplayName)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	// get what servers are the user part of, so message will be broadcasted to members of these servers
	// this should make sure users who don't have visual on the user who changed display name won't get the message
	serverIDsJson, notInAnyServers := database.GetJoinedServersList(c.userID)
	if notInAnyServers {
		log.Debug("User ID [%d] is not in any servers", c.userID)
		return BroadcastData{}, macros.PreparePacket(packetType, jsonBytes)
	}

	// deserialize the server ID list
	var serverIDs []uint64
	if err := json.Unmarshal(serverIDsJson, &serverIDs); err != nil {
		log.FatalError(err.Error(), "Error deserializing userServers in onChangeDisplayNameRequest for user ID [%d]", c.userID)
	}

	// prepare broadcast data that will be sent to affected users
	var broadcastData = BroadcastData{
		MessageBytes:    macros.PreparePacket(packetType, jsonBytes),
		Type:            packetType,
		AffectedServers: serverIDs,
	}

	return broadcastData, nil
}
