package macros

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	log "proto-chat/modules/logging"
	"time"
)

func GetTimestamp() int64 {
	return time.Now().UnixMilli()
}

func MeasureTime(start int64, msg string) {
	log.Time("%s took [%d ms]", msg, GetTimestamp()-start)
}

func ErrorDeserializing(errStr string, jsonType string, userID uint64) []byte {
	log.WarnError(errStr, "Error deserializing json type [%s] of user ID [%d]", jsonType, userID)
	return RespondFailureReason(fmt.Sprintf("Couldn't deserialize json of [%s] request", jsonType))
}

func ErrorSerializing(errStr string, jsonType string, userID uint64) {
	log.FatalError(errStr, "Fatal error serializing response json type [%s] for user ID [%d]", jsonType, userID)
}

func RespondFailureReason(format string, v ...any) []byte {
	type Failure struct {
		Reason string
	}
	var failure = Failure{
		Reason: fmt.Sprintf(format, v...),
	}

	json, err := json.Marshal(failure)
	if err != nil {
		log.FatalError(err.Error(), "Could not serialize issue in respondFailureReason")
	}

	return PreparePacket(0, json)
}

func PreparePacket(typeByte byte, jsonBytes []byte) []byte {
	// convert the end index uint32 value into 4 bytes
	var endIndex uint32 = uint32(5 + len(jsonBytes))
	var endIndexBytes []byte = make([]byte, 4)
	binary.LittleEndian.PutUint32(endIndexBytes, endIndex)

	// merge them into a single packet
	var packet []byte = make([]byte, 5+len(jsonBytes))
	copy(packet, endIndexBytes) // first 4 bytes will be the length
	packet[4] = typeByte        // 5th byte will be the packet type
	copy(packet[5:], jsonBytes) // rest will be the json byte array

	log.Trace("Prepared packet: endIndex [%d], type [%d], json [%s]", endIndex, packet[4], string(jsonBytes))

	return packet
}

func ShortenToken(tokenBytes []byte) string {
	var token string = hex.EncodeToString(tokenBytes)
	if len(token) > 8 {
		firstFour := token[:4]
		lastFour := token[len(token)-4:]
		return fmt.Sprintf("%s ... %s", firstFour, lastFour)
	} else {
		log.Hack("Can't shorten token [%s], it's shorter than 4 characters", token)
		return ""
	}
}
