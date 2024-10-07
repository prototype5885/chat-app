package database

import (
	log "proto-chat/modules/logging"
	"proto-chat/modules/structs"
)

type ChatMessage struct {
	MessageID uint64
	ChannelID uint64
	UserID    uint64
	Message   string
}

const insertChatMessageQuery string = "INSERT INTO messages (message_id, channel_id, user_id, message) VALUES (?, ?, ?, ?)"

type ChatMessageDeletion struct {
	MessageID uint64
	UserID    uint64
}

func (m *ChatMessages) CreateChatMessagesTable() {
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

func (m *ChatMessages) GetChatMessages(channelID uint64, userID uint64) []structs.ChatMessageResponse {
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
