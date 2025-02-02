package websocket

import (
	"chat-app/modules/clients"
	"chat-app/modules/database"
	log "chat-app/modules/logging"
	"chat-app/modules/macros"
	"encoding/json"
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

func OnServerBannerChanged(serverID uint64, fileName string) {
	type ChangedServerPic struct {
		ServerID uint64
		Banner   string
	}

	changedServerPic := ChangedServerPic{
		ServerID: serverID,
		Banner:   fileName,
	}

	jsonBytes, err := json.Marshal(changedServerPic)
	if err != nil {
		macros.ErrorSerializing(err.Error(), UPDATE_SERVER_BANNER, serverID)
	}

	members := database.GetServerMembersList(serverID)
	onlineMembers := clients.FilterOnlineMembers(members)

	broadcastChan <- BroadcastData{
		MessageBytes:   macros.PreparePacket(UPDATE_SERVER_BANNER, jsonBytes),
		Type:           UPDATE_SERVER_BANNER,
		AffectedUserID: onlineMembers,
	}
}

func OnUserJoinedServer(userID uint64, serverID uint64) {
	type UserJoinedServer struct {
		ServerID uint64
		Data     database.ServerMember
	}

	userJoinedServer := UserJoinedServer{
		ServerID: serverID,
		Data:     database.GetUserData(userID),
	}

	if userJoinedServer.Data.UserID == 0 {
		log.Error("User ID [%d] was not found in database in OnUserJoinedServer, user instantly left as soon as joined?", userID)
		return
	}

	userJoinedServer.Data.Online = clients.CheckIfUserIsOnline(userJoinedServer.Data.UserID)

	jsonBytes, err := json.Marshal(userJoinedServer)
	if err != nil {
		macros.ErrorSerializing(err.Error(), ADD_SERVER_MEMBER, serverID)
		return
	}

	// broadcast to users who are in that server
	broadcastChan <- BroadcastData{
		MessageBytes:    macros.PreparePacket(ADD_SERVER_MEMBER, jsonBytes),
		Type:            ADD_SERVER_MEMBER,
		AffectedServers: []uint64{serverID},
	}

	// now broadcast to every session of the user who joined a server
	serverData := database.GetServerData(serverID)

	var dataOfServer = database.JoinedServer{
		ServerID: serverID,
		Name:     serverData.Name,
		Picture:  serverData.Picture,
		Banner:   serverData.Banner,
	}

	if userID == serverData.UserID {
		dataOfServer.Owned = true
	} else {
		dataOfServer.Owned = false
	}

	jsonBytes2, err := json.Marshal(dataOfServer)
	if err != nil {
		macros.ErrorSerializing(err.Error(), ADD_SERVER, serverID)
		return
	}

	broadcastChan <- BroadcastData{
		MessageBytes:   macros.PreparePacket(ADD_SERVER, jsonBytes2),
		Type:           ADD_SERVER,
		AffectedUserID: []uint64{userID},
	}
}
