package websocket

import (
	"encoding/json"
	"fmt"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"proto-chat/modules/snowflake"
	"strconv"
)

// when client is requesting to add a new channel, type 31
func (c *Client) onAddChannelRequest(packetJson []byte, packetType byte) (BroadcastData, []byte) {
	const jsonType string = "add channel"

	type AddChannelRequest struct {
		Name     string
		ServerID uint64
	}

	var channelRequest = AddChannelRequest{}

	if err := json.Unmarshal(packetJson, &channelRequest); err != nil {
		return BroadcastData{}, macros.ErrorDeserializing(err.Error(), jsonType, c.UserID)
	}

	var errorMessage = fmt.Sprintf("Error adding channel called [%s]", channelRequest.Name)

	// check if client is authorized to add channel to given server
	var ownerID uint64 = database.GetServerOwner(channelRequest.ServerID)
	if ownerID != c.UserID {
		log.Hack("User [%d] is trying to add a channel to server ID [%d] that they dont own", c.UserID, channelRequest.ServerID)
		return BroadcastData{}, macros.RespondFailureReason(errorMessage)
	}

	var channelID uint64 = snowflake.Generate()

	// insert into database
	var channel = database.Channel{
		ChannelID: channelID,
		ServerID:  channelRequest.ServerID,
		Name:      channelRequest.Name,
	}

	success := database.Insert(channel)
	if !success {
		return BroadcastData{}, macros.RespondFailureReason(errorMessage)
	}

	type ChannelResponse struct { // this is what's sent to the client when client requests channel
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
		macros.ErrorSerializing(err.Error(), jsonType, c.UserID)
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
		macros.ErrorDeserializing(err.Error(), jsonType, c.UserID)
	}

	var serverID uint64 = channelListRequest.ServerID

	var isMember bool = database.ConfirmServerMembership(c.UserID, serverID)
	if isMember {
		c.setCurrentServerID(serverID)
		var jsonBytes []byte = database.GetChannelList(serverID)
		return macros.PreparePacket(32, jsonBytes)
	} else {
		return macros.RespondFailureReason("Rejected sending channel list of server ID [%d]", serverID)
	}
}
