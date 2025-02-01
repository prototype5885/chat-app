package database

import (
	log "chat-app/modules/logging"
	"chat-app/modules/macros"
	"encoding/json"
	"fmt"
)

type Message struct {
	MessageID      uint64
	ChannelID      uint64
	UserID         uint64
	Message        string
	HasAttachments bool
	ReplyID        uint64
}

type RetrievedMessage struct {
	MessageID      uint64
	UserID         uint64
	Message        string
	HasAttachments bool
	Edited         bool
	ReplyID        uint64
}

type UserMessages struct {
	UserID uint64
	Msgs   []interface{}
}

type DeleteMessage struct {
	MessageID uint64
	UserID    uint64
}

const insertChatMessageQuery = "INSERT INTO messages (message_id, channel_id, user_id, message, has_attachments, edited, reply_id) VALUES (?, ?, ?, ?, ?, ?, ?)"
const deleteChatMessageQuery = "DELETE FROM messages WHERE message_id = ? AND user_id = ?"

func CreateChatMessagesTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS messages (
			message_id BIGINT UNSIGNED PRIMARY KEY,
			channel_id BIGINT UNSIGNED NOT NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			message TEXT NOT NULL,
			edited BOOLEAN NOT NULL,
			has_attachments BOOLEAN NOT NULL,
			reply_id BIGINT UNSIGNED NOT NULL default 0,
			FOREIGN KEY (channel_id) REFERENCES channels(channel_id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
		)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating messages table")
	}
}

func GetChatHistory(channelID uint64, fromMessageID uint64, older bool, userID uint64) []byte {
	const query = "SELECT message_id, user_id, message, has_attachments, edited, reply_id FROM messages WHERE channel_id = ? AND (message_id < ? OR ? = 0) ORDER BY message_id DESC LIMIT 50"
	log.Query(query, channelID, fromMessageID, fromMessageID)

	rows, err := Conn.Query(query, channelID, fromMessageID, fromMessageID)
	DatabaseErrorCheck(err)
	var retrievedMsgs []RetrievedMessage
	for rows.Next() {
		retrievedMsg := RetrievedMessage{}

		err := rows.Scan(&retrievedMsg.MessageID, &retrievedMsg.UserID, &retrievedMsg.Message, &retrievedMsg.HasAttachments, &retrievedMsg.Edited, &retrievedMsg.ReplyID)
		DatabaseErrorCheck(err)

		retrievedMsgs = append(retrievedMsgs, retrievedMsg)
	}

	var userMessages []UserMessages
	for m := 0; m < len(retrievedMsgs); m++ {
		found := false
		index := 0
		for u := 0; u < len(userMessages); u++ {
			if userMessages[u].UserID == retrievedMsgs[m].UserID {
				found = true
				index = u
				break
			}
		}

		if !found {
			userMessages = append(userMessages, UserMessages{UserID: retrievedMsgs[m].UserID})
			index = len(userMessages) - 1
		}

		var attachmentHistory []AttachmentResponse
		if retrievedMsgs[m].HasAttachments {
			attachmentHistory = GetAttachmentsOfMessage(retrievedMsgs[m].MessageID)
		}

		log.Trace("Message ID [%d] has [%d] attachments", retrievedMsgs[m].MessageID, len(attachmentHistory))

		userMessages[index].Msgs = append(userMessages[index].Msgs, []interface{}{retrievedMsgs[m].MessageID, retrievedMsgs[m].Message, retrievedMsgs[m].Edited, attachmentHistory, retrievedMsgs[m].ReplyID})
	}

	if len(userMessages) == 0 {
		log.Trace("Channel ID [%d] does not have any messages or user reached top of chat", channelID)
		var emptyResponse []byte = []byte(fmt.Sprintf("[%d, []]", channelID))
		return emptyResponse
	} else {
		log.Trace("Retrieved [%d] messages from channel ID [%d]", len(userMessages), channelID)
	}

	var chatHistory = []interface{}{channelID, userMessages}

	jsonResult, err := json.Marshal(chatHistory)
	if err != nil {
		macros.ErrorSerializing(err.Error(), 2, userID)
	}

	return jsonResult
}

func GetChannelOfMessageID(messageID uint64, userID uint64) uint64 {
	const query1 = "SELECT channel_id FROM messages WHERE message_id = ? AND user_id = ?"
	log.Query(query1, messageID, userID)

	var channelID uint64
	err := Conn.QueryRow(query1, messageID, userID).Scan(&channelID)
	DatabaseErrorCheck(err)

	return channelID
}

func EditChatMessage(messageID uint64, userID uint64, message string) uint64 {
	const query string = "UPDATE messages SET message = ?, edited = true WHERE user_id = ? AND message_id = ? RETURNING channel_id"
	log.Query(query, message, userID, messageID)

	var channelID uint64

	err := Conn.QueryRow(query, message, userID, messageID).Scan(&channelID)
	DatabaseErrorCheck(err)

	return channelID
}

func RemoveHasAttachmentFlag(messageID uint64) {
	const query string = "UPDATE messages SET has_attachments = FALSE WHERE message_id = ?"
	log.Query(query, messageID)

	_, err := Conn.Exec(query, messageID)
	DatabaseErrorCheck(err)
}
