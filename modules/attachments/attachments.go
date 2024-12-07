package attachments

import (
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
)

var awaitingAttachments = make(map[[64]byte][]string)

func OnAttachmentUploaded(userID uint64, fileNames []string) []byte {
	log.Trace("Responding to user ID [%d] about the uploaded [%d] attachments", userID, len(fileNames))

	attachmentToken := macros.GenerateRandomBytes()
	awaitingAttachments[[64]byte(attachmentToken)] = fileNames
	return attachmentToken
}

func GetWaitingAttachment(attachmentToken [64]byte) []string {
	defer removeWaitingAttachment(attachmentToken)
	return awaitingAttachments[attachmentToken]
}

func removeWaitingAttachment(attachmentToken [64]byte) {
	delete(awaitingAttachments, attachmentToken)
}
