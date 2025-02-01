package database

import (
	log "chat-app/modules/logging"
)

type Bot struct {
	Key string
}

const insertBotQuery = `INSERT INTO bots (bot_id, owner_id, display_name, status, status_text, picture, token) VALUES (?, ?, ?, ?, ?, ?, ?)`
const deleteBotQuery = "DELETE FROM bots WHERE bot_id = ?"

func CreateBotTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS bots (
			bot_id BIGINT UNSIGNED PRIMARY KEY,
			owner_id BIGINT UNSIGNED NOT NULL,
			display_name VARCHAR(32) NOT NULL,
			status TINYINT UNSIGNED	NOT NULL,
			status_text VARCHAR(32) NOT NULL,
			picture VARCHAR(255) NOT NULL,
			token BINARY(128),
			unique (token)
		)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating invite keys table")
	}
}

//func ConfirmInviteKey(key string) bool {
//	const query = "SELECT EXISTS (SELECT 1 FROM invite_keys WHERE invite_key = ?)"
//	log.Query(query, key)
//
//	var exists bool = false
//	err := Conn.QueryRow(query, key).Scan(&exists)
//	DatabaseErrorCheck(err)
//
//	return exists
//}
