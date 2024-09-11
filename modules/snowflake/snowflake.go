package snowflake

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

const (
	timestampLength uint64 = 44                                    // 44
	timestampPos    uint64 = 64 - timestampLength                  // 20
	serverLength    uint64 = 8                                     // 8
	serverPos       uint64 = timestampPos - serverLength           // 17
	incrementLength        = 64 - (timestampLength + serverLength) // 12
	// timestampOffset uint64 = 1704067200000                         // 2024. january 1.
)

// var maxTimestamp uint64 = uint64(math.Pow(2, float64(timestampLength))) + timestampOffset // max possible timestamp value possible
var maxTimestamp uint64 = uint64(math.Pow(2, float64(timestampLength))) // max possible timestamp value possible

var lastIncrement, lastTimestamp uint64
var snowflakeMutex sync.Mutex

var serverID uint64 = 0
var alreadyHasServerID bool = false

func SetSnowflakeServerID(id uint64) {
	if !alreadyHasServerID {
		serverID = id
		alreadyHasServerID = true
	} else {
		log.Fatalln("Server ID for snowflake generator has been already set, exiting...")
	}
}

func Generate() uint64 {
	snowflakeMutex.Lock()
	defer snowflakeMutex.Unlock()

	// var timestamp uint64 = uint64(time.Now().UnixMilli()) - timestampOffset
	var timestamp uint64 = uint64(time.Now().UnixMilli())
	if timestamp == lastTimestamp {
		lastIncrement += 1
	} else {
		lastIncrement = 0
		lastTimestamp = timestamp
	}

	return timestamp<<timestampPos | serverID<<serverPos | lastIncrement
}

func Extract(snowflakeId uint64) (uint64, uint64, uint64) {
	var timestamp uint64 = snowflakeId >> timestampPos
	var serverId uint64 = (snowflakeId >> serverPos) & ((1 << serverLength) - 1)
	var increment uint64 = snowflakeId & ((1 << incrementLength) - 1)
	return timestamp, serverId, increment
}

func Print(snowflakeId uint64) {
	timestamp, serverId, increment := Extract(snowflakeId)
	// var realTimestamp = timestamp + timestampOffset
	var realTimestamp = timestamp
	fmt.Println("-----------------")
	fmt.Println("Date:", time.UnixMilli(int64(realTimestamp)))
	fmt.Println("Server timestamp:", timestamp, "/", maxTimestamp)
	fmt.Println("Years left:", (math.Pow(2.0, float64(timestampLength))-float64(timestamp))/1000/60/60/24/365)
	fmt.Println("Snowflake:", snowflakeId)
	fmt.Println("Real timestamp:", realTimestamp)
	fmt.Println("Server:", serverId, "/", uint64(math.Pow(2, float64(serverLength))))
	fmt.Println("Increment:", increment, "/", uint64(math.Pow(2, float64(incrementLength))))
	fmt.Println("-----------------")
}
