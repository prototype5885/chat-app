package database

import log "proto-chat/modules/logging"

type PrivateChatMember struct {
	UserID     uint64
	Name       string
	Pic        string
	Online     bool
	Status     byte
	StatusText string
}

func CreatePrivateChatMembersTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS privchat_members (
			chat_id BIGINT UNSIGNED NOT NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			FOREIGN KEY (chat_id) REFERENCES privchats(chat_id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
			PRIMARY KEY (chat_id, user_id)
		)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating server_members table")
	}
}
