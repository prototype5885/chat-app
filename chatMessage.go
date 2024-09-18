package main

import (
	"encoding/json"
	"log"
	"proto-chat/modules/snowflake"
	"strconv"
)

// when client sent a chat message
func onChatMessageRequest(jsonBytes []byte, userID uint64, displayName string) []byte {
	type ClientChatMsg struct {
		ChannelID string
		Message   string
	}

	var chatMessageRequest ClientChatMsg

	if err := json.Unmarshal(jsonBytes, &chatMessageRequest); err != nil {
		log.Printf("Error deserializing onChatMessageRequest json of user ID [%d], reason: %s\n", userID, err.Error())
		return nil
	}

	log.Printf("ChannelID: %s, Msg: %s", chatMessageRequest.ChannelID, chatMessageRequest.Message)

	// parse channel id string as uint64
	channelID, parseErr := strconv.ParseUint(chatMessageRequest.ChannelID, 10, 64)
	if parseErr != nil {
		log.Printf("Error parsing uint64 of user ID [%d] in onChatMessageRequest, reason: %s\n", userID, parseErr.Error())
		return nil
	}

	var messageID = snowflake.Generate()

	database.AddChatMessage(messageID, channelID, userID, chatMessageRequest.Message)

	var serverChatMsg = ServerChatMessage{
		MessageID: messageID,
		ChannelID: channelID,
		UserID:    userID,
		Username:  displayName,
		Message:   chatMessageRequest.Message,
	}

	jsonBytes, err := json.Marshal(serverChatMsg)
	if err != nil {
		log.Panic("Error serializing json at onChatMessage, reason:", err)
	}

	return preparePacket(1, jsonBytes)
}

// when client wants to delete a message they own
func onDeleteChatMessageRequest(jsonBytes []byte, userID uint64) []byte {
	type MessageToDelete struct {
		MessageID string
	}

	var messageDeleteRequest = MessageToDelete{}

	if err := json.Unmarshal(jsonBytes, &messageDeleteRequest); err != nil {
		log.Printf("Error deserializing onDeleteChatMessageRequest json of user ID [%d], reason: %s\n", userID, err.Error())
		return nil
	}

	// parse message ID string as uint64
	messageID, parseErr := strconv.ParseUint(messageDeleteRequest.MessageID, 10, 64)
	if parseErr != nil {
		log.Printf("Error parsing uint64 of user ID [%d] in onDeleteChatMessageRequest, reason: %s\n", userID, parseErr.Error())
		return nil
	}

	ownerID, dbSuccess := database.GetChatMessageOwner(messageID)
	if !dbSuccess {
		return nil
	}

	if ownerID != userID {
		log.Printf("User ID [%d] is trying to delete someone else's message [%d], aborting\n", userID, messageID)
		return nil
	}

	database.DeleteChatMessage(messageID)

	messagesBytes, err := json.Marshal(messageDeleteRequest)
	if err != nil {
		log.Panicf("Error serializing json at onDeleteChatMessageRequest for user ID [%d], reason: %s\n:", userID, err.Error())
	}
	return preparePacket(3, messagesBytes)
}

// when client is requesting chat history for a channel
func onChatHistoryRequest(packetJson []byte, userID uint64) []byte {
	type ChatHistoryRequest struct {
		ChannelID string
	}

	var chatHistoryRequest ChatHistoryRequest

	if err := json.Unmarshal(packetJson, &chatHistoryRequest); err != nil {
		log.Printf("Error deserializing onChatHistoryRequest json of user ID [%d], reason: %s\n", userID, err.Error())
		return nil
	}

	// parse channel id string as uint64
	parsedChannelID, parseErr := strconv.ParseUint(chatHistoryRequest.ChannelID, 10, 64)
	if parseErr != nil {
		log.Printf("Error parsing uint64 of user ID [%d] in onChatHistoryRequest, reason: %s\n", userID, parseErr.Error())
		return nil
	}

	type ServerChatMessages struct {
		Messages []ServerChatMessage
	}

	var messages = ServerChatMessages{
		Messages: database.GetMessagesFromChannel(parsedChannelID),
	}

	messagesBytes, err := json.Marshal(messages)
	if err != nil {
		log.Panicf("Error serializing json at onChatHistoryRequest for user ID [%d], reason: %s\n:", userID, err.Error())
	}
	return preparePacket(2, messagesBytes)
}
