package database

type Attachment struct {
	FileName      uint64
	FileExtension string
	MessageID     uint64
}

const (
	insertAttachmentQuery = "INSERT INTO attachments (name, hash, message_id) VALUES (?, ?, ?)"
	//deleteChatMessageQuery = "DELETE FROM messages WHERE message_id = ? AND user_id = ?"
)

//
//func CreateAttachmentsTable() {
//	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS attachments (
//		file BIGINT UNSIGNED NOT NULL PRIMARY KEY,
//		hash BINARY(32) NOT NULL,
//		message_id BIGINT UNSIGNED NOT NULL,
//		INDEX hash (hash),
//		INDEX message_id (message_id),
//		FOREIGN KEY (message_id) REFERENCES messages(message_id) ON DELETE CASCADE
//	)`)
//	if err != nil {
//		log.FatalError(err.Error(), "Error creating attachments table")
//	}
//}
//
//func InsertAttachment(fileName string, messageID uint64) {
//	log.Trace("Inserting attachment [%s] for message ID [%d]", fileName, messageID)
//
//	parts := strings.Split(fileName, ".")
//
//	name, err := strconv.ParseUint(parts[0], 10, 64)
//	if err != nil {
//		log.Hack(err.Error(), "Error parsing attachment fileName [%s] for message ID [%d]", fileName, messageID)
//	}
//
//	attachment := Attachment{
//		FileName:      name,
//		FileExtension: parts[1],
//		MessageID:     messageID,
//	}
//
//	success := Insert(attachment)
//	if !success {
//		log.Error("Failed inserting attachment [%s] for message ID [%d]", fileName, messageID)
//	}
//}
