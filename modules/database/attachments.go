package database

import (
	"encoding/hex"
	"path/filepath"
	log "proto-chat/modules/logging"
)

type Attachment struct {
	Hash      []byte
	MessageID uint64
	Name      string
}

const (
	insertAttachmentQuery = "INSERT INTO attachments (hash, message_id, name) VALUES (?, ?, ?)"
	//deleteChatMessageQuery = "DELETE FROM messages WHERE message_id = ? AND user_id = ?"
)

func CreateAttachmentsTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS attachments (
		hash BINARY(32) NOT NULL,
		message_id BIGINT UNSIGNED NOT NULL,
		name VARCHAR(255) NOT NULL DEFAULT '',
		FOREIGN KEY (message_id) REFERENCES messages(message_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating attachments table")
	}
}

func GetAttachmentsOfMessage(messageID uint64) []string {
	const query string = "SELECT hash, name FROM attachments WHERE message_id = ?"
	log.Query(query, messageID)

	log.Trace("Running query")
	rows, err := Conn.Query(query, messageID)
	log.Trace("Checking for error")
	DatabaseErrorCheck(err)
	defer rows.Close()

	log.Trace("Starting rows next")
	var names []string
	for rows.Next() {
		log.Trace("Next row")
		var hash []byte
		var name string
		err := rows.Scan(&hash, &name)
		DatabaseErrorCheck(err)

		hashString := hex.EncodeToString(hash)
		extension := filepath.Ext(name)

		names = append(names, hashString+extension)
	}
	DatabaseErrorCheck(rows.Err())

	if len(names) == 0 {
		log.Impossible("Fatal error in GetAttachmentsOfMessage, somehow message ID [%d] has no attachments despite it being flagged having", messageID)
	}
	return names
}
