package database

import (
	"database/sql"
	log "proto-chat/modules/logging"
	"proto-chat/modules/snowflake"
)

type ChatMessage struct {
	MessageID uint64
	ChannelID uint64
	UserID    uint64
	Timestamp uint64
	Message   string
}

const (
	insertChatMessageQuery = "INSERT INTO messages (message_id, channel_id, user_id, timestamp, message) VALUES (?, ?, ?, ?, ?)"
	deleteChatMessageQuery = "DELETE FROM messages WHERE message_id = ? AND user_id = ?"
)

func CreateChatMessagesTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		message_id BIGINT UNSIGNED PRIMARY KEY NOT NULL,
		channel_id BIGINT UNSIGNED NOT NULL,
		user_id BIGINT UNSIGNED NOT NULL,
		timestamp BIGINT UNSIGNED NOT NULL,
		message TEXT NOT NULL,
		INDEX timestamp (timestamp),
		FOREIGN KEY (channel_id) REFERENCES channels(channel_id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating messages table")
	}
}

func AddChatMessage(userID uint64, channelID uint64, message string) bool {
	var serverID uint64 = GetServerOfChannel(channelID)
	if serverID == 0 {
		return false
	}

	if !ConfirmServerMembership(userID, serverID) {
		return false
	}

	var snowflakeID uint64 = snowflake.Generate()

	Insert(ChatMessage{
		MessageID: snowflakeID,
		ChannelID: channelID,
		UserID:    userID,
		Timestamp: snowflake.ExtractTimestamp(snowflakeID),
		Message:   message,
	})
	return true
}

func GetChatHistory(channelID uint64, fromMessageID uint64, older bool, userID uint64) []byte {
	log.Debug("Getting chat message history of channel ID [%d] from database...", channelID)

	var serverID uint64 = GetServerOfChannel(channelID)
	if serverID == 0 {
		return nil
	}

	if !ConfirmServerMembership(userID, serverID) {
		// log.Hack("Can't add chat message from user ID [%d] into channel ID [%d] because user isn't in server ID [%d]", userID, channelID, serverID)
		return nil
	}

	const query string = `
		SELECT JSON_ARRAYAGG(JSON_OBJECT(
			'IDm', CAST(message_id AS CHAR),
			'IDu', CAST(user_id AS CHAR),
			'Msg', message
		)) AS json_result
		FROM (
			SELECT message_id, user_id, message
			FROM messages
			WHERE channel_id = ? AND (message_id < ? OR ? = 0)
			ORDER BY timestamp DESC
			LIMIT 30
		) AS messages_chunk;
	`

	var jsonResult []byte
	err := db.QueryRow(query, channelID, fromMessageID, fromMessageID).Scan(&jsonResult)
	if err != nil {
		log.FatalError(err.Error(), "Error getting chat history of channel ID [%d] for user ID [%d]", channelID, userID)
	}

	if len(jsonResult) == 0 {
		log.Trace("Channel ID [%d] does not have any messages", channelID)
		return nullJson
	}

	return jsonResult
}

func DeleteChatMessage(messageID uint64, userID uint64) uint64 {
	log.Debug("Deleting chat message ID [%d] and returning it's channel ID on the request of user ID [%d]...", messageID, userID)

	var channelID uint64
	const query string = "DELETE FROM messages WHERE message_id = ? AND user_id = ? RETURNING channel_id"
	err := db.QueryRow(query, messageID, userID).Scan(&channelID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Hack("There is no message ID [%d] owned by user ID [%d]", messageID, userID)
			return 0
		}
		log.FatalError(err.Error(), "Error deleting message ID [%d] on the request of user ID [%d]", messageID, userID)
	}
	return channelID
}
