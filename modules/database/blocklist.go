package database

import log "proto-chat/modules/logging"

type BlockUser struct {
	UserID        uint64
	BlockedUserID uint64
}

const insertBlockListQuery string = "INSERT INTO block_list (user_id, blocked_id) VALUES (?, ?)"
const deleteBlockListQuery string = "DELETE FROM block_list WHERE user_id = ? AND blocked_id = ?"

func CreateBlockListTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS block_list (
			user_id BIGINT UNSIGNED PRIMARY KEY,
			blocked_id BIGINT UNSIGNED NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE,
			FOREIGN KEY (blocked_id) REFERENCES users (user_id) ON DELETE CASCADE,
			CHECK (user_id != blocked_id)
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating block list table")
	}
}
