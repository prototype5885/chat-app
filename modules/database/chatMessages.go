package database

import (
	log "proto-chat/modules/logging"
)

type ChatMessage struct {
	MessageID uint64
	ChannelID uint64
	UserID    uint64
	Message   string
}

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
