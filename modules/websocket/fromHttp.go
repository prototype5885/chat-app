package websocket

import (
	"encoding/json"
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
		macros.ErrorSerializing(err.Error(), "change profile pic", userID)
	}

	serverIDs := database.GetJoinedServersList(userID)
	if len(serverIDs) > 0 {
		broadcastChan <- BroadcastData{
			MessageBytes:    macros.PreparePacket(updateMemberProfilePic, jsonBytes),
			Type:            updateMemberProfilePic,
			AffectedServers: serverIDs,
		}
	}

	broadcastChan <- BroadcastData{
		MessageBytes:   macros.PreparePacket(updateUserProfilePic, jsonBytes),
		Type:           updateUserProfilePic,
		AffectedUserID: userID,
	}
}
