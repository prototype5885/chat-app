package database

import (
	"encoding/json"
	"fmt"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
)

type Message struct {
	MessageID   uint64
	ChannelID   uint64
	UserID      uint64
	Message     string
	Attachments []byte
}

type UserMessages struct {
	UserID uint64
	Msgs   []interface{}
}

const insertChatMessageQuery = "INSERT INTO messages (message_id, channel_id, user_id, message, attachments) VALUES (?, ?, ?, ?, ?)"

func CreateChatMessagesTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS messages (
			message_id BIGINT UNSIGNED PRIMARY KEY NOT NULL,
			channel_id BIGINT UNSIGNED NOT NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			message TEXT NOT NULL,
			attachments BLOB,
			FOREIGN KEY (channel_id) REFERENCES channels(channel_id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
		)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating messages table")
	}
}
func AddChatMessage(messageID uint64, channelID uint64, userID uint64, chatMessage string, filenames []string) bool {
	var filenamesJson []byte = nil
	if len(filenames) > 0 {
		var err error
		filenamesJson, err = json.Marshal(filenames)
		if err != nil {
			macros.ErrorSerializing(err.Error(), 1, userID)
		}
	}
	err := Insert(Message{
		MessageID:   messageID,
		ChannelID:   channelID,
		UserID:      userID,
		Message:     chatMessage,
		Attachments: filenamesJson,
	})
	if err != nil {
		return false
	} else {
		return true
	}
}

func GetChatHistory(channelID uint64, fromMessageID uint64, older bool, userID uint64) *[]byte {
	const query = "SELECT message_id, user_id, message, attachments FROM messages WHERE channel_id = ? AND (message_id < ? OR ? = 0) ORDER BY message_id DESC LIMIT 50"
	log.Query(query, channelID, fromMessageID, fromMessageID)

	rows, err := Conn.Query(query, channelID, fromMessageID, fromMessageID)
	DatabaseErrorCheck(err)
	var userMessages []UserMessages
	var counter int
	for rows.Next() {
		var userID uint64
		var messageID uint64
		var message string

		var attachmentsJson []byte
		err := rows.Scan(&messageID, &userID, &message, &attachmentsJson)
		DatabaseErrorCheck(err)

		found := false
		index := 0
		for i := 0; i < len(userMessages); i++ {
			if userMessages[i].UserID == userID {
				found = true
				index = i
				break
			}
		}

		if !found {
			userMessages = append(userMessages, UserMessages{UserID: userID})
			index = len(userMessages) - 1
		}

		var attachmentHistory []string
		if attachmentsJson != nil {
			err = json.Unmarshal(attachmentsJson, &attachmentHistory)
			if err != nil {
				log.FatalError(err.Error(), "Error deserializing message attachments retrieved from database of message ID [%d]", messageID)
			}
		}

		userMessages[index].Msgs = append(userMessages[index].Msgs, []interface{}{messageID, message, attachmentHistory})
		counter++
	}
	DatabaseErrorCheck(rows.Err())

	if counter == 0 {
		log.Trace("Channel ID [%d] does not have any messages or user reached top of chat", channelID)
		var emptyResponse []byte = []byte(fmt.Sprintf("[%d, []]", channelID))
		return &emptyResponse
	} else {
		log.Trace("Retrieved [%d] messages from channel ID [%d]", counter, channelID)
	}

	var chatHistory = []interface{}{channelID, userMessages}

	jsonResult, err := json.Marshal(chatHistory)
	if err != nil {
		macros.ErrorSerializing(err.Error(), 2, userID)
	}

	return &jsonResult
}

func DeleteChatMessage(messageID uint64, userID uint64) uint64 {
	const query string = "DELETE FROM messages WHERE message_id = ? AND user_id = ? RETURNING channel_id"
	log.Query(query, messageID, userID)

	var channelID uint64
	err := Conn.QueryRow(query, messageID, userID).Scan(&channelID)
	DatabaseErrorCheck(err)

	if channelID == 0 {
		log.Hack("There is no message ID [%d] owned by user ID [%d]", messageID, userID)
	}

	return channelID
}
