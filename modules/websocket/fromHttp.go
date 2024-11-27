package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"strconv"
)

func OnProfilePicChanged(userID uint64, fileName string) {
	const jsonType string = "change profile pic"
	// get what servers are the user part of, so message will broadcast to members of these servers
	// this should make sure users who don't have visual on the user who changed profile pic won't get the message
	serverIDsJson, notInAnyServers := database.GetJoinedServersList(userID)
	if notInAnyServers {
		log.Debug("User ID [%d] is not in any servers", userID)
	}

	// deserialize the server ID list
	var serverIDs []uint64
	if err := json.Unmarshal(serverIDsJson, &serverIDs); err != nil {
		log.FatalError(err.Error(), "Error deserializing userServers in onUpdateUserDataRequest for user ID [%d]", userID)
	}

	type ChangedProfilePic struct {
		UserID string
		Pic    string
	}

	var changedProfilePic = ChangedProfilePic{
		UserID: strconv.FormatUint(userID, 10),
		Pic:    fileName,
	}

	jsonBytes, err := json.Marshal(changedProfilePic)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, userID)
	}

	var broadcastData = BroadcastData{
		MessageBytes:    macros.PreparePacket(updateProfilePic, jsonBytes),
		Type:            updateProfilePic,
		AffectedServers: serverIDs,
	}

	broadcastChan <- broadcastData
}
