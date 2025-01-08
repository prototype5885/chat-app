package websocket

import (
	"encoding/binary"
	"encoding/json"
	"net/http"
	"proto-chat/modules/clients"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//const sessionTokenLength = 16

const (
	ADD_CHAT_MESSAGE    byte = 1
	CHAT_HISTORY        byte = 2
	DELETE_CHAT_MESSAGE byte = 3

	ADD_SERVER         byte = 21
	SERVER_LIST        byte = 22
	DELETE_SERVER      byte = 23
	SERVER_INVITE_LINK byte = 24

	ADD_CHANNEL    byte = 31
	CHANNEL_LIST   byte = 32
	DELETE_CHANNEL byte = 33

	ADD_SERVER_MEMBER          byte = 41
	SERVER_MEMBER_LIST         byte = 42
	DELETE_SERVER_MEMBER       byte = 43
	UPDATE_MEMBER_DISPLAY_NAME byte = 44
	UPDATE_MEMBER_PROFILE_PIC  byte = 45

	UPDATE_STATUS byte = 53
	UPDATE_ONLINE byte = 55

	ADD_FRIEND byte = 61
	BLOCK_USER byte = 62
	UNFRIEND   byte = 63

	INITIAL_USER_DATA       byte = 241
	IMAGE_HOST_ADDRESS      byte = 242
	UPDATE_USER_DATA        byte = 243
	UPDATE_USER_PROFILE_PIC byte = 244
)

const (
	timeoutWrite   = 10 * time.Second // timeout in x seconds after writing fails for 10 seconds
	timeout        = 60 * time.Second // timeout in x seconds if no pong or message received
	pingPeriod     = 30 * time.Second // sends ping in x interval
	maxMessageSize = 8192             // sever won't continue reading message if it's larger than x bytes
)

var ImageHost string

var upgrader = websocket.Upgrader{
	ReadBufferSize:    4096,
	WriteBufferSize:   4096,
	EnableCompression: true,
}

type BroadcastData struct {
	MessageBytes    []byte
	Type            byte
	AffectedServers []uint64
	AffectedChannel uint64
	AffectedUserID  []uint64
}

type WsClient struct {
	SessionID uint64
	UserID    uint64
	WsConn    *websocket.Conn
	WriteChan chan []byte
	CloseChan chan bool
}

var broadcastChan = make(chan BroadcastData, 100)

var wsClients sync.Map

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

	// session ID is used as key value for Clients hashmap to make it possible
	// for a single user to connect to chat from multiple devices/browsers

	// add client to central area
	sessionID := clients.AddClient(userID)

	// sending and reading messages are two separate goroutines
	// it's so they cant block each other
	// they communicate using channels
	wsClient := &WsClient{
		SessionID: sessionID,
		UserID:    userID,
		WsConn:    wsConn,
		WriteChan: make(chan []byte, 10),
		CloseChan: make(chan bool),
	}

	// add to wsClients
	wsClients.Store(sessionID, wsClient)
	defer wsClient.removeWsClient()

	// create 2 goroutines for reading and writing messages
	var wg sync.WaitGroup
	wg.Add(2)

	go wsClient.readMessages(&wg)
	go wsClient.writeMessages(&wg)

	log.Trace("Session ID [%d] as user ID [%d] has been added to WsClients", sessionID, userID)

	// sends the initial data
	wsClient.onInitialDataRequest()

	setUserStatusText(userID, "Online")
	setUserOnline(userID, true)

	// this will block here while both the reading and writing goroutine are running
	// if one stops, the other should stop too
	wg.Wait()
}

func (c *WsClient) removeWsClient() {
	setUserStatusText(c.UserID, "Offline")
	setUserOnline(c.UserID, false)
	log.Trace("Removing session ID [%d] from WsClients", c.SessionID)
	clients.RemoveClient(c.SessionID)
	wsClients.Delete(c.SessionID)
}

func (c *WsClient) readMessages(wg *sync.WaitGroup) {
	defer func() { // this will run when readMessages goroutine returns
		c.CloseChan <- true // tells the writing goroutine to stop too
		wg.Done()
	}()

	c.WsConn.SetReadLimit(maxMessageSize) // received bytes after this limit will be discareded
	c.WsConn.SetReadDeadline(time.Now().Add(timeout))
	c.WsConn.SetPongHandler(func(string) error { c.WsConn.SetReadDeadline(time.Now().Add(timeout)); return nil })
	for {
		_, receivedBytes, err := c.WsConn.ReadMessage()
		if err != nil {
			log.WarnError(err.Error(), "Failed reading message from session ID [%d] of user ID [%d]", c.SessionID, c.UserID)
			break
		}

		// time.Sleep(500 * time.Millisecond)

		// check if array is at least 5 in length to avoid exceptions
		// because if client sends smaller byte array for some reason,
		// this func would throw an index out of range exception
		// not supposed to happen in normal cases
		if len(receivedBytes) < 5 {
			log.Hack("Session ID [%d] as user ID [%d] sent a byte array shorter than 5 length", c.SessionID, c.UserID)
			c.WriteChan <- macros.RespondFailureReason("Sent byte array length is less than 5")
			continue
		}

		// convert the first 4 bytes into uint32 to get the endIndex,
		// which marks the end of the packet
		var endIndex uint32 = binary.LittleEndian.Uint32(receivedBytes[:4])

		// check if the extracted endIndex is outside the received array bounds to avoid exception
		// not supposed to happen in normal cases
		if endIndex > uint32(len(receivedBytes)) {
			log.Hack("User ID [%d] sent a byte array where the extracted endIndex was larger than the received byte array", c.UserID)
			log.Hack("Byte array of user ID [%d]: [%s]", c.UserID, receivedBytes)
			c.WriteChan <- macros.RespondFailureReason("Sent byte array is longer than the given endIndex value")
			continue
		}

		// 5th byte is a 1 byte number which states the type of the packet
		var packetType byte = receivedBytes[4]

		// get the json byte array from the 6th byte to the end
		var packetJson []byte = receivedBytes[5:endIndex]

		log.Trace("Received packet: endIndex [%d], type [%d], json [%s]", endIndex, packetType, string(packetJson))
		switch packetType {
		case ADD_CHAT_MESSAGE: // user sent a chat message on x channel
			log.Trace("User ID [%d] sent a chat message", c.UserID)
			broadcastData, failData := c.onChatMessageRequest(packetJson)
			if failData != nil {
				c.WriteChan <- failData
			} else {
				broadcastChan <- broadcastData
			}

		case CHAT_HISTORY: // user entered a channel, requesting chat history
			log.Trace("User ID [%d] is asking for a chat history", c.UserID)
			chatHistoryBytes := c.onChatHistoryRequest(packetJson)
			if chatHistoryBytes != nil {
				c.WriteChan <- chatHistoryBytes
			}

		case DELETE_CHAT_MESSAGE: // user deleting a chat message
			log.Trace("User ID [%d] wants to delete a chat message", c.UserID)
			broadcastData, failData := c.onChatMessageDeleteRequest(packetJson)
			if failData != nil {
				c.WriteChan <- failData
			} else {
				broadcastChan <- broadcastData
			}

		case ADD_SERVER: // user adding a server
			log.Trace("User ID [%d] wants to create a server", c.UserID)
			c.WriteChan <- c.onAddServerRequest(packetJson)

		// case SERVER_LIST: // user requesting their server list
		// 	log.Trace("User ID [%d] is requesting their joined server list", c.UserID)
		// 	c.WriteChan <- macros.PreparePacket(22, *database.GetServerList(c.UserID))

		case DELETE_SERVER: // user deleting a server
			log.Trace("User ID [%d] wants to delete a server", c.UserID)
			broadcastChan <- c.onServerDeleteRequest(packetJson)

		case SERVER_INVITE_LINK: // user requested an invite link for a server
			c.WriteChan <- c.onServerInviteRequest(packetJson)

		case ADD_CHANNEL: // user added a channel to their server
			log.Trace("User ID [%d] wants to add a channel", c.UserID)
			broadcastData, failData := c.onAddChannelRequest(packetJson)
			if failData != nil {
				c.WriteChan <- failData
			} else {
				broadcastChan <- broadcastData
			}

		case CHANNEL_LIST: // user entered a server, requesting channel list
			log.Trace("User ID [%d] is requesting channel list of a server", c.UserID)
			channelListbytes := c.onChannelListRequest(packetJson)
			if channelListbytes != nil {
				c.WriteChan <- channelListbytes
			}
		case ADD_SERVER_MEMBER: // a new user connected to the server

		case SERVER_MEMBER_LIST: // user entered a server, requesting member list
			c.WriteChan <- c.onServerMemberListRequest(packetJson)

		case DELETE_SERVER_MEMBER: // a user left a server
			log.Trace("User ID [%d] is requesting to leave from a server", c.UserID)
			// broadcastData, failData := c.onLeaveServerRequest(packetJson)
			// if failData != nil {
			// 	c.WriteChan <- failData
			// } else {
			// 	broadcastChan <- broadcastData
			// 	c.WriteChan <- broadcastData.MessageBytes
			// }
		case UPDATE_STATUS: // user wants to update their status value
			log.Trace("User ID [%d] is requesting to update their status value", c.UserID)
			c.onUpdateUserStatusValue(packetJson)
		case ADD_FRIEND: // user wants to add an other user as friend
			c.onAddFriendRequest(packetJson)
		case BLOCK_USER: // user wants to block a user
			c.onBlockUserRequest(packetJson)
		case UNFRIEND: // user wants to unfriend a user
			c.onUnfriendRequest(packetJson)
		case INITIAL_USER_DATA: // user requests initial data
			c.onInitialDataRequest()
		case IMAGE_HOST_ADDRESS:
			log.Trace("User ID [%d] is requesting address of image host server", c.UserID)
			imageHostJson, err := json.Marshal(ImageHost)
			if err != nil {
				log.FatalError(err.Error(), "Error serializing ImageHost [%s]", ImageHost)
			}
			c.WriteChan <- macros.PreparePacket(IMAGE_HOST_ADDRESS, imageHostJson)
		case UPDATE_USER_DATA: // user wants to update their account data
			log.Trace("User ID [%d] is requesting to update their account data", c.UserID)
			c.onUpdateUserDataRequest(packetJson)

		default: // if unknown
			log.Hack("User ID [%d] sent invalid packet type: [%d]", c.UserID, packetType)
			c.WriteChan <- macros.RespondFailureReason("Packet type is invalid")
		}
	}
}

func (c *WsClient) writeMessages(wg *sync.WaitGroup) {
	ticker := time.NewTicker(pingPeriod) // client will be pinged in intervals using this
	defer ticker.Stop()
	defer wg.Done()

	errorWriting := func(errMsg string) {
		log.WarnError(errMsg, "Error writing message to session ID [%d] as user ID [%d]", c.SessionID, c.UserID)
	}

	for {
		select {
		case messageBytes := <-c.WriteChan:
			c.WsConn.SetWriteDeadline(time.Now().Add(timeoutWrite))
			if err := c.WsConn.WriteMessage(websocket.BinaryMessage, messageBytes); err != nil {
				errorWriting(err.Error())
				return
			}
			log.Trace("Wrote to user ID [%d] session token [%d]", c.UserID, c.SessionID)
		case <-ticker.C:
			// log.Trace("Pinging:", c.userID)
			c.WsConn.SetWriteDeadline(time.Now().Add(timeoutWrite))
			if err := c.WsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				errorWriting(err.Error())
				return
			}
		case closed := <-c.CloseChan:
			if closed {
				log.Debug("Session ID [%d] as user ID [%d] received a signal to close writeMessages goroutine", c.SessionID, c.UserID)
				return
			}
		}
	}
}

func broadCastChannel() {
	broadcastLog := func(typ byte, userID uint64, session uint64) {
		log.Trace("Broadcasting message type [%d] to user ID [%d] session token [%d]", typ, userID, session)
	}

	for {
		select {
		case broadcastData := <-broadcastChan:
			switch broadcastData.Type {
			case ADD_CHAT_MESSAGE, DELETE_CHAT_MESSAGE: // things that only affect a single channel
				wsClients.Range(func(key, value interface{}) bool {
					wsClient, ok := value.(*WsClient)
					if !ok {
						log.Warn("Invalid WsClient")
						return true
					}
					channelID, found := clients.GetCurrentChannelID(wsClient.SessionID)
					if !found {
						return true
					}
					if channelID == broadcastData.AffectedChannel { // if client is in affected channel
						broadcastLog(broadcastData.Type, wsClient.UserID, wsClient.SessionID)
						wsClient.WriteChan <- broadcastData.MessageBytes
					}
					return true
				})
			case ADD_CHANNEL, DELETE_CHANNEL, ADD_SERVER_MEMBER, DELETE_SERVER_MEMBER: // things that only affect a single server
				wsClients.Range(func(key, value interface{}) bool {
					wsClient, ok := value.(*WsClient)
					if !ok {
						log.Warn("Invalid WsClient")
						return true
					}
					serverID, found := clients.GetCurrentServerID(wsClient.SessionID)
					if !found {
						return true
					}
					if serverID == broadcastData.AffectedServers[0] { // if client is currently in that server
						broadcastLog(broadcastData.Type, wsClient.UserID, wsClient.SessionID)
						wsClient.WriteChan <- broadcastData.MessageBytes
					}
					return true
				})

			case UPDATE_MEMBER_PROFILE_PIC, UPDATE_ONLINE, UPDATE_STATUS, UPDATE_MEMBER_DISPLAY_NAME, DELETE_SERVER: // things that affect multiple servers
				wsClients.Range(func(key, value interface{}) bool {
					wsClient, ok := value.(*WsClient)
					if !ok {
						log.Warn("Invalid WsClient")
						return true
					}
					for s := 0; s < len(broadcastData.AffectedServers); s++ {
						serverID, found := clients.GetCurrentServerID(wsClient.SessionID)
						if !found {
							return true
						}
						if serverID == broadcastData.AffectedServers[s] { // if client is member of any affected server
							broadcastLog(broadcastData.Type, wsClient.UserID, wsClient.SessionID)
							wsClient.WriteChan <- broadcastData.MessageBytes
						}
					}
					return true
				})
			case UPDATE_USER_DATA, UPDATE_USER_PROFILE_PIC: // things that only affect a single user, sending to all connected sessions/devices
				wsClients.Range(func(key, value interface{}) bool {
					wsClient, ok := value.(*WsClient)
					if !ok {
						log.Warn("Invalid WsClient")
						return true
					}
					if wsClient.UserID == broadcastData.AffectedUserID[0] {
						broadcastLog(broadcastData.Type, wsClient.UserID, wsClient.SessionID)
						wsClient.WriteChan <- broadcastData.MessageBytes
					}
					return true
				})
			case ADD_FRIEND, BLOCK_USER, UNFRIEND: // things that affect multiple users
				wsClients.Range(func(key, value interface{}) bool {
					wsClient, ok := value.(*WsClient)
					if !ok {
						log.Warn("Invalid WsClient")
						return true
					}
					for u := 0; u < len(broadcastData.AffectedUserID); u++ {
						if wsClient.UserID == broadcastData.AffectedUserID[u] {
							broadcastLog(broadcastData.Type, wsClient.UserID, wsClient.SessionID)
							wsClient.WriteChan <- broadcastData.MessageBytes
						}
					}
					return true
				})
			}
		}
	}
}
