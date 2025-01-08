package database

import (
	"encoding/hex"
	log "proto-chat/modules/logging"
	"strconv"
	"strings"
)

type Attachment struct {
	FileName      uint64
	FileExtension string
	MessageID     uint64
}

const (
	insertAttachmentQuery = "INSERT INTO attachments (name, hash, message_id) VALUES (?, ?, ?)"
	//deleteChatMessageQuery = "DELETE FROM messages WHERE message_id = ? AND user_id = ?"
)

func CreateAttachmentsTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS attachments (
		hash BINARY(32) NOT NULL,
		original_hash BINARY(32) NOT NULL,
		message_id BIGINT UNSIGNED NOT NULL,
		position TINYINT UNSIGNED NOT NULL,
		INDEX hash (hash),
		INDEX original_hash (original_hash),
		INDEX message_id (message_id),
		FOREIGN KEY (message_id) REFERENCES messages(message_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating attachments table")
	}
}

func InsertAttachment(originalHash [32]byte, messageID uint64) {
	fileName := hex.EncodeToString(originalHash[:])

	log.Trace("Inserting attachment [%s] for message ID [%d]", fileName, messageID)

	parts := strings.Split(fileName, ".")

	name, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		log.Hack(err.Error(), "Error parsing attachment fileName [%s] for message ID [%d]", fileName, messageID)
	}

	attachment := Attachment{
		FileName:      name,
		FileExtension: parts[1],
		MessageID:     messageID,
	}

	err = Insert(attachment)
	if err != nil {
		log.Error("Failed inserting attachment [%s] for message ID [%d]", fileName, messageID)
	}
}
