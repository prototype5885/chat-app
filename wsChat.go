package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second // Time allowed to write a Message to the peer.
	pongWait       = 60 * time.Second // Time allowed to read the next pong Message from the peer.
	pingPeriod     = 5 * time.Second  // Send pings to peer with this period. Must be less than pongWait.
	maxMessageSize = 8192             // Maximum Message size allowed from peer.
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	displayName string
	wsConn      *websocket.Conn
}

var mutex sync.Mutex // used so only 1 goroutine can access the clients list at one time
var clients = make(map[uint64]*Client)

// client is connecting to the websocket
func acceptWsClient(userID uint64, w http.ResponseWriter, r *http.Request) {
	printWithID(userID, "Accepting websocket connection")
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error accepting websocket connection: ", err)
		return
	}
	username, nameResult := database.GetUsername(userID)
	if !nameResult.Success {
		panicWithID(userID, "For some reason no username was associated with the given user id", "PANIC")
	}

	client := &Client{
		displayName: username,
		wsConn:      wsConn,
	}

	addClient(client, userID)

	go client.readMessages(userID)
	printWithID(userID, "Client has connected to the websocket")
}

func addClient(client *Client, userID uint64) {
	mutex.Lock()
	defer mutex.Unlock()

	clients[userID] = client
	log.Printf("Added user ID %d to the connected clients\n", userID)
}

func removeClient(userID uint64, reason string) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(clients, userID)
	log.Printf("Removed user ID %d from the connected clients, reason: %s\n", userID, reason)
}

func (c *Client) readMessages(userID uint64) {
	// c.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
	// c.wsConn.SetReadDeadline(time.Now().Add(pongWait))
	// c.wsConn.SetPongHandler(func(string) error { c.wsConn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, receivedBytes, err := c.wsConn.ReadMessage()
		if err != nil {
			removeClient(userID, fmt.Sprintf("Error reading message, error: %s\n", err.Error()))
			// if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			// 	log.Println(err)
			// }
			break
		}
		// convert the first 4 bytes into uint32 to get the endIndex,
		// which marks the end of the packet
		var endIndex uint32 = binary.LittleEndian.Uint32(receivedBytes[:4])

		// 5th byte is a 1 byte number which states the type of the packet
		var packetType byte = receivedBytes[4]

		// get the json byte array from the 6th byte to the end
		var packetJson []byte = receivedBytes[5:endIndex]

		log.Println("Received packet:", endIndex, packetType, string(packetJson))

		var respondOnlyToSender bool = false
		var responseBytes []byte
		switch packetType {
		case 1: // client sent a chat message
			log.Printf("User ID [%d] sent a chat message\n", userID)
			responseBytes = onChatMessageRequest(packetJson, userID, c.displayName)
		case 2: // client requested server history
			log.Printf("User ID [%d] is asking for chat history\n", userID)
			responseBytes = onChatHistoryRequest(packetJson, userID)
			respondOnlyToSender = true
		case 3: // client sent a delete message request
			log.Printf("User ID [%d] wants to delete a chat message\n", userID)
			responseBytes = onDeleteChatMessageRequest(packetJson, userID)
		case 21: // client is requesting to add a server
			log.Printf("User ID [%d] wants to create a server\n", userID)
			responseBytes = onAddServerRequest(packetJson, userID)
			respondOnlyToSender = true
		case 22: // client requested server list
			log.Printf("User ID [%d] is requesting server list\n", userID)
			responseBytes = onServerListRequest(userID)
			respondOnlyToSender = true
		case 31: // client is requeting to add a channel
			log.Printf("User ID [%d] wants to add a channel\n", userID)
			responseBytes = onAddChannelRequest(packetJson, userID)
		case 32: // client requested channel list
			log.Printf("User ID [%d] is requesting channel list\n", userID)
			responseBytes = onChannelListRequest(packetJson, userID)
		}

		if responseBytes == nil {
			printWithID(userID, "User sent a websocket message with unprocessable packet type")
			return
		}

		// reply only to the sender
		if respondOnlyToSender {
			if err := c.wsConn.WriteMessage(websocket.BinaryMessage, responseBytes); err != nil {
				removeClient(userID, fmt.Sprintf("Error sending message, error: %s\n", err.Error()))
			}
			continue
		}
		// send to everyone otherwise
		for id := range clients {
			if err := clients[id].wsConn.WriteMessage(websocket.BinaryMessage, responseBytes); err != nil {
				removeClient(id, fmt.Sprintf("Error sending message, error: %s\n", err.Error()))
			}
		}
	}
}

func pingClients() {
	ticker := time.NewTicker(pingPeriod)
	for {
		select {
		case <-ticker.C:
			for userID, client := range clients {
				// log.Println("Pinging client:", userID)
				if err := client.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
					removeClient(userID, fmt.Sprintf("Error pinging, error: %s\n", err.Error()))
					return
				}
			}

		}
	}

}
