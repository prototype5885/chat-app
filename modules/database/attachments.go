package database

import (
	log "proto-chat/modules/logging"
)

type Attachment struct {
	FileName string
	UserID   uint64
}

const (
	insertAttachmentQuery = "INSERT INTO attachments (file_name, user_id) VALUES (?, ?)"
	//deleteChatMessageQuery = "DELETE FROM messages WHERE message_id = ? AND user_id = ?"
)

func CreateAttachmentsTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS attachments (
    	file_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY, 
		file_name CHAR(32) NOT NULL,
		user_id BIGINT UNSIGNED NOT NULL,
		INDEX file_name (file_name),
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating attachments table")
	}
}

//func InsertAttachment(fileName string, userID uint64) string {
//	log.Trace("Inserting attachment [%s] for user ID [%d]", fileName, userID)
//	const query = "INSERT INTO attachments (message_id) VALUES (?) RETURNING file_name"
//
//	var fileName string
//	err := db.QueryRow(query, messageID).Scan(&fileName)
//	if err != nil {
//		log.FatalError(err.Error(), "Error inserting attachment for message ID [%d]", messageID)
//	}
//	if fileName == "" {
//		log.Impossible("The returned database generated incremental filename for attachment belonging to message ID [%d] is empty", messageID)
//	}
//	return fileName
//}
