package snowflake

import (
	"time"
)

const timestampLength uint8 = 44
const timestampPos uint8 = 64 - timestampLength

const serverLength uint8 = 8
const serverPos uint8 = timestampPos - serverLength // 17

// 2024. january 1.: 1704067200000
// discord one: 1420070400000
const timestampOffset uint64 = 1704067200000

var lastIncrement uint64 = 0
var lastTimestamp uint64 = 0

// type snowflake struct {
// 	snowflakeId int
// 	timestamp   int
// 	serverId    int
// 	increment   int
// }

func Generate(serverId uint64) uint64 {
	var timestamp = uint64(time.Now().UnixMilli()) - timestampOffset

	if timestamp == lastTimestamp {
		lastIncrement += 1
	} else {
		lastIncrement = 0
		lastTimestamp = timestamp
	}

	return timestamp<<timestampPos | serverId<<serverPos | lastIncrement
}
