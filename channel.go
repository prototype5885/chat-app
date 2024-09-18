package main

import (
	"encoding/json"
	"log"
	"proto-chat/modules/snowflake"
	"strconv"
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
		ServerID string
	}

	var channelRequest = AddChannelRequest{}

	if err := json.Unmarshal(packetJson, &channelRequest); err != nil {
		log.Printf("Error deserializing addChannelRequest json of user ID [%d], reason: %s\n", userID, err.Error())
		return nil
	}

	// parse channel id string as uint64
	parsedServerID, parseErr := strconv.ParseUint(channelRequest.ServerID, 10, 64)
	if parseErr != nil {
		log.Printf("Error parsing uint64 of user ID [%d] in onAddChannelRequest, reason: %s\n", userID, parseErr.Error())
		return nil
	}

	// TODO check if user has permission to add the channel to the server

	var channelID = snowflake.Generate()

	database.AddChannel(channelID, parsedServerID, channelRequest.Name)

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
		ServerID string
	}
	var channelListRequest ChannelListRequest

	if err := json.Unmarshal(packetJson, &channelListRequest); err != nil {
		log.Printf("Error deserializing onC json of user ID [%d], reason: %s\n", userID, err.Error())
		return nil
	}

	// parse channel id string as uint64
	parsedServerID, parseErr := strconv.ParseUint(channelListRequest.ServerID, 10, 64)
	if parseErr != nil {
		log.Printf("Error parsing uint64 of user ID [%d] in onChannelListRequest, reason: %s\n", userID, parseErr.Error())
		return nil
	}

	// TODO check if user has permission to access the server

	type ChannelListResponse struct {
		Channels []ChannelResponse
	}

	var channelListResponse = ChannelListResponse{
		Channels: database.GetChannelList(parsedServerID),
	}

	messagesBytes, err := json.Marshal(channelListResponse)
	if err != nil {
		log.Panicf("Error serializing json at onChannelListRequest for user ID [%d], reason: %s\n:", userID, err.Error())
	}
	return preparePacket(32, messagesBytes)
}
