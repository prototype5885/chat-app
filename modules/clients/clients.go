package clients

import (
	"chat-app/modules/database"
	log "chat-app/modules/logging"
	"chat-app/modules/snowflake"
	"sync"
)

type Client struct {
	UserID           uint64
	CurrentChannelID uint64
	CurrentServerID  uint64
	Status           byte
}

// accessed using session id
var Clients sync.Map

func AddClient(userID uint64) uint64 {
	var sessionID uint64 = snowflake.Generate()
	log.Trace("Adding user ID [%d] as session ID [%d] to Clients", userID, sessionID)
	client := &Client{
		UserID:           userID,
		CurrentChannelID: 0,
		CurrentServerID:  2000,
		Status:           1,
	}
	Clients.Store(sessionID, client)

	return sessionID
}

func RemoveClient(sessionID uint64) {
	log.Trace("Removing session ID [%d] from Clients", sessionID)
	Clients.Delete(sessionID)
}

func CheckIfUserIsOnline(userID uint64) bool {
	var online bool = false

	Clients.Range(func(key, value interface{}) bool {
		client, ok := value.(*Client)
		if !ok {
			log.Warn("Invalid WsClient")
			return true
		}
		if client.UserID == userID {
			online = true
			return false
		}
		return true
	})
	if online {
		log.Trace("User ID [%d] is online", userID)
	} else {
		log.Trace("User ID [%d] is offline", userID)
	}

	return online
}

func FilterOnlineMembers(members []database.ServerMember) []uint64 {
	var onlineMembers []uint64

	// only get the online members
	for i := 0; i < len(members); i++ {
		if CheckIfUserIsOnline(members[i].UserID) {
			onlineMembers = append(onlineMembers, members[i].UserID)
		}
	}
	return onlineMembers
}

func GetUserSessions(userID uint64) []uint64 {
	var sessionIDs []uint64

	Clients.Range(func(key, value interface{}) bool {
		sessionID, ok := key.(uint64)
		if !ok {
			log.Warn("Invalid key type")
			return true
		}
		client, ok := value.(*Client)
		if !ok {
			log.Warn("Invalid WsClient type while getting user sessions")
			return true
		}
		if client.UserID == userID {
			sessionIDs = append(sessionIDs, sessionID)
			return false
		}
		return true
	})

	return sessionIDs
}

func GetCurrentChannelID(sessionID uint64) uint64 {
	log.Trace("[Session %d] Getting current channel ID", sessionID)
	client, found := Clients.Load(sessionID)
	if found {
		client, ok := client.(*Client)
		if !ok {
			log.Warn("[Session %d] Invalid Client type while getting current channel ID for session", sessionID)
			return 0
		}
		log.Trace("[Session %d] Current channel is: [%d]", sessionID, client.CurrentChannelID)
		return client.CurrentChannelID
	} else {
		log.Trace("[Session %d] Session was not found while looking for current channel ID", sessionID)
		return 0
	}
}

func SetCurrentChannelID(sessionID uint64, channelID uint64) bool {
	log.Trace("[Session %d] Setting current channel", sessionID)
	client, found := Clients.Load(sessionID)
	if found {
		client, ok := client.(*Client)
		if !ok {
			log.Warn("[Session %d] Invalid Client type while setting current channel ID for session", sessionID)
			return false
		}
		client.CurrentChannelID = channelID
		log.Trace("[Session %d] Current channel set to channel ID [%d]", sessionID, channelID)
		return true
	} else {
		log.Trace("[Session %d] Session was not found while setting current channel ID", sessionID)
		return false
	}
}

func GetCurrentServerID(sessionID uint64) (uint64, bool) {
	log.Trace("[Session %d] Getting current server ID", sessionID)
	client, found := Clients.Load(sessionID)
	if found {
		client, ok := client.(*Client)
		if !ok {
			log.Warn("[Session %d] Invalid Client type while getting current server ID", sessionID)
			return 0, false
		}
		log.Trace("[Session %d] Current server is: [%d]", sessionID, client.CurrentServerID)
		return client.CurrentServerID, true
	} else {
		log.Trace("[Session %d] Session was not found while looking for current server ID", sessionID)
		return 0, false
	}
}

func SetCurrentServerID(sessionID uint64, serverID uint64) bool {
	log.Trace("[Session %d] Setting current server", sessionID)
	client, found := Clients.Load(sessionID)
	if found {
		client, ok := client.(*Client)
		if !ok {
			log.Warn("[Session %d] Invalid Client type while setting current server ID", sessionID)
			return false
		}
		client.CurrentServerID = serverID
		log.Trace("[Session %d] Current server set to server ID [%d]", sessionID, serverID)
		return true
	} else {
		log.Trace("[Session %d] Session was not found while setting current server ID", sessionID)
		return false
	}
}
