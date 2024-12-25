package websocket

import (
	"encoding/binary"
	"encoding/json"
	"net/http"
	"proto-chat/modules/clients"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//const sessionTokenLength = 16

const (
	addChatMessage          byte = 1
	chatHistory             byte = 2
	deleteChatMessage       byte = 3
	addServer               byte = 21
	serverList              byte = 22
	deleteServer            byte = 23
	serverInviteLink        byte = 24
	addChannel              byte = 31
	channelList             byte = 32
	deleteChannel           byte = 33
	addServerMember         byte = 41
	serverMemberList        byte = 42
	deleteServerMember      byte = 43
	updateMemberDisplayName byte = 44
	updateMemberProfilePic  byte = 45
	updateStatus            byte = 53
	updateOnline            byte = 55
	imageHostAddress        byte = 242
	updateUserData          byte = 243
	updateUserProfilePic    byte = 244
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
	AffectedUserID  uint64
}

type WsClient struct {
	SessionID uint64
	UserID    uint64
	WsConn    *websocket.Conn
	WriteChan chan []byte
	CloseChan chan bool
}

type InitialUserData struct {
	UserID      uint64
	DisplayName string
	ProfilePic  string
	Pronouns    string
	StatusText  string
}

var broadcastChan = make(chan BroadcastData, 100)

var mu sync.RWMutex // used so only 1 goroutine can access the Clients hashmap at one time

// var wsClients = make(map[uint64]WsClient)

var wsClients []WsClient

func Init() {
	go broadCastChannel()
}

func checkClient(sessionID uint64, issue string, returnedID uint64) bool {
	if issue != "" {
		removeWsClient(sessionID)
		log.Impossible("%s", issue)
		return true
	} else if returnedID == 0 {
		removeWsClient(sessionID)
		log.Impossible("Couldn't get current server or channel ID of session ID [%d] because session was not found", sessionID)
		return true
	} else {
		return false
	}
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

	// add client
	sessionID := clients.AddClient(userID)

	// sending and reading messages are two separate goroutines
	// it's so they cant block each other
	// they communicate using channels
	wsClient := WsClient{
		SessionID: sessionID,
		UserID:    userID,
		WsConn:    wsConn,
		WriteChan: make(chan []byte, 10),
		CloseChan: make(chan bool),
	}

	// add to wsClients
	mu.Lock()
	// wsClients[sessionID] = wsClient
	wsClients = append(wsClients, wsClient)
	mu.Unlock()

	// create 2 goroutines for reading and writing messages
	var wg sync.WaitGroup
	wg.Add(2)

	go wsClient.readMessages(&wg)
	go wsClient.writeMessages(&wg)

	log.Trace("Session ID [%d] as user ID [%d] has been added to WsClients", sessionID, userID)

	// sends the client its own user ID and display name
	displayName, profilePic, statusText, pronouns := database.GetUserData(userID)

	var userData InitialUserData = InitialUserData{
		UserID:      userID,
		DisplayName: displayName,
		ProfilePic:  profilePic,
		Pronouns:    pronouns,
		StatusText:  statusText,
	}

	jsonUserID, err := json.Marshal(userData)
	if err != nil {
		macros.ErrorSerializing(err.Error(), "userID", userID)
	}

	wsClient.WriteChan <- macros.PreparePacket(241, jsonUserID)

	//setUserStatusText(client.userID, "Online")
	setUserOnline(userID, true)

	// this will block here while both the reading and writing goroutine are running
	// if one stops, the other should stop too
	wg.Wait()

	//setUserStatusText(client.userID, "Offline")
	setUserOnline(userID, false)

	// close websocket connection
	// if err := wsConn.Close(); err != nil {
	// 	log.WarnError(err.Error(), "Error closing websocket connection for user ID [%d]", userID)
	// }

	// lastly remove the client
	removeWsClient(sessionID)
	clients.RemoveClient(sessionID)
}

func removeWsClient(sessionID uint64) {
	mu.Lock()
	defer mu.Unlock()
	log.Trace("Removing session ID [%d] from WsClients, len: [%d], cap: [%d]", sessionID, len(wsClients), cap(wsClients))
	// delete(wsClients, sessionID)
	for i := 0; i < len(wsClients); i++ {
		if wsClients[i].SessionID == sessionID {
			wsClients = append(wsClients[:i], wsClients[i+1:]...)
			log.Trace("Successfully removed session ID [%d] from WsClients, len: [%d], cap: [%d]", sessionID, len(wsClients), cap(wsClients))
			return
		}
	}
	log.Impossible("Couldnt remove session ID [%d] from WsClients", sessionID)
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
		// log.Println("endIndex:", endIndex)

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
		case addChatMessage: // user sent a chat message on x channel
			log.Debug("User ID [%d] sent a chat message", c.UserID)
			broadcastData, failData := c.onChatMessageRequest(packetJson, packetType)
			if failData != nil {
				c.WriteChan <- failData
			} else {
				broadcastChan <- broadcastData
			}

		case chatHistory: // user entered a channel, requesting chat history
			log.Debug("User ID [%d] is asking for a chat history", c.UserID)
			chatHistoryBytes := c.onChatHistoryRequest(packetJson, packetType)
			if chatHistoryBytes != nil {
				c.WriteChan <- chatHistoryBytes
			}

		case deleteChatMessage: // user deleting a chat message
			log.Debug("User ID [%d] wants to delete a chat message", c.UserID)
			broadcastData, failData := c.onChatMessageDeleteRequest(packetJson, packetType)
			if failData != nil {
				c.WriteChan <- failData
			} else {
				broadcastChan <- broadcastData
			}

		case addServer: // user adding a server
			log.Debug("User ID [%d] wants to create a server", c.UserID)
			c.WriteChan <- c.onAddServerRequest(packetJson)

		case serverList: // user requesting their server list
			log.Debug("User ID [%d] is requesting their joined server list", c.UserID)
			c.WriteChan <- macros.PreparePacket(22, database.GetServerList(c.UserID))

		case deleteServer: // user deleting a server
			log.Debug("User ID [%d] wants to delete a server", c.UserID)
			broadcastChan <- c.onServerDeleteRequest(packetJson, packetType)

		case serverInviteLink: // user requested an invite link for a server
			log.Debug("User ID [%d] is requesting an invite link for a server", c.UserID)
			c.WriteChan <- c.onServerInviteRequest(packetJson)

		case addChannel: // user added a channel to their server
			log.Debug("User ID [%d] wants to add a channel", c.UserID)
			broadcastData, failData := c.onAddChannelRequest(packetJson, packetType)
			if failData != nil {
				c.WriteChan <- failData
			} else {
				broadcastChan <- broadcastData
			}

		case channelList: // user entered a server, requesting channel list
			log.Debug("User ID [%d] is requesting channel list of a server", c.UserID)
			channelListbytes := c.onChannelListRequest(packetJson)
			if channelListbytes != nil {
				c.WriteChan <- channelListbytes
			}
		case addServerMember: // a new user connected to the server

		case serverMemberList: // user entered a server, requesting member list
			c.WriteChan <- c.onServerMemberListRequest(packetJson)

		case deleteServerMember: // a user left a server
			log.Debug("User ID [%d] is requesting to leave from a server", c.UserID)
			// broadcastData, failData := c.onLeaveServerRequest(packetJson)
			// if failData != nil {
			// 	c.WriteChan <- failData
			// } else {
			// 	broadcastChan <- broadcastData
			// 	c.WriteChan <- broadcastData.MessageBytes
			// }
		case updateStatus: // user wants to update their status value
			log.Debug("User ID [%d] is requesting to update their status value", c.UserID)
			c.onUpdateUserStatusValue(packetJson)
		case imageHostAddress:
			log.Debug("User ID [%d] is requesting address of image host server", c.UserID)
			imageHostJson, err := json.Marshal(ImageHost)
			if err != nil {
				log.FatalError(err.Error(), "Error serializing ImageHost [%s]", ImageHost)
			}
			c.WriteChan <- macros.PreparePacket(imageHostAddress, imageHostJson)
		case updateUserData: // user wants to update their account data
			log.Debug("User ID [%d] is requesting to update their account data", c.UserID)
			c.onUpdateUserDataRequest(packetJson)

		default: // if unknown
			log.Hack("User ID [%d] sent invalid packet type: [%d]", c.UserID, packetType)
			c.WriteChan <- macros.RespondFailureReason("Packet type is invalid")
		}
	}
}

func (c *WsClient) writeMessages(wg *sync.WaitGroup) {
	ticker := time.NewTicker(pingPeriod) // client will be pinged in intervals using this

	defer func() { // this will run when writeMessages goroutine returns
		ticker.Stop()
		wg.Done()
	}()

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
			case addChatMessage, deleteChatMessage: // things that only affect a single channel
				for i := 0; i < len(wsClients); i++ {
					channelID := clients.GetCurrentChannelID(wsClients[i].SessionID)
					if checkClient(wsClients[i].SessionID, "", channelID) {
						return
					}
					if channelID == broadcastData.AffectedChannel { // if client is in affected channel
						broadcastLog(broadcastData.Type, wsClients[i].UserID, wsClients[i].SessionID)
						wsClients[i].WriteChan <- broadcastData.MessageBytes
					}
				}
			case addChannel, deleteChannel, addServerMember, deleteServerMember, updateOnline: // things that only affect a single server
				for i := 0; i < len(wsClients); i++ {
					serverID := clients.GetCurrentServerID(wsClients[i].SessionID)
					if checkClient(wsClients[i].SessionID, "", serverID) {
						return
					}
					if serverID == broadcastData.AffectedServers[0] { // if client is currently in that server
						broadcastLog(broadcastData.Type, wsClients[i].UserID, wsClients[i].SessionID)
						wsClients[i].WriteChan <- broadcastData.MessageBytes
					}
				}

			case updateMemberProfilePic, updateStatus, updateMemberDisplayName, deleteServer: // things that affect multiple servers
				for i := 0; i < len(wsClients); i++ {
					for s := 0; s < len(broadcastData.AffectedServers); s++ {
						serverID := clients.GetCurrentServerID(wsClients[i].SessionID)
						if checkClient(wsClients[i].SessionID, "", serverID) {
							return
						}
						if serverID == broadcastData.AffectedServers[s] { // if client is member of any affected server
							broadcastLog(broadcastData.Type, wsClients[i].UserID, wsClients[i].SessionID)
							wsClients[i].WriteChan <- broadcastData.MessageBytes
						}
					}
				}
			case updateUserData, updateUserProfilePic: // things that only affect a single user, sending to all connected devices
				for i := 0; i < len(wsClients); i++ {
					if wsClients[i].UserID == broadcastData.AffectedUserID {
						broadcastLog(broadcastData.Type, wsClients[i].UserID, wsClients[i].SessionID)
						wsClients[i].WriteChan <- broadcastData.MessageBytes
					}
				}
			}
		}
	}
}
