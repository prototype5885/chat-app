package database

import (
	log "proto-chat/modules/logging"
)

type InviteKey struct {
	Key string
}

const deleteInviteKeyQuery = "DELETE FROM invite_keys WHERE key = ?"

func CreateInviteKeysTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS invite_keys (
			key TEXT PRIMARY KEY NOT NULL
		)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating invite keys table")
	}
}

func ConfirmInviteKey(key string) bool {
	const query = "SELECT EXISTS (SELECT 1 FROM invite_keys WHERE key = ?)"
	log.Query(query, key)

	var exists bool = false
	err := Conn.QueryRow(query, key).Scan(&exists)
	DatabaseErrorCheck(err)

	return exists
}
