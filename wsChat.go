package main

import (
	"encoding/binary"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	timeoutWrite   = 10 * time.Second // timeout in x seconds after writing fails for 10 seconds
	timeout        = 60 * time.Second // timeout in x seconds if no pong or message received
	pingPeriod     = 30 * time.Second // sends ping in x interval
	maxMessageSize = 8192             // sever won't continue reading message if it's larger than x bytes
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
}

type Client struct {
	displayName    string
	wsConn         *websocket.Conn
	userID         uint64
	currentChannel uint64
	writeChan      chan []byte
	closeChan      chan bool
}

type BroadcastData struct {
	MessageBytes []byte
	ChannelID    uint64
}

var broadcastChan = make(chan BroadcastData)

// var mutex sync.Mutex // used so only 1 goroutine can access the clients list at one time

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

	// sending and reading messages are two separate goroutines
	// it's so they cant block each other
	// they communicate using channels

	client := &Client{
		displayName:    username,
		wsConn:         wsConn,
		userID:         userID,
		currentChannel: 0,
		writeChan:      make(chan []byte),
		closeChan:      make(chan bool),
	}

	clients[userID] = client
	log.Printf("Added user ID %d to the connected clients list\n", userID)

	var wg sync.WaitGroup

	wg.Add(2)

	go client.readMessages(&wg)
	go client.writeMessages(&wg)

	log.Printf("User ID [%d] has connected to the websocket\n", userID)

	wg.Wait()

	log.Printf("User ID [%d] has been disconnected successfully\n", userID)
}

func (c *Client) removeClient() {
	c.wsConn.Close()
	delete(clients, c.userID)
	log.Printf("Removed user ID [%d] from the connected clients\n", c.userID)
}

func (c *Client) readMessages(wg *sync.WaitGroup) {
	defer func() {
		c.removeClient()
		c.closeChan <- true
		wg.Done()
	}()

	c.wsConn.SetReadLimit(maxMessageSize)
	c.wsConn.SetReadDeadline(time.Now().Add(timeout))
	c.wsConn.SetPongHandler(func(string) error { c.wsConn.SetReadDeadline(time.Now().Add(timeout)); return nil })
	for {
		_, receivedBytes, err := c.wsConn.ReadMessage()
		if err != nil {
			log.Printf("User ID [%d]: %s\n", c.userID, err.Error())
			break
		}

		// check if array is at least 5 in length to avoid exceptions
		// because if client sends smaller byte array for some reason,
		// this func would throw an index out of range exception
		// not supposed to happen in normal cases
		if len(receivedBytes) < 5 {
			log.Printf("HACK: User ID [%d] sent a byte array shorter than 5 length\n", c.userID)
			c.writeChan <- respondFailureReason("Sent byte array length is less than 5")
			continue
		}

		// convert the first 4 bytes into uint32 to get the endIndex,
		// which marks the end of the packet
		var endIndex uint32 = binary.LittleEndian.Uint32(receivedBytes[:4])
		// log.Println("endIndex:", endIndex)

		// check if the extracted endIndex is outside of the received array bounds to avoid exception
		// not supposed to happen in normal cases
		if endIndex > uint32(len(receivedBytes)) {
			log.Printf("HACK: User ID [%d] sent a byte array where the extracted endIndex was larger than the received byte array\n", c.userID)
			c.writeChan <- respondFailureReason("Sent byte array is longer than the given endIndex value")
			continue
		}

		// 5th byte is a 1 byte number which states the type of the packet
		var packetType byte = receivedBytes[4]
		// log.Println("packetType:", packetType)

		// get the json byte array from the 6th byte to the end
		var packetJson []byte = receivedBytes[5:endIndex]

		log.Println("Received packet:", endIndex, packetType, string(packetJson))

		switch packetType {
		case 1: // user sent a chat message on x channel
			log.Printf("User ID [%d] sent a chat message\n", c.userID)
			broadcastChan <- BroadcastData{
				MessageBytes: c.onChatMessageRequest(packetJson),
				ChannelID:    1811029797519753216,
			}

		case 2: // user entered a channel, requesting chat history
			log.Printf("User ID [%d] is asking for chat history\n", c.userID)
			c.writeChan <- c.onChatHistoryRequest(packetJson)

		case 3: // user deleting a chat message
			log.Printf("User ID [%d] wants to delete a chat message\n", c.userID)
			// broadcastChan <- c.onChatMessageDeleteRequest(packetJson)

		case 21: // user adding a server
			log.Printf("User ID [%d] wants to create a server\n", c.userID)
			// broadcastChan <- c.onAddServerRequest(packetJson)

		case 22: // user requesting their joined server list
			log.Printf("User ID [%d] is requesting server list\n", c.userID)
			c.writeChan <- c.onServerListRequest()

		case 23: // user deleting a server
			log.Printf("User ID [%d] wants to delete a server\n", c.userID)
			// broadcastChan <- c.onServerDeleteRequest(packetJson)

		case 31: // user added a channel to their server
			log.Printf("User ID [%d] wants to add a channel\n", c.userID)
			// broadcastChan <- c.onAddChannelRequest(packetJson)

		case 32: // user requesting channel list for x server
			log.Printf("User ID [%d] is requesting channel list\n", c.userID)
			c.writeChan <- c.onChannelListRequest(packetJson)

		// case 42: // client is requesting to send names
		// 	log.Printf("User ID [%d] is requesting name/names of servers/channels/users\n", userID)

		default: // if unknown
			log.Printf("User ID [%d] sent invalid packet type: [%d]\n", c.userID, packetType)
			c.writeChan <- respondFailureReason("Packet type is invalid")
		}
	}
}

func (c *Client) writeMessages(wg *sync.WaitGroup) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.wsConn.Close()
		wg.Done()
	}()

	for {
		select {
		case messageBytes := <-c.writeChan:
			c.wsConn.SetWriteDeadline(time.Now().Add(timeoutWrite))
			if err := c.wsConn.WriteMessage(websocket.BinaryMessage, messageBytes); err != nil {
				return
			}
			log.Printf("Wrote to user ID [%d]\n", c.userID)
		case <-ticker.C:
			// log.Println("Pinging:", c.userID)
			c.wsConn.SetWriteDeadline(time.Now().Add(timeoutWrite))
			if err := c.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case close := <-c.closeChan:
			if close {
				log.Printf("User ID [%d] received a signal to close writeMessages goroutine\n", c.userID)
				return
			}
		}

	}
}

func broadCastChannel() {
	for {
		select {
		case broadcastData := <-broadcastChan:
			for _, client := range clients {
				if client.currentChannel == broadcastData.ChannelID {
					client.writeChan <- broadcastData.MessageBytes
				}
			}
		}
	}
}
