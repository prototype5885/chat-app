package clients

import (
	"fmt"
	log "proto-chat/modules/logging"
	"proto-chat/modules/snowflake"
	"sync"
)

type Client struct {
	UserID           uint64
	CurrentChannelID uint64
	CurrentServerID  uint64
	Status           byte
}

// accessed using session id
var Clients = make(map[uint64]*Client)

var mu sync.Mutex

func AddClient(userID uint64) uint64 {
	mu.Lock()
	defer mu.Unlock()

	var sessionID uint64 = snowflake.Generate()
	log.Trace("Adding user ID [%d] as session ID [%d] to Clients", userID, sessionID)
	client := &Client{
		UserID:           userID,
		CurrentChannelID: 0,
		CurrentServerID:  2000,
		Status:           1,
	}
	Clients[sessionID] = client
	return sessionID
}

func RemoveClient(sessionID uint64) {
	mu.Lock()
	defer mu.Unlock()

	log.Trace("Removing session ID [%d] from Clients", sessionID)
	delete(Clients, sessionID)
}

func CheckIfUserIsOnline(userID uint64) bool {
	mu.Lock()
	defer mu.Unlock()

	for _, client := range Clients {
		if client.UserID == userID {
			return true
		}
	}
	return false
}

func GetUserSessions(userID uint64) []uint64 {
	mu.Lock()
	defer mu.Unlock()
	var sessionIDs []uint64
	for key, _ := range Clients {
		sessionIDs = append(sessionIDs, key)
	}

	return sessionIDs
}

func GetCurrentChannelID(sessionID uint64) uint64 {
	mu.Lock()
	defer mu.Unlock()
	if Clients[sessionID] != nil {
		return Clients[sessionID].CurrentChannelID
	} else {
		return 0
	}
}

func SetCurrentChannelID(sessionID uint64, channelID uint64) string {
	mu.Lock()
	defer mu.Unlock()
	if Clients[sessionID] != nil {
		Clients[sessionID].CurrentChannelID = channelID
		log.Trace("User ID [%d] session ID [%d] moved to channel ID [%d]", Clients[sessionID].UserID, sessionID, channelID)
		return ""
	} else {
		return fmt.Sprintf("Couldn't set channel ID for session ID [%d] because session was not found", sessionID)
	}
}

func GetCurrentServerID(sessionID uint64) uint64 {
	mu.Lock()
	defer mu.Unlock()
	if Clients[sessionID] != nil {
		return Clients[sessionID].CurrentServerID
	} else {
		return 0
	}
}

func SetCurrentServerID(sessionID uint64, serverID uint64) string {
	mu.Lock()
	defer mu.Unlock()
	for _, client := range Clients {
		fmt.Println(client)
	}
	if Clients[sessionID] != nil {
		Clients[sessionID].CurrentServerID = serverID
		log.Trace("User ID [%d] session ID [%d] moved to server ID [%d]", Clients[sessionID].UserID, sessionID, Clients[sessionID].CurrentServerID)
		return ""
	} else {
		return fmt.Sprintf("Couldn't set server ID for session ID [%d] because session was not found", sessionID)
	}
}
