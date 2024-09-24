package main

import (
	"encoding/json"
	"proto-chat/modules/snowflake"
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
func (c *Client) onAddChannelRequest(packetJson []byte) BroadcastData {
	const jsonType string = "add channel"

	type AddChannelRequest struct {
		Name     string
		ServerID uint64
	}

	var channelRequest = AddChannelRequest{}

	if err := json.Unmarshal(packetJson, &channelRequest); err != nil {
		return BroadcastData{
			MessageBytes: errorDeserializing(err.Error(), jsonType, c.userID),
		}
	}

	// TODO check if user has permission to add the channel to the server

	var channelID = snowflake.Generate()

	if !database.AddChannel(channelID, channelRequest.ServerID, channelRequest.Name) {
		return BroadcastData{
			MessageBytes: respondFailureReason("Error adding channel"),
		}
	}

	var channelResponse = ChannelResponse{
		ChannelID: channelID,
		Name:      channelRequest.Name,
	}

	messagesBytes, err := json.Marshal(channelResponse)
	if err != nil {
		errorSerializing(err.Error(), jsonType, c.userID)
	}
	return BroadcastData{
		MessageBytes: preparePacket(31, messagesBytes),
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
		errorDeserializing(err.Error(), jsonType, c.userID)
	}

	// TODO check if user has permission to access the server

	var channels []ChannelResponse = database.GetChannelList(channelListRequest.ServerID)

	messagesBytes, err := json.Marshal(channels)
	if err != nil {
		errorSerializing(err.Error(), jsonType, c.userID)
	}
	return preparePacket(32, messagesBytes)
}
