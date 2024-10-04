package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"proto-chat/modules/snowflake"
)

// when client is requesting to add a new server, type 21
func (c *Client) onAddServerRequest(packetJson []byte) []byte {
	const jsonType string = "add new server"

	type AddServerRequest struct {
		Name string
	}

	var addServerRequest = AddServerRequest{}

	if err := json.Unmarshal(packetJson, &addServerRequest); err != nil {
		return macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
	}

	var serverID uint64 = snowflake.Generate()
	var picture string = "profilepic2.webp"

	var server = database.Server{
		ServerID: serverID,
		OwnerID:  c.userID,
		Name:     addServerRequest.Name,
		Picture:  picture,
	}

	if !database.Insert(server) {
		log.Fatal("Error adding server ID [%d] on the request of user ID [%d]", serverID, c.userID)
	}

	var channelID uint64 = snowflake.Generate()

	var channel = database.Channel{
		ChannelID: channelID,
		ServerID:  serverID,
		Name:      "Default Channel",
	}

	if !database.Insert(channel) {
		log.Fatal("Error adding channel ID [%d] to the newly created server ID [%d]", channelID, serverID)
	}

	messagesBytes, err := json.Marshal(server)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}
	return macros.PreparePacket(21, messagesBytes)
}

// when client requests list of server they are in, type 22
func (c *Client) onServerListRequest() []byte {
	const jsonType string = "server list"

	var servers []database.Server = database.ServersTable.GetServerList(c.userID)

	messagesBytes, err := json.Marshal(servers)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}
	return macros.PreparePacket(22, messagesBytes)
}

// when client wants to delete a server, type 23
func (c *Client) onServerDeleteRequest(jsonBytes []byte, packetType byte) BroadcastData {
	const jsonType string = "server deletion"

	type ServerToDelete struct {
		ServerID uint64
	}

	var serverDeleteRequest = ServerToDelete{}

	if err := json.Unmarshal(jsonBytes, &serverDeleteRequest); err != nil {
		return BroadcastData{
			MessageBytes: macros.ErrorDeserializing(err.Error(), jsonType, c.userID),
		}
	}

	var serverToDelete = database.ServerDeletion{
		ServerID: serverDeleteRequest.ServerID,
		UserID:   c.userID,
	}

	success := database.Delete(serverToDelete)
	if success == 0 {
		return BroadcastData{
			MessageBytes: macros.RespondFailureReason("Couldn't delete server"),
		}
	}

	return BroadcastData{
		MessageBytes: macros.PreparePacket(23, jsonBytes),
		Type:         packetType,
		ID:           serverDeleteRequest.ServerID,
	}
}
