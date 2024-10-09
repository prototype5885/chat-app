package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"
	"proto-chat/modules/macros"
	"proto-chat/modules/snowflake"
	"proto-chat/modules/structs"
	"strconv"
)

// when client sent a chat message, type 1
func (c *Client) onChatMessageRequest(packetJson []byte, packetType byte) BroadcastData {
	const jsonType string = "add chat message"

	type ClientChatMsg struct {
		ChannelID uint64
		Message   string
	}

	var chatMessageRequest ClientChatMsg

	if err := json.Unmarshal(packetJson, &chatMessageRequest); err != nil {
		return BroadcastData{
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
		return BroadcastData{
			MessageBytes: macros.RespondFailureReason("Failed adding message"),
		}
	}

	var serverChatMsg = structs.ChatMessageResponse{
		IDm: strconv.FormatUint(messageID, 10),
		IDu: strconv.FormatUint(c.userID, 10),
		Msg: chatMessageRequest.Message,
	}

	jsonBytes, err := json.Marshal(serverChatMsg)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	return BroadcastData{
		MessageBytes: macros.PreparePacket(1, jsonBytes),
		Type:         packetType,
		ID:           chatMessageRequest.ChannelID,
	}
}

// when client is requesting chat history for a channel, type 2
func (c *Client) onChatHistoryRequest(packetJson []byte, packetType byte) []byte {
	const jsonType string = "chat history"

	type ChatHistoryRequest struct {
		ChannelID uint64
	}

	var chatHistoryRequest ChatHistoryRequest

	if err := json.Unmarshal(packetJson, &chatHistoryRequest); err != nil {
		return macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
	}
	var channelID uint64 = chatHistoryRequest.ChannelID

	var chatHistory []structs.ChatMessageResponse = database.GetChatMessages(chatHistoryRequest.ChannelID, c.userID)

	// var chatHistoryBytes []byte

	// for i := 0; i < len(chatHistory); i++ {
	// 	var messageIDbytes []byte = make([]byte, 8)
	// 	binary.BigEndian.PutUint64(messageIDbytes, chatHistory[i].IDm)

	// 	var userIDbytes []byte = make([]byte, 8)
	// 	binary.BigEndian.PutUint64(userIDbytes, chatHistory[i].IDu)

	// 	var messageBytes []byte = []byte(chatHistory[i].Msg)
	// 	var messageLength []byte = make([]byte, 4)
	// 	binary.BigEndian.PutUint32(messageLength, uint32(len(messageBytes)))

	// 	var chatMessageBytes []byte = make([]byte, len(messageBytes)+20)

	// 	copy(chatMessageBytes[:8], messageBytes)
	// 	copy(chatMessageBytes[8:], userIDbytes)
	// 	copy(chatMessageBytes[16:], messageLength)
	// 	copy(chatMessageBytes[20:], messageBytes)

	// 	chatHistoryBytes = append(chatHistoryBytes, chatMessageBytes...)
	// }

	jsonBytes, err := json.Marshal(chatHistory)
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	// log.Debug("json length: [%d]", len(jsonBytes))
	// log.Debug("optimized length: [%d]", len(chatHistoryBytes))

	c.setCurrentChannelID(channelID)

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
	// so can broadcoast it to affected clients
	var channelID uint64 = database.DeleteChatMessage(messageDeleteRequest.MessageID, c.userID)
	if channelID == 0 {
		return BroadcastData{}, macros.RespondFailureReason("Couldn't delete chat message")
	}

	responseBytes, err := json.Marshal(strconv.FormatUint(messageDeleteRequest.MessageID, 10))
	if err != nil {
		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
	}

	return BroadcastData{
		MessageBytes: macros.PreparePacket(packetType, responseBytes),
		ID:           channelID,
		Type:         packetType,
	}, nil
}
