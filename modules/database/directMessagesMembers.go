package database

import log "proto-chat/modules/logging"

func CreateDmMembersTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS dm_members (
			chat_id BIGINT UNSIGNED NOT NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			FOREIGN KEY (chat_id) REFERENCES dm_chats(chat_id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
			PRIMARY KEY (chat_id, user_id)
		)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating direct message members table")
	}
}
