package webRequests

import (
	"math"
	log "proto-chat/modules/logging"
	"strconv"
	"time"
)

var maxIncrementValue = uint64(math.Pow(2, float64(22)) - 1)

var lastTimestamp, lastIncrement uint64

func GenerateAttachmentName() string {
	var timestamp uint64 = uint64(time.Now().UnixMilli())
	if timestamp == lastTimestamp {
		lastIncrement += 1
		if lastIncrement > maxIncrementValue { // this is physically impossible to happen
			log.Fatal("Increment overflow generating attachment name")
		}
	} else {
		lastIncrement = 0
		lastTimestamp = timestamp
	}

	return strconv.FormatUint(timestamp<<22|lastIncrement, 10)
}
