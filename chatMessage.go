package main

import (
	"encoding/json"
	"fmt"
	"log"
	"proto-chat/modules/snowflake"
)

type ChatMessageResponse struct {
	IDm uint64 // message ID
	IDu uint64 // user ID
	Msg string // message
}

// when client sent a chat message
func onChatMessageRequest(jsonBytes []byte, userID uint64) []byte {
	type ClientChatMsg struct {
		ChannelID uint64
		Message   string
	}

	var chatMessageRequest ClientChatMsg

	if err := json.Unmarshal(jsonBytes, &chatMessageRequest); err != nil {
		log.Printf("Error deserializing onChatMessageRequest json of user ID [%d], reason: %s\n", userID, err.Error())
		return nil
	}

	log.Printf("ChannelID: %d, Msg: %s", chatMessageRequest.ChannelID, chatMessageRequest.Message)

	var messageID = snowflake.Generate()

	problem := database.AddChatMessage(messageID, chatMessageRequest.ChannelID, userID, chatMessageRequest.Message)
	if problem != "" {
		return setProblem(problem)
	}

	var serverChatMsg = ChatMessageResponse{
		IDm: messageID,
		IDu: userID,
		Msg: chatMessageRequest.Message,
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
		MessageID uint64
	}

	var messageDeleteRequest = MessageToDelete{}

	if err := json.Unmarshal(jsonBytes, &messageDeleteRequest); err != nil {
		log.Printf("Error deserializing onDeleteChatMessageRequest json of user ID [%d], reason: %s\n", userID, err.Error())
		return nil
	}

	ownerID, dbSuccess := database.GetChatMessageOwner(messageDeleteRequest.MessageID)
	if !dbSuccess {
		return nil
	}

	if ownerID != userID {
		log.Printf("User ID [%d] is trying to delete someone else's message [%d], aborting\n", userID, messageDeleteRequest.MessageID)
		return setProblem(fmt.Sprintf("Could not delete message ID [%d]\n", userID))
	}

	success := database.DeleteChatMessage(messageDeleteRequest.MessageID)
	if !success {
		return setProblem(fmt.Sprintf("Could not delete message ID [%d]\n", userID))
	}

	messagesBytes, err := json.Marshal(messageDeleteRequest)
	if err != nil {
		log.Panicf("Error serializing json at onDeleteChatMessageRequest for user ID [%d], reason: %s\n:", userID, err.Error())
	}
	return preparePacket(3, messagesBytes)
}

// when client is requesting chat history for a channel
func onChatHistoryRequest(packetJson []byte, userID uint64) []byte {
	type ChatHistoryRequest struct {
		ChannelID uint64
	}

	var chatHistoryRequest ChatHistoryRequest

	if err := json.Unmarshal(packetJson, &chatHistoryRequest); err != nil {
		log.Printf("Error deserializing onChatHistoryRequest json of user ID [%d], reason: %s\n", userID, err.Error())
		return nil
	}

	// type ChatMessageResponseList struct {
	// 	Messages []ChatMessageResponse
	// }

	// var messages = ChatMessageResponseList{
	// 	Messages: database.GetMessagesFromChannel(chatHistoryRequest.ChannelID),
	// }

	var messages []ChatMessageResponse = database.GetMessagesFromChannel(chatHistoryRequest.ChannelID)

	messagesBytes, err := json.Marshal(messages)
	if err != nil {
		log.Panicf("Error serializing json at onChatHistoryRequest for user ID [%d], reason: %s\n:", userID, err.Error())
	}
	return preparePacket(2, messagesBytes)
}
