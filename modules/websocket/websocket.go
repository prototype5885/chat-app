package websocket

import (
	"encoding/binary"
	"encoding/json"
	"net/http"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"proto-chat/modules/snowflake"
	"strconv"
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
	ReadBufferSize:    4096,
	WriteBufferSize:   4096,
	EnableCompression: false,
}

type BroadcastData struct {
	MessageBytes []byte
	Type         byte
	ID           uint64
}

type Client struct {
	displayName      string
	wsConn           *websocket.Conn
	sessionID        uint64
	userID           uint64
	currentChannelID uint64
	currentServerID  uint64
	writeChan        chan []byte
	closeChan        chan bool
}

var broadcastChan = make(chan BroadcastData, 100)

// var mutex sync.Mutex // used so only 1 goroutine can access the clients list at one time

var clients = make(map[uint64]*Client)

func Init() {
	go broadCastChannel()
}

// client is connecting to the websocket
func AcceptWsClient(userID uint64, w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WarnError(err.Error(), "Error upgrading connection of user ID [%d] to websocket protocol", userID)
		return
	}
	username := database.GetUsername(userID)
	if username == "" {
		// no idea why this would happen
		log.Impossible("After accepting websocket client, user ID [%d] has no username set in the database", userID)
	}

	// sending and reading messages are two separate goroutines
	// it's so they cant block each other
	// they communicate using channels

	// session ID is used as key value for clients hashmap to make it possible
	// for a single user to connect to chat from multiple devices/browsers
	var sessionID uint64 = snowflake.Generate()

	client := &Client{
		displayName:      username,
		wsConn:           wsConn,
		sessionID:        sessionID,
		userID:           userID,
		currentChannelID: 0,
		writeChan:        make(chan []byte, 10),
		closeChan:        make(chan bool),
	}

	// add to clients hashmap
	clients[sessionID] = client
	// log.Info("Added user ID [%d] with session ID [%d] to the connected websocket clients list", userID, sessionID)

	// create 2 goroutines for reading and writing messages
	var wg sync.WaitGroup
	wg.Add(2)

	go client.readMessages(&wg)
	go client.writeMessages(&wg)

	log.Info("Session ID [%d] as user ID [%d] has connected to websocket", sessionID, userID)

	// sends the client it's own user ID
	jsonUserID, err := json.Marshal(strconv.FormatUint(userID, 10))
	if err != nil {
		macros.ErrorSerializing(err.Error(), "userID", client.userID)
	}

	client.writeChan <- macros.PreparePacket(241, jsonUserID)

	// this will block here while both the reading and writing goroutine are running
	// if one stops, the other should stop too
	wg.Wait()

	// close websocket connection
	// if err := wsConn.Close(); err != nil {
	// 	log.WarnError(err.Error(), "Error closing websocket connection for user ID [%d]", userID)
	// }

	// lastly remove the client from hashmap
	delete(clients, sessionID)
	log.Info("Removed session ID [%d] as user ID [%d] from the connected clients", sessionID, userID)
}

func (c *Client) setCurrentChannelID(channelID uint64) {
	c.currentChannelID = channelID
	log.Trace("User ID [%d] is now on channel ID [%d]", c.userID, channelID)
}

func (c *Client) setCurrentServerID(serverID uint64) {
	c.currentServerID = serverID
	log.Trace("User ID [%d] is now on server ID [%d]", c.userID, c.currentServerID)
}

func (c *Client) readMessages(wg *sync.WaitGroup) {
	defer func() { // this will run when readMessages goroutine returns
		c.closeChan <- true // tells the writing goroutine to stop too
		wg.Done()
	}()

	c.wsConn.SetReadLimit(maxMessageSize) // received bytes after this limit will be discareded
	c.wsConn.SetReadDeadline(time.Now().Add(timeout))
	c.wsConn.SetPongHandler(func(string) error { c.wsConn.SetReadDeadline(time.Now().Add(timeout)); return nil })
	for {
		_, receivedBytes, err := c.wsConn.ReadMessage()
		if err != nil {
			log.WarnError(err.Error(), "Failed reading message from session ID [%d] as user ID [%d]", c.sessionID, c.userID)
			break
		}

		// check if array is at least 5 in length to avoid exceptions
		// because if client sends smaller byte array for some reason,
		// this func would throw an index out of range exception
		// not supposed to happen in normal cases
		if len(receivedBytes) < 5 {
			log.Hack("Session ID [%d] as user ID [%d] sent a byte array shorter than 5 length", c.sessionID, c.userID)
			c.writeChan <- macros.RespondFailureReason("Sent byte array length is less than 5")
			continue
		}

		// convert the first 4 bytes into uint32 to get the endIndex,
		// which marks the end of the packet
		var endIndex uint32 = binary.LittleEndian.Uint32(receivedBytes[:4])
		// log.Println("endIndex:", endIndex)

		// check if the extracted endIndex is outside of the received array bounds to avoid exception
		// not supposed to happen in normal cases
		if endIndex > uint32(len(receivedBytes)) {
			log.Hack("User ID [%d] sent a byte array where the extracted endIndex was larger than the received byte array", c.userID)
			log.Hack("Byte array of user ID [%d]: [%s]", c.userID, receivedBytes)
			c.writeChan <- macros.RespondFailureReason("Sent byte array is longer than the given endIndex value")
			continue
		}

		// 5th byte is a 1 byte number which states the type of the packet
		var packetType byte = receivedBytes[4]

		// get the json byte array from the 6th byte to the end
		var packetJson []byte = receivedBytes[5:endIndex]

		log.Trace("Received packet: endIndex [%d], type [%d], json [%s]", endIndex, packetType, string(packetJson))
		switch packetType {
		case 1: // user sent a chat message on x channel
			log.Debug("User ID [%d] sent a chat message", c.userID)
			broadcastChan <- c.onChatMessageRequest(packetJson, packetType)

		case 2: // user entered a channel, requesting chat history
			log.Debug("User ID [%d] is asking for chat history", c.userID)
			c.writeChan <- c.onChatHistoryRequest(packetJson, packetType)

		case 3: // user deleting a chat message
			log.Debug("User ID [%d] wants to delete a chat message", c.userID)
			broadcastData, failData := c.onChatMessageDeleteRequest(packetJson, packetType)
			if failData != nil {
				c.writeChan <- failData
			} else {
				broadcastChan <- broadcastData
			}

		case 21: // user adding a server
			log.Debug("User ID [%d] wants to create a server", c.userID)
			c.writeChan <- c.onAddServerRequest(packetJson)

		case 22: // user requesting their joined server list
			log.Debug("User ID [%d] is requesting server list", c.userID)
			c.writeChan <- c.onServerListRequest()

		case 23: // user deleting a server
			log.Debug("User ID [%d] wants to delete a server", c.userID)
			broadcastChan <- c.onServerDeleteRequest(packetJson, packetType)

		case 24: // user requested an invite link for a server
			log.Debug("User ID [%d] is requesting an invite link for a server", c.userID)
			c.writeChan <- c.onServerInviteRequest(packetJson)

		case 31: // user added a channel to their server
			log.Debug("User ID [%d] wants to add a channel", c.userID)
			broadcastData, failData := c.onAddChannelRequest(packetJson, packetType)
			if failData != nil {
				c.writeChan <- failData
			} else {
				broadcastChan <- broadcastData
			}

		case 32: // user entered a server, requesting channel list
			log.Debug("User ID [%d] is requesting channel list of a server", c.userID)
			c.writeChan <- c.onChannelListRequest(packetJson)

		case 41: // a new user connected to the server

		case 42: // user entered a server, requesting member list
			log.Debug("User ID [%d] is requesting list of members of server ID [%d]", c.userID, c.currentServerID)
			c.writeChan <- c.onServerMemberListRequest(packetJson)

		case 43: // a user left a server
			log.Debug("User ID [%d] is requesting to leave from a server", c.userID)
			broadcastData, failData := c.onServerMemberDeleteRequest(packetJson, packetType)
			if failData != nil {
				c.writeChan <- failData
			} else {
				broadcastChan <- broadcastData
				c.writeChan <- broadcastData.MessageBytes
			}

		// case 42: // client is requesting to send names
		// 	log.Printf("User ID [%d] is requesting name/names of servers/channels/users", userID)

		default: // if unknown
			log.Hack("User ID [%d] sent invalid packet type: [%d]", c.userID, packetType)
			c.writeChan <- macros.RespondFailureReason("Packet type is invalid")
		}
	}
}

func (c *Client) writeMessages(wg *sync.WaitGroup) {
	ticker := time.NewTicker(pingPeriod) // client will be pinged in intervals using this

	defer func() { // this will run when writeMessages goroutine returns
		ticker.Stop()
		wg.Done()
	}()

	errorWriting := func(errMsg string) {
		log.WarnError(errMsg, "Error writing message to session ID [%d] as user ID [%d]", c.sessionID, c.userID)
	}

	for {
		select {
		case messageBytes := <-c.writeChan:
			c.wsConn.SetWriteDeadline(time.Now().Add(timeoutWrite))
			if err := c.wsConn.WriteMessage(websocket.BinaryMessage, messageBytes); err != nil {
				errorWriting(err.Error())
				return
			}
			log.Trace("Wrote to user ID [%d]", c.userID)
		case <-ticker.C:
			// log.Trace("Pinging:", c.userID)
			c.wsConn.SetWriteDeadline(time.Now().Add(timeoutWrite))
			if err := c.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				errorWriting(err.Error())
				return
			}
		case close := <-c.closeChan:
			if close {
				log.Debug("Session ID [%d] as user ID [%d] received a signal to close writeMessages goroutine", c.sessionID, c.userID)
				return
			}
		}

	}
}

func broadCastChannel() {
	broadcastLog := func(typ byte, userID uint64) {
		log.Trace("Broadcasting message type [%d] to user ID [%d]", typ, userID)
	}

	for {
		select {
		case broadcastData := <-broadcastChan:
			switch broadcastData.Type {
			case 1, 3: // chat messages
				for _, client := range clients {
					if client.currentChannelID == broadcastData.ID { // if client is in affected channel
						broadcastLog(broadcastData.Type, client.userID)
						client.writeChan <- broadcastData.MessageBytes
					}
				}
			case 21, 23: // servers
				for _, client := range clients {
					client.writeChan <- broadcastData.MessageBytes

				}
			case 31, 33: //channels
				for _, client := range clients {
					if client.currentServerID == broadcastData.ID { // if client is in affected server
						broadcastLog(broadcastData.Type, client.userID)
						client.writeChan <- broadcastData.MessageBytes
					}
				}
			case 41, 43: // server members
				for _, client := range clients {
					if client.currentServerID == broadcastData.ID { // if client is in affected server
						broadcastLog(broadcastData.Type, client.userID)
						client.writeChan <- broadcastData.MessageBytes
					}
				}
			}
		}
	}
}

// func Compress(dataToCompress []byte) []byte {
// 	log.Debug("Size before compression: [%d]", len(dataToCompress))

// 	var buffer bytes.Buffer
// 	writer := gzip.NewWriter(&buffer)

// 	writer.Write(dataToCompress)
// 	writer.Close()

// 	log.Debug("Size after compression: [%d]", len(buffer.Bytes()))

// 	return buffer.Bytes()
// }

// func Decompress(compressedData []byte) []byte {
// 	reader, _ := gzip.NewReader(bytes.NewReader(compressedData))

// 	decompressed, _ := io.ReadAll(reader)

// 	log.Debug("Size after decompression: [%d]", len(decompressed))

// 	return decompressed
// }
