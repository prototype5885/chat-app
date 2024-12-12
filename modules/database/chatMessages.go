package database

import (
	"database/sql"
	"encoding/json"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"proto-chat/modules/snowflake"
)

type Message struct {
	MessageID uint64
	ChannelID uint64
	UserID    uint64
	// Timestamp   uint64
	Message     string
	Attachments []byte
}

type ChatMessageHistory struct {
	IDm uint64
	IDu uint64
	Msg string
	Att []string
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
func AddChatMessage(userID uint64, channelID uint64, chatMessage string, filenames []string) bool {
	var filenamesJson []byte = nil
	if len(filenames) > 0 {
		var err error
		filenamesJson, err = json.Marshal(filenames)
		if err != nil {
			macros.ErrorSerializing(err.Error(), "add chat chatMessage", userID)
		}
	}
	success := Insert(Message{
		MessageID:   snowflake.Generate(),
		ChannelID:   channelID,
		UserID:      userID,
		Message:     chatMessage,
		Attachments: filenamesJson,
	})
	if !success {
		return false
	}

	return true
}

func GetChatHistory(channelID uint64, fromMessageID uint64, older bool, userID uint64) []byte {
	log.Debug("Retrieving chat message history of channel ID [%d] from database...", channelID)

	const query = `SELECT message_id, user_id, message, attachments
			FROM messages
			WHERE channel_id = ? AND (message_id < ? OR ? = 0)
			ORDER BY message_id DESC
			LIMIT 50`

	rows, err := Conn.Query(query, channelID, fromMessageID, fromMessageID)
	DatabaseErrorCheck(err)

	var chatMessageHistory []ChatMessageHistory
	var counter int
	for rows.Next() {
		var cm ChatMessageHistory
		var attachmentsJson []byte

		err := rows.Scan(&cm.IDm, &cm.IDu, &cm.Msg, &attachmentsJson)
		DatabaseErrorCheck(err)

		if attachmentsJson != nil {
			err = json.Unmarshal(attachmentsJson, &cm.Att)
			if err != nil {
				log.FatalError(err.Error(), "Error deserializing message attachments retrieved from database of message ID [%d]", cm.IDm)
			}
		}

		chatMessageHistory = append(chatMessageHistory, cm)
		counter++
	}
	DatabaseErrorCheck(rows.Err())

	if counter == 0 {
		log.Trace("Channel ID [%d] does not have any messages or user reached top of chat", channelID)
		return nullJson
	} else {
		log.Trace("Retrieved [%d] messages from channel ID [%d]", counter, channelID)
	}

	jsonResult, _ := json.Marshal(chatMessageHistory)

	return jsonResult
}

func DeleteChatMessage(messageID uint64, userID uint64) uint64 {
	log.Debug("Deleting chat message ID [%d] and returning it's channel ID on the request of user ID [%d]...", messageID, userID)

	var channelID uint64
	const query string = "DELETE FROM messages WHERE message_id = ? AND user_id = ? RETURNING channel_id"
	err := Conn.QueryRow(query, messageID, userID).Scan(&channelID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Hack("There is no message ID [%d] owned by user ID [%d]", messageID, userID)
			return 0
		}
		DatabaseErrorCheck(err)
		// log.FatalError(err.Error(), "Error deleting message ID [%d] on the request of user ID [%d]", messageID, userID)
	}
	return channelID
}
