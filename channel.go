package main

import (
	"encoding/json"
	"log"
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

// when client is requesting to add a new channel
func onAddChannelRequest(packetJson []byte, userID uint64) []byte {
	type AddChannelRequest struct {
		Name     string
		ServerID uint64
	}

	var channelRequest = AddChannelRequest{}

	if err := json.Unmarshal(packetJson, &channelRequest); err != nil {
		log.Printf("Error deserializing addChannelRequest json of user ID [%d], reason: %s\n", userID, err.Error())
		return nil
	}

	// TODO check if user has permission to add the channel to the server

	var channelID = snowflake.Generate()

	if !database.AddChannel(channelID, channelRequest.ServerID, channelRequest.Name) {
		return nil
	}

	var channelResponse = ChannelResponse{
		ChannelID: channelID,
		Name:      channelRequest.Name,
	}

	messagesBytes, err := json.Marshal(channelResponse)
	if err != nil {
		log.Panicf("Error serializing json at onAddChannelRequest for user ID [%d], reason: %s\n:", userID, err.Error())
	}
	return preparePacket(31, messagesBytes)
}

// when client requests list of server they are in
func onChannelListRequest(packetJson []byte, userID uint64) []byte {
	type ChannelListRequest struct {
		ServerID uint64
	}
	var channelListRequest ChannelListRequest

	if err := json.Unmarshal(packetJson, &channelListRequest); err != nil {
		log.Printf("Error deserializing onChannelListRequest json of user ID [%d], reason: %s\n", userID, err.Error())
		return nil
	}

	// TODO check if user has permission to access the server

	type ChannelListResponse struct {
		Channels []ChannelResponse
	}

	var channelListResponse = ChannelListResponse{
		Channels: database.GetChannelList(channelListRequest.ServerID),
	}

	messagesBytes, err := json.Marshal(channelListResponse)
	if err != nil {
		log.Panicf("Error serializing json at onChannelListRequest for user ID [%d], reason: %s\n:", userID, err.Error())
	}
	return preparePacket(32, messagesBytes)
}
