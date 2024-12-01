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
		macros.ErrorDeserializing(err.Error(), jsonType, userID)
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

func setUserDisplayName(userID uint64, displayName string) bool {
	if database.UpdateUserRow(userID, displayName, 0, "display_name") {
		return true
	}
	return false
}

func setUserOnline(userID uint64, online bool) {
	const jsonType = "set user online or offline"

	type OnlineStatus struct {
		UserID string
		Online bool
	}

	var onlineStatus = OnlineStatus{
		UserID: strconv.FormatUint(userID, 10),
		Online: online,
	}

	jsonBytes, err := json.Marshal(onlineStatus)
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
		MessageBytes:    macros.PreparePacket(updateOnline, jsonBytes),
		Type:            updateOnline,
		AffectedServers: serverIDs,
	}

	broadcastChan <- broadcastData
}
