package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"proto-chat/modules/snowflake"
	"proto-chat/modules/structs"
	"strconv"
)

// type Channel struct {
// 	ChannelID uint64
// 	ServerID  uint64
// 	Name      string
// }

// when client is requesting to add a new channel, type 31
func (c *Client) onAddChannelRequest(packetJson []byte, packetType byte) (BroadcastData, []byte) {
	const jsonType string = "add channel"

	type AddChannelRequest struct {
		Name     string
		ServerID uint64
	}

	var channelRequest = AddChannelRequest{}

	if err := json.Unmarshal(packetJson, &channelRequest); err != nil {
		return BroadcastData{}, macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
	}

	// check if client is authorized to add channel to given server
	var ownerID uint64 = database.ServersTable.GetServerOwner(channelRequest.ServerID)
	if ownerID != c.userID {
		log.Hack("User [%d] is trying to add a channel to server ID [%d] that they dont own", c.userID, channelRequest.ServerID)
		return BroadcastData{}, macros.RespondFailureReason("Error adding channel called [%s]", channelRequest.Name)
	}

	var channelID uint64 = snowflake.Generate()

	// insert into database
	var channel = database.Channel{
		ChannelID: channelID,
		ServerID:  channelRequest.ServerID,
		Name:      channelRequest.Name,
	}

	if !database.Insert(channel) {
		return BroadcastData{}, macros.RespondFailureReason("Error adding channel called [%s]", channelRequest.Name)
	}

	// serialize response about success
	var channelResponse = structs.ChannelResponse{
		ChannelID: strconv.FormatUint(channelID, 10),
		Name:      channelRequest.Name,
	}

	messagesBytes, err := json.Marshal(channelResponse)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	return BroadcastData{
		MessageBytes: macros.PreparePacket(31, messagesBytes),
		Type:         packetType,
		ID:           channelRequest.ServerID,
	}, nil
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

	var channelList []structs.ChannelResponse = database.ChannelsTable.GetChannelList(channelListRequest.ServerID)

	messagesBytes, err := json.Marshal(channelList)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	c.setCurrentServerID(serverID)

	return macros.PreparePacket(32, messagesBytes)
}
