package websocket

import (
	"database/sql"
	"encoding/json"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"proto-chat/modules/snowflake"
	"proto-chat/modules/structs"
)

type Channel struct {
	ChannelID uint64
	ServerID  uint64
	Name      string
}

type ChannelResponse struct { // this is whats sent to the client when client requests channel
	ChannelID uint64
	Name      string
}

// when client is requesting to add a new channel, type 31
func (c *Client) onAddChannelRequest(packetJson []byte) structs.BroadcastData {
	const jsonType string = "add channel"

	type AddChannelRequest struct {
		Name     string
		ServerID uint64
	}

	var channelRequest = AddChannelRequest{}

	if err := json.Unmarshal(packetJson, &channelRequest); err != nil {
		return structs.BroadcastData{
			MessageBytes: macros.ErrorDeserializing(err.Error(), jsonType, c.userID),
		}
	}

	// TODO check if user has permission to add the channel to the server

	var channel = database.Channel{
		ChannelID: snowflake.Generate(),
		ServerID:  channelRequest.ServerID,
		Name:      channelRequest.Name,
	}

	if !database.Insert(channel) {
		return structs.BroadcastData{
			MessageBytes: macros.RespondFailureReason("Error adding channel"),
		}
	}

	var channelResponse = ChannelResponse{
		ChannelID: channel.ChannelID,
		Name:      channel.Name,
	}

	messagesBytes, err := json.Marshal(channelResponse)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}
	return structs.BroadcastData{
		MessageBytes: macros.PreparePacket(31, messagesBytes),
		ID:           channelRequest.ServerID,
	}
}

// when client requests list of server they are in, type 32
func (c *Client) onChannelListRequest(packetJson []byte) []byte {
	const jsonType string = "channel list"

	type ChannelListRequest struct {
		ServerID uint64
	}
	var channelListRequest ChannelListRequest

	if err := json.Unmarshal(packetJson, &channelListRequest); err != nil {
		macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
	}
	var serverID uint64 = channelListRequest.ServerID

	// TODO check if user has permission to access the server

	var rows *sql.Rows = database.ChannelsTable.GetChannelList(channelListRequest.ServerID)

	var channels []ChannelResponse

	var counter int = 0
	for rows.Next() {
		counter++
		var channel = ChannelResponse{}
		err := rows.Scan(&channel.ChannelID, &channel.Name)
		if err != nil {
			log.FatalError(err.Error(), "Error scanning channel row into struct from server ID [%d]:", serverID)
		}
		channels = append(channels, channel)
	}

	if counter == 0 {
		log.Debug("Server ID [%d] doesn't have any channels", serverID)
	} else {
		log.Debug("Channels from server ID [%d] were retrieved successfully", serverID)
	}

	messagesBytes, err := json.Marshal(channels)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	return macros.PreparePacket(32, messagesBytes)
}
