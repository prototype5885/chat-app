package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"strconv"
)

func setUserStatus(userID uint64, statusValue byte) {
	const jsonType = "change user status value"

	// change status in database
	success := database.UpdateUserRow(userID, "", statusValue, "status")
	if !success {
		log.Warn("Failed to update user status value.")
		return
	}

	// serialize response
	type NewStatus struct {
		UserID string
		Status byte
	}
	var newStatus = NewStatus{
		UserID: strconv.FormatUint(userID, 10),
		Status: statusValue,
	}

	jsonBytes, err := json.Marshal(newStatus)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, userID)
	}

	// get what servers are the user part of, so message will broadcast to members of these servers
	// this should make sure users who don't have visual on the user who changed user status won't get the message
	serverIDsJson, notInAnyServers := database.GetJoinedServersList(userID)
	if notInAnyServers {
		log.Debug("User ID [%d] is not in any servers", userID)
		return
	}

	// deserialize the server ID list
	var serverIDs []uint64
	if err := json.Unmarshal(serverIDsJson, &serverIDs); err != nil {
		log.FatalError(err.Error(), "Error deserializing userServers in onUpdateUserStatusValue for user ID [%d]", userID)
	}

	// prepare broadcast data that will be sent to affected users
	var broadcastData = BroadcastData{
		MessageBytes:    macros.PreparePacket(updateStatus, jsonBytes),
		Type:            updateStatus,
		AffectedServers: serverIDs,
	}

	broadcastChan <- broadcastData
}

func setUserStatusText(userID uint64, statusText string) {
	const jsonType = "change user status text"

	// change status in database
	success := database.UpdateUserRow(userID, statusText, 0, "status_text")
	if !success {
		log.Warn("Failed to update user status text.")
		return
	}

	// serialize response
	type NewStatusText struct {
		UserID     string
		StatusText string
	}
	var newStatusText = NewStatusText{
		UserID:     strconv.FormatUint(userID, 10),
		StatusText: statusText,
	}

	jsonBytes, err := json.Marshal(newStatusText)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, userID)
	}

	// get what servers are the user part of, so message will broadcast to members of these servers
	// this should make sure users who don't have visual on the user who changed user status text won't get the message
	serverIDsJson, notInAnyServers := database.GetJoinedServersList(userID)
	if notInAnyServers {
		log.Debug("User ID [%d] is not in any servers", userID)
		return
	}

	// deserialize the server ID list
	var serverIDs []uint64
	if err := json.Unmarshal(serverIDsJson, &serverIDs); err != nil {
		log.FatalError(err.Error(), "Error deserializing userServers in onUpdateUserStatusValue for user ID [%d]", userID)
	}

	// prepare broadcast data that will be sent to affected users
	var broadcastData = BroadcastData{
		MessageBytes:    macros.PreparePacket(updateStatusText, jsonBytes),
		Type:            updateStatusText,
		AffectedServers: serverIDs,
	}

	broadcastChan <- broadcastData
}
