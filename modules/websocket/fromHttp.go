package websocket

import (
	"encoding/json"
	"proto-chat/modules/clients"
	"proto-chat/modules/database"
	"proto-chat/modules/macros"
)

func OnProfilePicChanged(userID uint64, fileName string) {
	type ChangedProfilePic struct {
		UserID uint64
		Pic    string
	}

	var changedProfilePic = ChangedProfilePic{
		UserID: userID,
		Pic:    fileName,
	}

	jsonBytes, err := json.Marshal(changedProfilePic)
	if err != nil {
		macros.ErrorSerializing(err.Error(), UPDATE_MEMBER_PROFILE_PIC, userID)
	}

	serverIDs := database.GetJoinedServersList(userID)
	if len(serverIDs) > 0 {
		broadcastChan <- BroadcastData{
			MessageBytes:    macros.PreparePacket(UPDATE_MEMBER_PROFILE_PIC, jsonBytes),
			Type:            UPDATE_MEMBER_PROFILE_PIC,
			AffectedServers: serverIDs,
		}
	}

	broadcastChan <- BroadcastData{
		MessageBytes:   macros.PreparePacket(UPDATE_USER_PROFILE_PIC, jsonBytes),
		Type:           UPDATE_USER_PROFILE_PIC,
		AffectedUserID: []uint64{userID},
	}
}

func OnServerPicChanged(serverID uint64, fileName string) {
	type ChangedServerPic struct {
		ServerID uint64
		Pic      string
	}

	changedServerPic := ChangedServerPic{
		ServerID: serverID,
		Pic:      fileName,
	}

	jsonBytes, err := json.Marshal(changedServerPic)
	if err != nil {
		macros.ErrorSerializing(err.Error(), UPDATE_SERVER_PIC, serverID)
	}

	members := database.GetServerMembersList(serverID)
	onlineMembers := clients.FilterOnlineMembers(members)

	broadcastChan <- BroadcastData{
		MessageBytes:   macros.PreparePacket(UPDATE_SERVER_PIC, jsonBytes),
		Type:           UPDATE_SERVER_PIC,
		AffectedUserID: onlineMembers,
	}
}
