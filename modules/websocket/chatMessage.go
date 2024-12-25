package websocket

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"proto-chat/modules/attachments"
	"proto-chat/modules/clients"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"proto-chat/modules/snowflake"
	"strconv"
)

type ChatMessageResponse struct {
	IDm string // message ID
	IDu string // user ID
	Msg string // message
	Att []string
}

func SerializeChatMessage(messageID uint64, userID uint64, message string, attachments []string) []byte {
	const jsonType string = "response chat message"

	var serverChatMsg = ChatMessageResponse{
		IDm: strconv.FormatUint(messageID, 10),
		IDu: strconv.FormatUint(userID, 10),
		Msg: message,
		Att: attachments,
	}
	jsonBytes, err := json.Marshal(serverChatMsg)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, userID)
	}

	return jsonBytes
}

// when client sent a chat message, type 1
func (c *WsClient) onChatMessageRequest(packetJson []byte, packetType byte) (BroadcastData, []byte) {
	const jsonType string = "received chat message"

	type ClientChatMsg struct {
		ChannelID uint64
		Message   string
		AttTok    string
	}

	var chatMessageRequest ClientChatMsg

	if err := json.Unmarshal(packetJson, &chatMessageRequest); err != nil {
		return BroadcastData{}, macros.ErrorDeserializing(err.Error(), jsonType, c.UserID)
	}

	var rejectMessage = fmt.Sprintf("Denied sending chat message to channel ID [%d]", chatMessageRequest.ChannelID)

	// check if user is member of the server which the channel belongs to
	var serverID uint64 = database.GetServerIdOfChannel(chatMessageRequest.ChannelID)
	if serverID == 0 {
		return BroadcastData{}, macros.RespondFailureReason("%s", rejectMessage)
	}
	if !database.ConfirmServerMembership(c.UserID, serverID) {
		return BroadcastData{}, macros.RespondFailureReason("%s", rejectMessage)
	}

	attachmentToken, err := base64.StdEncoding.DecodeString(chatMessageRequest.AttTok)
	if err != nil {
		log.Hack("User ID [%d] sent an attachmentToken base64 string that can't be decoded", c.UserID)
		return BroadcastData{}, macros.RespondFailureReason("%s", rejectMessage)
	}

	var fileNames []string
	if len(attachmentToken) > 0 {
		fileNames = attachments.GetWaitingAttachment([64]byte(attachmentToken))
	}

	var messageID = snowflake.Generate()

	success := database.AddChatMessage(c.UserID, chatMessageRequest.ChannelID, chatMessageRequest.Message, fileNames)
	if !success {
		return BroadcastData{}, macros.RespondFailureReason("%s", rejectMessage)
	}

	jsonBytes := SerializeChatMessage(messageID, c.UserID, chatMessageRequest.Message, fileNames)

	return BroadcastData{
		MessageBytes:    macros.PreparePacket(1, jsonBytes),
		Type:            packetType,
		AffectedChannel: chatMessageRequest.ChannelID,
	}, nil
}

// when client is requesting chat history for a channel, type 2
func (c *WsClient) onChatHistoryRequest(packetJson []byte, packetType byte) []byte {
	type ChatHistoryRequest struct {
		ChannelID     uint64
		FromMessageID uint64
		Older         bool
	}

	var req ChatHistoryRequest

	if err := json.Unmarshal(packetJson, &req); err != nil {
		return macros.ErrorDeserializing(err.Error(), "chat history", c.UserID)
	}

	issue := clients.SetCurrentChannelID(c.SessionID, req.ChannelID)
	if checkClient(c.SessionID, issue, 1) {
		return nil
	}

	const rejectionMessage = "Denied chat history request"
	// check if user is member of server channel is part of
	serverID := database.GetServerIdOfChannel(req.ChannelID)
	if serverID == 0 {
		return macros.RespondFailureReason(rejectionMessage)
	}
	if !database.ConfirmServerMembership(c.UserID, serverID) {
		return macros.RespondFailureReason(rejectionMessage)
	}

	var jsonBytes []byte = database.GetChatHistory(req.ChannelID, req.FromMessageID, req.Older, c.UserID)
	if jsonBytes == nil {
		return macros.RespondFailureReason(rejectionMessage)
	}

	return macros.PreparePacket(packetType, jsonBytes)
}

// when client wants to delete a message they own, type 3
func (c *WsClient) onChatMessageDeleteRequest(packetJson []byte, packetType byte) (BroadcastData, []byte) {
	const jsonType string = "chat message deletion"

	type MessageToDelete struct {
		MessageID uint64
	}

	var messageDeleteRequest = MessageToDelete{}

	if err := json.Unmarshal(packetJson, &messageDeleteRequest); err != nil {
		return BroadcastData{
			MessageBytes: macros.ErrorDeserializing(err.Error(), jsonType, c.UserID),
		}, nil
	}

	// get the channel ID where the message was deleted,
	// so can broadcoast it to affected Clients
	var channelID uint64 = database.DeleteChatMessage(messageDeleteRequest.MessageID, c.UserID)
	if channelID == 0 {
		return BroadcastData{}, macros.RespondFailureReason("Denied to delete chat message")
	}

	responseBytes, err := json.Marshal(strconv.FormatUint(messageDeleteRequest.MessageID, 10))
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.UserID)
	}

	return BroadcastData{
		MessageBytes:    macros.PreparePacket(packetType, responseBytes),
		AffectedChannel: channelID,
		Type:            packetType,
	}, nil
}
