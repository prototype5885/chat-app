package database

import log "proto-chat/modules/logging"

func CreateDmTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS dm_chats (
				chat_id BIGINT UNSIGNED PRIMARY KEY NOT NULL
			)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating direct message chats table")
	}
}
