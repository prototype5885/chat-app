package attachments

import (
	log "chat-app/modules/logging"
	"chat-app/modules/macros"
	"crypto/rand"
	"sync"
	"time"
)

type UploadedAttachment struct {
	Hash [32]byte
	Name string
}

var awaitingAttachmentsMap sync.Map

func generateAttachmentToken() [64]byte {
	var bytes [64]byte
	_, err := rand.Read(bytes[:])
	if err != nil {
		log.FatalError(err.Error(), "Could not generate random bytes for attachment token")
	}
	return bytes
}

func OnAttachmentUploaded(userID uint64, uploadedAttachments []UploadedAttachment) [64]byte {
	log.Trace("Responding to user ID [%d] about the uploaded [%d] attachments", userID, len(uploadedAttachments))

	attachmentToken := generateAttachmentToken()

	log.Trace("User ID [%d] uploaded [%d] attachments into waiting list as attachment token [%s], will expire in 15 seconds", userID, len(uploadedAttachments), macros.ShortenToken(attachmentToken[:]))
	awaitingAttachmentsMap.Store(attachmentToken, uploadedAttachments)
	go removeUnusedAttachmentToken(attachmentToken)
	return attachmentToken
}

func GetWaitingAttachment(attachmentToken [64]byte) []UploadedAttachment {
	log.Trace("A user is claiming attachment token [%s]", macros.ShortenToken(attachmentToken[:]))
	defer removeWaitingAttachment(attachmentToken)
	value, loadOk := awaitingAttachmentsMap.Load(attachmentToken)
	if loadOk {
		uploadedAttachments, ok := value.([]UploadedAttachment)
		if ok {
			log.Trace("Retrieved [%d] attachments with attachment token [%s]", len(uploadedAttachments), macros.ShortenToken(attachmentToken[:]))
			return uploadedAttachments
		} else {
			log.Impossible("Retrieved attachment from awaitingAttachments are not in AwaitingAttachment struct format")
		}
	} else {
		log.Warn("No attachment filenames were found in attachment token [%s]", macros.ShortenToken(attachmentToken[:]))
	}
	return []UploadedAttachment{}
}

func removeWaitingAttachment(attachmentToken [64]byte) {
	log.Trace("Removing attachment token [%s]", macros.ShortenToken(attachmentToken[:]))
	awaitingAttachmentsMap.Delete(attachmentToken)
}

func removeUnusedAttachmentToken(attachmentToken [64]byte) {
	// the attachment token will be removed if user for some reason doesn't send the chat message within 15
	time.Sleep(15 * time.Second)
	_, found := awaitingAttachmentsMap.Load(attachmentToken)
	if found {
		log.Warn("Attachment token wasn't claimed by uploader, removing [%s]", macros.ShortenToken(attachmentToken[:]))
		awaitingAttachmentsMap.Delete(attachmentToken)
	}
}
