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

type ChatMessageResponse struct {
	IDm uint64 // message ID
	IDu uint64 // user ID
	Msg string // message
}

// when client sent a chat message, type 1
func (c *Client) onChatMessageRequest(jsonBytes []byte) structs.BroadcastData {
	const jsonType string = "add chat message"

	type ClientChatMsg struct {
		ChannelID uint64
		Message   string
	}

	var chatMessageRequest ClientChatMsg

	if err := json.Unmarshal(jsonBytes, &chatMessageRequest); err != nil {
		return structs.BroadcastData{
			MessageBytes: macros.ErrorDeserializing(err.Error(), jsonType, c.userID),
		}

	}

	var messageID = snowflake.Generate()

	success := database.Insert(database.ChatMessage{
		MessageID: messageID,
		ChannelID: chatMessageRequest.ChannelID,
		UserID:    c.userID,
		Message:   chatMessageRequest.Message,
	})
	if !success {
		return structs.BroadcastData{
			MessageBytes: macros.RespondFailureReason("Failed adding message"),
		}
	}

	var serverChatMsg = ChatMessageResponse{
		IDm: messageID,
		IDu: c.userID,
		Msg: chatMessageRequest.Message,
	}

	jsonBytes, err := json.Marshal(serverChatMsg)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	return structs.BroadcastData{
		MessageBytes: macros.PreparePacket(1, jsonBytes),
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
		return macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
	}
	var channelID uint64 = chatHistoryRequest.ChannelID

	var rows *sql.Rows = database.ChannelsTable.GetChatMessages(chatHistoryRequest.ChannelID)
	var messages []ChatMessageResponse

	var counter int = 0
	for rows.Next() {
		counter++
		var message ChatMessageResponse
		err := rows.Scan(&message.IDm, &message.IDu, &message.Msg)
		if err != nil {
			log.FatalError(err.Error(), "Error scanning message row into struct in channel ID [%d]:", channelID)
		}
		messages = append(messages, message)
	}

	if counter == 0 {
		log.Debug("No messages found on channel ID: [%d]", channelID)
	} else {
		log.Debug("[%d] messages from channel ID [%d] were retrieved successfully", counter, channelID)
	}

	jsonBytes, err := json.Marshal(messages)
	if err != nil {
		log.Error(err.Error())
		log.Warn("Fatal error serializing json in GetMessagesFromChannel")
	}

	c.changedChannel(channelID)

	return macros.PreparePacket(2, jsonBytes)
}

// when client wants to delete a message they own, type 3
func (c *Client) onChatMessageDeleteRequest(jsonBytes []byte) structs.BroadcastData {
	const jsonType string = "chat message deletion"

	type MessageToDelete struct {
		MessageID uint64
	}

	var messageDeleteRequest = MessageToDelete{}

	if err := json.Unmarshal(jsonBytes, &messageDeleteRequest); err != nil {
		return structs.BroadcastData{
			MessageBytes: macros.ErrorDeserializing(err.Error(), jsonType, c.userID),
		}
	}

	var messageToDelete = database.ChatMessageDeletion{
		MessageID: messageDeleteRequest.MessageID,
		UserID:    c.userID,
	}

	channelID := database.Delete(messageToDelete)
	if channelID == 0 {
		return structs.BroadcastData{
			MessageBytes: macros.RespondFailureReason("Couldn't delete chat message"),
		}
	}

	return structs.BroadcastData{
		MessageBytes: macros.PreparePacket(3, jsonBytes),
		ID:           channelID,
	}
}
