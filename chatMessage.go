package main

import (
	"encoding/json"
	"proto-chat/modules/snowflake"
)

type ChatMessageResponse struct {
	IDm uint64 // message ID
	IDu uint64 // user ID
	Msg string // message
}

// when client sent a chat message, type 1
func (c *Client) onChatMessageRequest(jsonBytes []byte) BroadcastData {
	const jsonType string = "add chat message"

	type ClientChatMsg struct {
		ChannelID uint64
		Message   string
	}

	var chatMessageRequest ClientChatMsg

	if err := json.Unmarshal(jsonBytes, &chatMessageRequest); err != nil {
		return BroadcastData{
			MessageBytes: errorDeserializing(err.Error(), jsonType, c.userID),
		}

	}

	var messageID = snowflake.Generate()

	problem := database.AddChatMessage(messageID, chatMessageRequest.ChannelID, c.userID, chatMessageRequest.Message)
	if problem != "" {
		return BroadcastData{
			MessageBytes: respondFailureReason(problem),
		}
	}

	var serverChatMsg = ChatMessageResponse{
		IDm: messageID,
		IDu: c.userID,
		Msg: chatMessageRequest.Message,
	}

	jsonBytes, err := json.Marshal(serverChatMsg)
	if err != nil {
		errorSerializing(err.Error(), jsonType, c.userID)
	}

	return BroadcastData{
		MessageBytes: preparePacket(1, jsonBytes),
		ID:           chatMessageRequest.ChannelID,
	}
}

// when client is requesting chat history for a channel, type 2
func (c *Client) onChatHistoryRequest(packetJson []byte) []byte {
	const jsonType string = "chat history"

	type ChatHistoryRequest struct {
		ChannelID uint64
	}

	var chatHistoryRequest ChatHistoryRequest

	if err := json.Unmarshal(packetJson, &chatHistoryRequest); err != nil {
		return errorDeserializing(err.Error(), jsonType, c.userID)
	}

	var messages []ChatMessageResponse = database.GetMessagesFromChannel(chatHistoryRequest.ChannelID)

	messagesBytes, err := json.Marshal(messages)
	if err != nil {
		errorSerializing(err.Error(), jsonType, c.userID)
	}
	c.currentChannel = chatHistoryRequest.ChannelID
	return preparePacket(2, messagesBytes)
}

// when client wants to delete a message they own, type 3
func (c *Client) onChatMessageDeleteRequest(jsonBytes []byte) BroadcastData {
	const jsonType string = "chat message deletion"

	type MessageToDelete struct {
		MessageID uint64
	}

	var messageDeleteRequest = MessageToDelete{}

	if err := json.Unmarshal(jsonBytes, &messageDeleteRequest); err != nil {
		return BroadcastData{
			MessageBytes: errorDeserializing(err.Error(), jsonType, c.userID),
		}
	}

	channelID := database.DeleteChatMessage(messageDeleteRequest.MessageID, c.userID)
	if channelID == 0 {
		return BroadcastData{
			MessageBytes: respondFailureReason("Couldn't delete chat message"),
		}
	}

	return BroadcastData{
		MessageBytes: preparePacket(3, jsonBytes),
		ID:           channelID,
	}
}
