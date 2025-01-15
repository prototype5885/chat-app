package database

import (
	"encoding/hex"
	log "proto-chat/modules/logging"
)

type Attachment struct {
	Hash      []byte
	MessageID uint64
	Name      string
}

type AttachmentResponse struct {
	Hash []byte
	Name string
}

const (
	insertAttachmentQuery = "INSERT INTO attachments (hash, message_id, name) VALUES (?, ?, ?)"
	//deleteChatMessageQuery = "DELETE FROM messages WHERE message_id = ? AND user_id = ?"
)

func CreateAttachmentsTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS attachments (
		hash BINARY(32) NOT NULL,
		message_id BIGINT UNSIGNED NOT NULL,
		name VARCHAR(255) NOT NULL,
		FOREIGN KEY (message_id) REFERENCES messages(message_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating attachments table")
	}
}

func GetAttachmentsOfMessage(messageID uint64) []AttachmentResponse {
	const query string = "SELECT hash, name FROM attachments WHERE message_id = ?"
	log.Query(query, messageID)

	log.Trace("Running query")
	rows, err := Conn.Query(query, messageID)
	log.Trace("Checking for error")
	DatabaseErrorCheck(err)
	defer rows.Close()

	log.Trace("Starting rows next")
	var attachments []AttachmentResponse
	for rows.Next() {
		log.Trace("Next row")
		//var hash []byte
		//var name string
		attachment := AttachmentResponse{}
		err := rows.Scan(&attachment.Hash, &attachment.Name)
		DatabaseErrorCheck(err)

		attachments = append(attachments, attachment)
	}
	DatabaseErrorCheck(rows.Err())

	if len(attachments) == 0 {
		log.Error("Error in GetAttachmentsOfMessage, somehow message ID [%d] has no attachments despite it being flagged having, removing flag from database...", messageID)
		RemoveHasAttachmentFlag(messageID)
	}
	return attachments
}

func CheckIfAttachmentExists(hash []byte) bool {
	const query string = "SELECT EXISTS (SELECT 1 FROM attachments WHERE hash = ?)"
	log.Query(query, hash)

	var exists bool = false
	err := Conn.QueryRow(query, hash).Scan(&exists)
	DatabaseErrorCheck(err)

	if exists {
		log.Trace("Attachment with hash exists [%s] ", hex.EncodeToString(hash))
	} else {
		log.Hack("Attachment with hash doesn't exist [%s]", hex.EncodeToString(hash))
	}

	return exists
}
