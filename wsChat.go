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
	log.Printf("User ID [%d] is connecting to websocket...\n", userID)
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		log.Printf("Error upgrading connection of user ID [%d] to websocket protocol\n", userID)
		return
	}
	username := database.GetUsername(userID)
	if username == "" {
		log.Panicf("After accepting websocket client, user ID [%d] has no username set in the database\n", userID)
	}

	client := &Client{
		displayName: username,
		wsConn:      wsConn,
	}

	addClient(client, userID)

	go client.readMessages(userID)
	log.Printf("User ID [%d] has connected to the websocket\n", userID)
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

		// this will be sent back to the sender
		var responseBytes []byte

		// check if array is at least 5 in length to avoid exceptions
		// because if client sends smaller byte array for some reason,
		// this func would throw an index out of range exception
		// not supposed to happen in normal cases
		if len(receivedBytes) < 5 {
			log.Printf("HACK: User ID [%d] sent a byte array shorter than 5 length\n", userID)
			responseBytes = setProblem("Sent byte array length is less than 5")
			c.respondOnlyToSender(userID, responseBytes)
			continue
		}

		// convert the first 4 bytes into uint32 to get the endIndex,
		// which marks the end of the packet
		var endIndex uint32 = binary.LittleEndian.Uint32(receivedBytes[:4])
		// log.Println("endIndex:", endIndex)

		// check if the extracted endIndex is outside of the received array bounds to avoid exception
		// not supposed to happen in normal cases
		if endIndex > uint32(len(receivedBytes)) {
			log.Printf("HACK: User ID [%d] sent a byte array where the extracted endIndex was larger than the received byte array\n", userID)
			responseBytes = setProblem("Sent byte array is longer than the given endIndex value")
			c.respondOnlyToSender(userID, responseBytes)
			continue
		}

		// 5th byte is a 1 byte number which states the type of the packet
		var packetType byte = receivedBytes[4]
		// log.Println("packetType:", packetType)

		// get the json byte array from the 6th byte to the end
		var packetJson []byte = receivedBytes[5:endIndex]

		log.Println("Received packet:", endIndex, packetType, string(packetJson))

		switch packetType {
		case 1: // client sent a chat message
			log.Printf("User ID [%d] sent a chat message\n", userID)
			responseBytes = onChatMessageRequest(packetJson, userID)
			c.sendToEveryone(responseBytes)

		case 2: // client requested server history
			log.Printf("User ID [%d] is asking for chat history\n", userID)
			responseBytes = onChatHistoryRequest(packetJson, userID)
			c.respondOnlyToSender(userID, responseBytes)

		case 3: // client sent a delete message request
			log.Printf("User ID [%d] wants to delete a chat message\n", userID)
			responseBytes = onDeleteChatMessageRequest(packetJson, userID)
			c.sendToEveryone(responseBytes)

		case 21: // client is requesting to add a server
			log.Printf("User ID [%d] wants to create a server\n", userID)
			responseBytes = onAddServerRequest(packetJson, userID)
			c.respondOnlyToSender(userID, responseBytes)

		case 22: // client requested server list
			log.Printf("User ID [%d] is requesting server list\n", userID)
			responseBytes = onServerListRequest(userID)
			c.respondOnlyToSender(userID, responseBytes)

		case 31: // client is requeting to add a channel
			log.Printf("User ID [%d] wants to add a channel\n", userID)
			responseBytes = onAddChannelRequest(packetJson, userID)
			c.sendToEveryone(responseBytes)

		case 32: // client requested channel list
			log.Printf("User ID [%d] is requesting channel list\n", userID)
			responseBytes = onChannelListRequest(packetJson, userID)
			c.respondOnlyToSender(userID, responseBytes)

		case 42: // client is requesting to send names
			log.Printf("User ID [%d] is requesting name/names of servers/channels/users\n", userID)

		default:
			log.Printf("Unable to process message that user ID [%d] sent\n", userID)
			responseBytes = preparePacket(0, nil)
			c.respondOnlyToSender(userID, responseBytes)
		}

		// if responseBytes == nil {
		// 	log.Printf("Unable to process message that user ID [%d] sent", userID)
		// 	responseBytes = preparePacket(0, nil)
		// }

		// // reply only to the sender
		// if respondOnlyToSender {
		// 	if err := c.wsConn.WriteMessage(websocket.BinaryMessage, responseBytes); err != nil {
		// 		removeClient(userID, fmt.Sprintf("Error sending message, error: %s\n", err.Error()))
		// 	}
		// 	continue
		// }
		// // send to everyone otherwise
		// for id := range clients {
		// 	if err := clients[id].wsConn.WriteMessage(websocket.BinaryMessage, responseBytes); err != nil {
		// 		removeClient(id, fmt.Sprintf("Error sending message, error: %s\n", err.Error()))
		// 	}
		// }
	}
}

func (c *Client) respondOnlyToSender(userID uint64, responseBytes []byte) {
	if err := c.wsConn.WriteMessage(websocket.BinaryMessage, responseBytes); err != nil {
		removeClient(userID, fmt.Sprintf("Error sending message, error: %s\n", err.Error()))
	}
}

func (c *Client) sendToEveryone(responseBytes []byte) {
	for id := range clients {
		if err := clients[id].wsConn.WriteMessage(websocket.BinaryMessage, responseBytes); err != nil {
			removeClient(id, fmt.Sprintf("Error sending message, error: %s\n", err.Error()))
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
