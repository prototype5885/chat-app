package attachments

import (
	"crypto/rand"
	"encoding/base64"
	log "proto-chat/modules/logging"
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

	awaitingAttachmentsMap.Store(attachmentToken, uploadedAttachments)
	go removeUnusedAttachmentToken(attachmentToken)
	return attachmentToken
}

func GetWaitingAttachment(attachmentToken [64]byte) []UploadedAttachment {
	log.Trace("Getting attachment token for a message [%s]", base64.StdEncoding.EncodeToString(attachmentToken[:]))
	defer removeWaitingAttachment(&attachmentToken)
	value, loadOk := awaitingAttachmentsMap.Load(attachmentToken)
	if loadOk {
		uploadedAttachments, ok := value.([]UploadedAttachment)
		if ok {
			return uploadedAttachments
		} else {
			log.Impossible("Retrieved attachment from awaitingAttachments are not in AwaitingAttachment struct format")
		}
	} else {
		log.Warn("No attachment filenames were found in attachment token [%s]", base64.StdEncoding.EncodeToString(attachmentToken[:]))
	}
	return []UploadedAttachment{}
}

func removeWaitingAttachment(attachmentToken *[64]byte) {
	log.Trace("Removing attachment token [%s]", base64.StdEncoding.EncodeToString(attachmentToken[:]))
	awaitingAttachmentsMap.Delete(*attachmentToken)
}

func removeUnusedAttachmentToken(attachmentToken [64]byte) {
	// the attachment token will be removed if user for some reason doesn't send the chat message within 15
	time.Sleep(15 * time.Second)
	_, found := awaitingAttachmentsMap.Load(attachmentToken)
	if found {
		log.Warn("Attachment token wasn't claimed by uploader, removing [%s]", base64.StdEncoding.EncodeToString(attachmentToken[:]))
		awaitingAttachmentsMap.Delete(attachmentToken)
	}
}
