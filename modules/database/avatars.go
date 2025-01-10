package database

import log "proto-chat/modules/logging"

type Avatar struct {
	Hash         [32]byte
	OriginalHash [32]byte
	UserID       uint64
	ServerID     uint64
}

const (
	insertAvatarQuery = "INSERT INTO avatars (hash, original_hash, user_id, server_id) VALUES (?, ?, ?, ?)"
	//deleteChatMessageQuery = "DELETE FROM messages WHERE message_id = ? AND user_id = ?"
)

func CreateAvatarsTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS avatars (
		hash BINARY(32) NOT NULL,
		original_hash BINARY(32) NOT NULL,
		user_id BIGINT UNSIGNED,
		server_id BIGINT UNSIGNED,
		FOREIGN KEY (user_id) REFERENCES servers(server_id) ON DELETE CASCADE,
		FOREIGN KEY (server_id) REFERENCES servers(server_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating avatars table")
	}
}
