package database

import (
	"database/sql"
	log "proto-chat/modules/logging"
	"proto-chat/modules/structs"
)

type ChatMessage struct {
	MessageID uint64
	ChannelID uint64
	UserID    uint64
	Message   string
}

const (
	insertChatMessageQuery = "INSERT INTO messages (message_id, channel_id, user_id, message) VALUES (?, ?, ?, ?)"
	deleteChatMessageQuery = "DELETE FROM messages WHERE message_id = ? AND user_id = ?"
)

func CreateChatMessagesTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		message_id BIGINT UNSIGNED PRIMARY KEY NOT NULL,
		channel_id BIGINT UNSIGNED NOT NULL,
		user_id BIGINT UNSIGNED NOT NULL,
		message TEXT NOT NULL,
		FOREIGN KEY (channel_id) REFERENCES channels(channel_id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating messages table")
	}
}

func GetChatMessages(channelID uint64, userID uint64) []structs.ChatMessageResponse {
	log.Debug("Getting chat message history of channel ID [%d] from database...", channelID)
	const query string = "SELECT message_id, user_id, message FROM messages WHERE channel_id = ?"

	rows, err := db.Query(query, channelID)
	if err != nil {
		log.FatalError(err.Error(), "Error searching for messages on channel ID [%d] in database", channelID)
	}

	var messages []structs.ChatMessageResponse

	var counter int = 0
	for rows.Next() {
		counter++
		var message structs.ChatMessageResponse
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

	return messages
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
