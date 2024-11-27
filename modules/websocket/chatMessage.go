package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"
	"proto-chat/modules/macros"
	"proto-chat/modules/snowflake"
	"strconv"
)

// when client sent a chat message, type 1
func (c *Client) onChatMessageRequest(packetJson []byte, packetType byte) (BroadcastData, []byte) {
	const jsonType string = "add chat message"

	type ClientChatMsg struct {
		ChannelID uint64
		Message   string
	}

	var chatMessageRequest ClientChatMsg

	if err := json.Unmarshal(packetJson, &chatMessageRequest); err != nil {
		return BroadcastData{}, macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
	}

	var messageID = snowflake.Generate()

	success := database.AddChatMessage(c.userID, chatMessageRequest.ChannelID, chatMessageRequest.Message)
	if !success {
		return BroadcastData{}, macros.RespondFailureReason("Denied adding chat message to channel ID [%d]", chatMessageRequest.ChannelID)
	}

	type ChatMessageResponse struct {
		IDm string // message ID
		IDu string // user ID
		Msg string // message
	}

	var serverChatMsg = ChatMessageResponse{
		IDm: strconv.FormatUint(messageID, 10),
		IDu: strconv.FormatUint(c.userID, 10),
		Msg: chatMessageRequest.Message,
	}

	jsonBytes, err := json.Marshal(serverChatMsg)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	return BroadcastData{
		MessageBytes:    macros.PreparePacket(1, jsonBytes),
		Type:            packetType,
		AffectedChannel: chatMessageRequest.ChannelID,
	}, nil
}

// when client is requesting chat history for a channel, type 2
func (c *Client) onChatHistoryRequest(packetJson []byte, packetType byte) []byte {
	const jsonType string = "chat history"

	type Req struct {
		ChannelID     uint64
		FromMessageID uint64
		Older         bool
	}

	var req Req

	if err := json.Unmarshal(packetJson, &req); err != nil {
		return macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
	}

	var jsonBytes []byte = database.GetChatHistory(req.ChannelID, req.FromMessageID, req.Older, c.userID)
	if jsonBytes == nil {
		return macros.RespondFailureReason("Denied chat history request")
	}

	c.setCurrentChannelID(req.ChannelID)

	return macros.PreparePacket(packetType, jsonBytes)
}

// when client wants to delete a message they own, type 3
func (c *Client) onChatMessageDeleteRequest(packetJson []byte, packetType byte) (BroadcastData, []byte) {
	const jsonType string = "chat message deletion"

	// deserialize json into this
	type MessageToDelete struct {
		MessageID uint64
	}
	var messageDeleteRequest = MessageToDelete{}

	if err := json.Unmarshal(packetJson, &messageDeleteRequest); err != nil {
		return BroadcastData{
			MessageBytes: macros.ErrorDeserializing(err.Error(), jsonType, c.userID),
		}, nil
	}

	// get the channel ID where the message was deleted,
	// so can broadcoast it to affected Clients
	var channelID uint64 = database.DeleteChatMessage(messageDeleteRequest.MessageID, c.userID)
	if channelID == 0 {
		return BroadcastData{}, macros.RespondFailureReason("Denied to delete chat message")
	}

	responseBytes, err := json.Marshal(strconv.FormatUint(messageDeleteRequest.MessageID, 10))
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	return BroadcastData{
		MessageBytes:    macros.PreparePacket(packetType, responseBytes),
		AffectedChannel: channelID,
		Type:            packetType,
	}, nil
}
