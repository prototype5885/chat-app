package database

import log "proto-chat/modules/logging"

// type Server struct {
// 	ServerID uint64
// 	UserID   uint64
// 	Name     string
// 	Picture  string
// }

// const insertPrivateMsgQuery = "INSERT INTO private_chats (server_id, user_id, name, picture) VALUES (?, ?, ?, ?)"
// const deletePrivateMsgQuery = "DELETE FROM private_chats WHERE server_id = ? AND user_id = ?"

func CreatePrivateChatsTable() {
	// _, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS private_chats (
	// 			chat_id BIGINT UNSIGNED PRIMARY KEY NOT NULL,
	// 			user1_id BIGINT UNSIGNED NOT NULL,
	// 			user2_id BIGINT UNSIGNED NOT NULL,
	// 			FOREIGN KEY (user1_id) REFERENCES users(user_id) ON DELETE CASCADE,
	// 			FOREIGN KEY (user2_id) REFERENCES USERS(user_id) ON DELETE CASCADE,
	// 		)`)
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS privchats (
				chat_id BIGINT UNSIGNED PRIMARY KEY NOT NULL
			)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating private chats table")
	}
}
