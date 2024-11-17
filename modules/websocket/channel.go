package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"proto-chat/modules/snowflake"
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
	var ownerID uint64 = database.GetServerOwner(channelRequest.ServerID)
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

	type ChannelResponse struct { // this is whats sent to the client when client requests channel
		ChannelID string
		Name      string
	}

	// serialize response about success
	var channelResponse = ChannelResponse{
		ChannelID: strconv.FormatUint(channelID, 10),
		Name:      channelRequest.Name,
	}

	messagesBytes, err := json.Marshal(channelResponse)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	return BroadcastData{
		MessageBytes:    macros.PreparePacket(31, messagesBytes),
		Type:            packetType,
		AffectedServers: []uint64{channelRequest.ServerID},
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

	var isMember bool = database.ConfirmServerMembership(c.userID, serverID)
	if isMember {
		c.setCurrentServerID(serverID)
		var jsonBytes []byte = database.GetChannelList(serverID)
		return macros.PreparePacket(32, jsonBytes)
	} else {
		return macros.RespondFailureReason("Rejected sending channel list of server ID [%d]", serverID)
	}
}
