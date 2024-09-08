package main

import (
	"log"
)

func handleWebsocketMessage(receivedBytes []byte, userID uint64, displayName string) []byte {
	log.Println("Received bytes:", string(receivedBytes))

	var packetType byte = receivedBytes[0]
	var packetJson string = string(receivedBytes[1:])

	log.Println("Packet type:", packetType)
	log.Println("Packet json:", packetJson)

	switch packetType {
	case 1: // ClientChatMsg
		return addChatMessage(receivedBytes[1:], userID, displayName)
	}
	return []byte("Unknown packet type")
}
