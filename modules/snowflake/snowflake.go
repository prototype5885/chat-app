package snowflake

import (
	"fmt"
	"math"
	log "proto-chat/modules/logging"
	"sync"
	"time"
)

const (
	timestampLength uint64 = 42                                    // 42
	timestampPos    uint64 = 64 - timestampLength                  // 20
	workerLength    uint64 = 10                                    // 10
	workerPos       uint64 = timestampPos - workerLength           // 17
	incrementLength        = 64 - (timestampLength + workerLength) // 12
)

var (
	maxWorkerValue    = uint64(math.Pow(2, float64(workerLength)) - 1)
	maxIncrementValue = uint64(math.Pow(2, float64(incrementLength)) - 1)

	maxTimestamp uint64 = uint64(math.Pow(2, float64(timestampLength))) // max possible timestamp value possible

	lastIncrement, lastTimestamp uint64
	snowflakeMutex               sync.Mutex

	workerID           uint64 = 0
	alreadyHasWorkerID bool   = false
)

func SetSnowflakeWorkerID(id uint64) {
	if id > maxWorkerValue {
		log.Fatal("Worker ID value exceeds maximum value of [%d]", maxWorkerValue)
	}
	if !alreadyHasWorkerID {
		workerID = id
		alreadyHasWorkerID = true
	} else {
		log.Fatal("Worker ID for snowflake generator has been already set, exiting...")
	}
}

func Generate() uint64 {
	snowflakeMutex.Lock()
	defer snowflakeMutex.Unlock()

	var timestamp uint64 = uint64(time.Now().UnixMilli())
	if timestamp == lastTimestamp {
		lastIncrement += 1
		if lastIncrement > maxIncrementValue {
			log.Fatal("Increment overflow")
		}
	} else {
		lastIncrement = 0
		lastTimestamp = timestamp
	}

	return timestamp<<timestampPos | workerID<<workerPos | lastIncrement
}

func Extract(snowflakeId uint64) (uint64, uint64, uint64) {
	var timestamp uint64 = snowflakeId >> timestampPos
	var workerID uint64 = (snowflakeId >> workerPos) & ((1 << workerLength) - 1)
	var increment uint64 = snowflakeId & ((1 << incrementLength) - 1)
	return timestamp, workerID, increment
}

func ExtractTimestamp(snowflakeId uint64) uint64 {
	return snowflakeId >> timestampPos
}

func Print(snowflakeId uint64) {
	timestamp, workerID, increment := Extract(snowflakeId)
	// var realTimestamp = timestamp + timestampOffset

	fmt.Println("-----------------")
	fmt.Println("Snowflake:", snowflakeId)
	fmt.Println("Unix timestamp:", timestamp, "/", maxTimestamp)
	fmt.Println("Date:", time.UnixMilli(int64(timestamp)))
	fmt.Println("Years left:", (math.Pow(2.0, float64(timestampLength))-float64(timestamp))/1000/60/60/24/365)
	// fmt.Println("Real timestamp:", timestamp)
	fmt.Println("Worker:", workerID, "/", maxWorkerValue)
	fmt.Println("Increment:", increment, "/", maxIncrementValue)
	fmt.Println("-----------------")
}
