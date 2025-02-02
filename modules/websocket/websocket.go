package websocket

import (
	"chat-app/modules/clients"
	"chat-app/modules/database"
	log "chat-app/modules/logging"
	"chat-app/modules/macros"
	"encoding/binary"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//const sessionTokenLength = 16

const (
	ADD_CHAT_MESSAGE    byte = 1
	CHAT_HISTORY        byte = 2
	DELETE_CHAT_MESSAGE byte = 3
	STARTED_TYPING      byte = 4
	EDIT_CHAT_MESSAGE   byte = 5

	ADD_SERVER           byte = 21
	UPDATE_SERVER_PIC    byte = 22
	DELETE_SERVER        byte = 23
	SERVER_INVITE_LINK   byte = 24
	UPDATE_SERVER_DATA   byte = 25
	UPDATE_SERVER_BANNER byte = 26

	ADD_CHANNEL         byte = 31
	CHANNEL_LIST        byte = 32
	DELETE_CHANNEL      byte = 33
	UPDATE_CHANNEL_DATA byte = 34

	ADD_SERVER_MEMBER         byte = 41
	SERVER_MEMBER_LIST        byte = 42
	DELETE_SERVER_MEMBER      byte = 43
	UPDATE_MEMBER_DATA        byte = 44
	UPDATE_MEMBER_PROFILE_PIC byte = 45

	UPDATE_STATUS byte = 53
	UPDATE_ONLINE byte = 55

	ADD_FRIEND byte = 61
	BLOCK_USER byte = 62
	UNFRIEND   byte = 63

	OPEN_DM             byte = 71
	REQUEST_DM_LIST     byte = 72
	ADD_DM_CHAT_MESSAGE byte = 73

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
var ParsedImageHost *url.URL
var ImageHostAddress string

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

type SpamProtection struct {
	UserID           uint64
	LastMsgTimestamp int64
	TooFastCount     int
	Timer            *time.Timer
	StopTimer        chan bool
}

var maxTooFastCount int = 40
var resetAfter int64 = 20000
var deleteFromSpamProtectionAfter time.Duration = 29

var broadcastChan = make(chan BroadcastData, 100)

var wsClients sync.Map

var spamClients sync.Map

func Init() {
	go broadCastChannel()
}

// AcceptWsClient client is connecting to the websocket
func AcceptWsClient(userID uint64, w http.ResponseWriter, r *http.Request) {
	log.Trace("Accepting user ID [%d] to websocket...", userID)
	val, exists := spamClients.Load(userID)
	if exists {
		log.Trace("SpamClients thing exists for user ID [%d]", userID)
		spam := val.(SpamProtection)
		if spam.TooFastCount >= maxTooFastCount {
			log.Hack("User ID [%d] who spammed earlier came back, not accepting until spam timer expires", userID)
			return
		}
		log.Trace("Wasn't spamming earlier")
		spam.StopTimer <- true
	}

	log.Trace("Upgrading user ID [%d] to websocket connection", userID)
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
	wsClient.onInitialDataRequest(INITIAL_USER_DATA)

	//setUserStatusText(userID, "Online")
	setUserOnline(userID, true)

	// this will block here while both the reading and writing goroutine are running
	// if one stops, the other should stop too
	wg.Wait()
}

func (c *WsClient) removeWsClient() {
	log.Trace("Removing session ID [%d] from WsClients", c.SessionID)

	c.WriteChan <- macros.RespondFailureReason("You sent too many packets in a short time")

	err := c.WsConn.Close()
	if err != nil {
		log.WarnError(err.Error(), "Error while closing websocket for session ID [%d]", c.SessionID)
	}

	clients.RemoveClient(c.SessionID)
	wsClients.Delete(c.SessionID)

	sessions := clients.GetUserSessions(c.UserID)
	if len(sessions) == 0 {
		setUserOnline(c.UserID, false)

		val, exists := spamClients.Load(c.UserID)
		if !exists {
			log.Error("Why user ID [%d] didn't exist in spamClients while trying to delete it? It was connected to websocket", c.UserID)
			return
		} else {
			spam := val.(SpamProtection)
			spam.Timer.Reset(deleteFromSpamProtectionAfter * time.Second)
			log.Trace("Will remove user ID [%d] from spamClients in [%d] seconds, unless user rejoins earlier", deleteFromSpamProtectionAfter, spam.UserID)
		}
	}
}

func (s *SpamProtection) spamManager() {
	defer spamClients.Delete(s.UserID)
	defer s.Timer.Stop()

	for {
		select {
		case <-s.Timer.C:
			log.Trace("[%d] seconds passed, deleting user ID [%d] from spamClients", resetAfter/1000, s.UserID)
			return
		case <-s.StopTimer:
			log.Trace("User ID [%d] came back early, no need to remove from spamClients anymore", s.UserID)
			s.Timer.Stop()
		}
	}
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

		val, exists := spamClients.LoadOrStore(c.UserID, SpamProtection{
			UserID:           c.UserID,
			LastMsgTimestamp: time.Now().UnixMilli(),
			TooFastCount:     0,
			Timer:            time.NewTimer(deleteFromSpamProtectionAfter * time.Second),
			StopTimer:        make(chan bool),
		})

		spam := val.(SpamProtection)

		if !exists {
			spam.Timer.Stop()
			go spam.spamManager()
		}

		currentTime := time.Now().UnixMilli()
		difference := currentTime - spam.LastMsgTimestamp
		log.Trace("Difference: %d", difference)
		if difference < 1000 {
			spam.TooFastCount++
			//log.Warn("Too fast [%d] times", spam.TooFastCount)

			if spam.TooFastCount > maxTooFastCount {
				log.Hack("User ID [%d] sent too many messages in short time, disconnecting...", c.UserID)
				break
			}
		} else if difference > resetAfter {
			log.Trace("User ID [%d] hasn't sent messages for a while, resetting TooFastCount", c.UserID)
			spam.TooFastCount = 0
		}

		spam.LastMsgTimestamp = time.Now().UnixMilli()

		spamClients.Store(c.UserID, spam)

		//time.Sleep(500 * time.Millisecond)

		// check if array is at least 5 in length to avoid exceptions
		// because if client sends smaller byte array for some reason,
		// this func would throw an index out of range exception
		// not supposed to happen in normal cases
		if len(receivedBytes) < 5 {
			log.Hack("Session ID [%d] as user ID [%d] sent a byte array shorter than 5 length", c.SessionID, c.UserID)
			c.WriteChan <- macros.RespondFailureReason("Sent byte array length is less than 5")
			break
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
			break
		}

		// 5th byte is a 1 byte number which states the type of the packet
		var packetType byte = receivedBytes[4]

		// get the json byte array from the 6th byte to the end
		var packetJson []byte = receivedBytes[5:endIndex]

		log.Trace("Received packet: endIndex [%d], type [%d], json [%s]", endIndex, packetType, string(packetJson))
		switch packetType {
		case ADD_CHAT_MESSAGE: // user sent a chat message on x channel
			c.onAddChatMessageRequest(packetJson, packetType)
		case CHAT_HISTORY: // user entered a channel, requesting chat history
			c.onChatHistoryRequest(packetJson, packetType)
		case DELETE_CHAT_MESSAGE: // user deleting a chat message
			c.onChatMessageDeleteRequest(packetJson, packetType)
		case STARTED_TYPING:
			c.onChatMessageTyping(packetJson, packetType)
		case EDIT_CHAT_MESSAGE:
			c.onChatMessageEditRequest(packetJson, packetType)
		case ADD_SERVER: // user adding a server
			c.onAddServerRequest(packetJson, packetType)
		case DELETE_SERVER: // user deleting a server
			c.onServerDeleteRequest(packetJson, packetType)
		case SERVER_INVITE_LINK: // user requested an invite link for a server
			c.onServerInviteRequest(packetJson, packetType)
		case UPDATE_SERVER_DATA: // user is requesting to update server data of their server
			c.onServerDataUpdateRequest(packetJson, packetType)
		case ADD_CHANNEL: // user added a channel to their server
			c.onAddChannelRequest(packetJson, packetType)
		case CHANNEL_LIST: // user entered a server, requesting channel list
			c.onChannelListRequest(packetJson, packetType)
		case DELETE_CHANNEL: // user wants to delete a channel
			c.onChannelDeleteRequest(packetJson, packetType)
		case UPDATE_CHANNEL_DATA: // user wants to change name of a channel
			c.onChannelDataUpdateRequest(packetJson, packetType)
		case SERVER_MEMBER_LIST: // user entered a server, requesting member list
			c.onServerMemberListRequest(packetJson, packetType)
		case DELETE_SERVER_MEMBER: // a user left a server
			c.onLeaveServerRequest(packetJson, packetType)
		case UPDATE_STATUS: // user wants to update their status value
			c.onUpdateUserStatusValue(packetJson, packetType)
		case ADD_FRIEND: // user wants to add another user as friend
			c.onAddFriendRequest(packetJson, packetType)
		case BLOCK_USER: // user wants to block a user
			c.onBlockUserRequest(packetJson, packetType)
		case UNFRIEND: // user wants to unfriend a user
			c.onUnfriendRequest(packetJson, packetType)
		case OPEN_DM: // user wants to open a dm
			c.onOpenDmRequest(packetJson, packetType)
		case REQUEST_DM_LIST: // user requests list of direct messages they have
			c.WriteChan <- macros.PreparePacket(packetType, database.GetDmListOfUser(c.UserID))
		//case ADD_DM_CHAT_MESSAGE:
		//	c.onAddChatMessageRequest(packetJson, packetType, true)
		case INITIAL_USER_DATA: // user requests initial data
			c.onInitialDataRequest(packetType)
		case IMAGE_HOST_ADDRESS:
			c.onImageHostAddressRequest(packetType)
		case UPDATE_USER_DATA: // user wants to update their account data
			c.onUpdateUserDataRequest(packetJson, packetType)
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
	log.Trace("Started broadcasting...")
	broadcastLog := func(typ byte, userID uint64, session uint64) {
		log.Trace("Broadcasting message type [%d] to user ID [%d] session token [%d]", typ, userID, session)
	}

	for {
		select {
		case broadcastData := <-broadcastChan:
			switch broadcastData.Type {
			case ADD_CHAT_MESSAGE, DELETE_CHAT_MESSAGE, STARTED_TYPING, EDIT_CHAT_MESSAGE: // things that only affect a single channel
				wsClients.Range(func(key, value interface{}) bool {
					wsClient, ok := value.(*WsClient)
					if !ok {
						log.Warn("Invalid WsClient")
						return true
					}
					channelID := clients.GetCurrentChannelID(wsClient.SessionID)
					if channelID == 0 {
						return true
					}
					if channelID == broadcastData.AffectedChannel { // if client is in affected channel
						broadcastLog(broadcastData.Type, wsClient.UserID, wsClient.SessionID)
						wsClient.WriteChan <- broadcastData.MessageBytes
					}
					return true
				})
			case ADD_CHANNEL, DELETE_CHANNEL, ADD_SERVER_MEMBER, DELETE_SERVER_MEMBER, UPDATE_CHANNEL_DATA: // things that only affect a single server
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

			case UPDATE_MEMBER_PROFILE_PIC, UPDATE_ONLINE, UPDATE_STATUS, UPDATE_MEMBER_DATA: // if client is currently on an affected server
				wsClients.Range(func(key, value interface{}) bool {
					wsClient, ok := value.(*WsClient)
					if !ok {
						log.Warn("Invalid WsClient")
						return true
					}
					for s := 0; s < len(broadcastData.AffectedServers); s++ {
						serverID, found := clients.GetCurrentServerID(wsClient.SessionID)
						if !found {
							log.Warn("Failed to get current server ID for session ID [%d] as user ID [%d]", wsClient.SessionID, wsClient.UserID)
							return true
						}
						if serverID == broadcastData.AffectedServers[s] { // if client is member of any affected server
							broadcastLog(broadcastData.Type, wsClient.UserID, wsClient.SessionID)
							wsClient.WriteChan <- broadcastData.MessageBytes
						}
					}
					return true
				})
			case UPDATE_USER_DATA, UPDATE_USER_PROFILE_PIC, ADD_SERVER: // things that only affect a single user, sending to all connected sessions/devices
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
			case ADD_FRIEND, BLOCK_USER, UNFRIEND, UPDATE_SERVER_PIC, DELETE_SERVER, UPDATE_SERVER_DATA, UPDATE_SERVER_BANNER: // things that affect multiple users directly
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
