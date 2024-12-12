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

//const sessionTokenLength = 16

const (
	addChatMessage     byte = 1
	chatHistory        byte = 2
	deleteChatMessage  byte = 3
	addServer          byte = 21
	serverList         byte = 22
	deleteServer       byte = 23
	serverInviteLink   byte = 24
	addChannel         byte = 31
	channelList        byte = 32
	deleteChannel      byte = 33
	addServerMember    byte = 41
	serverMemberList   byte = 42
	deleteServerMember byte = 43
	updateUserData     byte = 51
	updateProfilePic   byte = 52
	updateStatus       byte = 53
	updateStatusText   byte = 54
	updateOnline       byte = 55
	imageHostAddress   byte = 242
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
}

type Client struct {
	DisplayName      string
	WsConn           *websocket.Conn
	SessionToken     uint64
	UserID           uint64
	CurrentChannelID uint64
	currentServerID  uint64
	Status           byte
	StatusText       string
	WriteChan        chan []byte
	CloseChan        chan bool
}

type UserData struct {
	UserID      string
	DisplayName string
	ProfilePic  string
}

var broadcastChan = make(chan BroadcastData, 100)

var mu sync.Mutex // used so only 1 goroutine can access the Clients hashmap at one time

var Clients = make(map[uint64]*Client)

func Init() {
	go broadCastChannel()
}

// client is connecting to the websocket
func AcceptWsClient(userID uint64, w http.ResponseWriter, r *http.Request) {
	// session ID is used as key value for Clients hashmap to make it possible
	// for a single user to connect to chat from multiple devices/browsers

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WarnError(err.Error(), "Error upgrading connection of user ID [%d] to websocket protocol", userID)
		return
	}
	username := database.GetUsername(userID)
	if username == "" {
		// no idea why this would happen
		log.Impossible("After accepting websocket client, user ID [%d] has no username set in the database", userID)
		return
	}

	// sending and reading messages are two separate goroutines
	// it's so they cant block each other
	// they communicate using channels

	var sessionToken uint64 = snowflake.Generate()

	client := &Client{
		DisplayName:      username,
		WsConn:           wsConn,
		SessionToken:     sessionToken,
		UserID:           userID,
		CurrentChannelID: 0,
		Status:           1,
		StatusText:       "custom status",
		WriteChan:        make(chan []byte, 10),
		CloseChan:        make(chan bool),
	}

	// add to Clients hashmap
	mu.Lock()
	Clients[sessionToken] = client
	mu.Unlock()

	// create 2 goroutines for reading and writing messages
	var wg sync.WaitGroup
	wg.Add(2)

	go client.readMessages(&wg)
	go client.writeMessages(&wg)

	log.Info("Session ID [%d] as user ID [%d] has connected to websocket", sessionToken, userID)

	// sends the client its own user ID and display name
	displayName, profilePic := database.GetUserData(userID)

	var userData UserData = UserData{
		UserID:      strconv.FormatUint(userID, 10),
		DisplayName: displayName,
		ProfilePic:  profilePic,
	}

	jsonUserID, err := json.Marshal(userData)
	if err != nil {
		macros.ErrorSerializing(err.Error(), "userID", client.UserID)
	}

	client.WriteChan <- macros.PreparePacket(241, jsonUserID)

	//setUserStatusText(client.userID, "Online")
	setUserOnline(client.UserID, true)

	// this will block here while both the reading and writing goroutine are running
	// if one stops, the other should stop too
	wg.Wait()

	//setUserStatusText(client.userID, "Offline")
	setUserOnline(client.UserID, false)

	// close websocket connection
	// if err := wsConn.Close(); err != nil {
	// 	log.WarnError(err.Error(), "Error closing websocket connection for user ID [%d]", userID)
	// }

	// lastly remove the client from hashmap
	mu.Lock()
	delete(Clients, sessionToken)
	mu.Unlock()
	log.Info("Removed session ID [%d] as user ID [%d] from the connected Clients", sessionToken, userID)
}

func (c *Client) CheckIfUserIsOnline() bool {
	for i := range Clients {
		if Clients[i].UserID == c.UserID {
			return true
		}
	}
	return false
}

func GetSessionTokens(userID uint64) []uint64 {
	var sessionTokens []uint64
	for _, client := range Clients {
		sessionTokens = append(sessionTokens, client.SessionToken)
	}
	return sessionTokens
}

func (c *Client) setCurrentChannelID(channelID uint64) {
	c.CurrentChannelID = channelID
	log.Trace("User ID [%d] is now on channel ID [%d]", c.UserID, channelID)
}

func (c *Client) setCurrentServerID(serverID uint64) {
	c.currentServerID = serverID
	log.Trace("User ID [%d] is now on server ID [%d]", c.UserID, c.currentServerID)
}

func (c *Client) readMessages(wg *sync.WaitGroup) {
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
			log.WarnError(err.Error(), "Failed reading message from session ID [%d] as user ID [%d]", c.SessionToken, c.UserID)
			break
		}

		// check if array is at least 5 in length to avoid exceptions
		// because if client sends smaller byte array for some reason,
		// this func would throw an index out of range exception
		// not supposed to happen in normal cases
		if len(receivedBytes) < 5 {
			log.Hack("Session ID [%d] as user ID [%d] sent a byte array shorter than 5 length", c.SessionToken, c.UserID)
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
			c.WriteChan <- c.onChatHistoryRequest(packetJson, packetType)

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
			c.WriteChan <- c.onChannelListRequest(packetJson)

		case addServerMember: // a new user connected to the server

		case serverMemberList: // user entered a server, requesting member list
			log.Debug("User ID [%d] is requesting list of members of server ID [%d]", c.UserID, c.currentServerID)
			c.WriteChan <- c.onServerMemberListRequest(packetJson)

		case deleteServerMember: // a user left a server
			log.Debug("User ID [%d] is requesting to leave from a server", c.UserID)
			broadcastData, failData := c.onLeaveServerRequest(packetJson)
			if failData != nil {
				c.WriteChan <- failData
			} else {
				broadcastChan <- broadcastData
				c.WriteChan <- broadcastData.MessageBytes
			}
		case updateUserData: // user wants to update their account data
			log.Debug("User ID [%d] is requesting to update their account data", c.UserID)
			broadcastData, failData := c.onUpdateUserDataRequest(packetJson)
			if failData != nil {
				c.WriteChan <- failData
			} else {
				broadcastChan <- broadcastData
			}
		case updateStatus: // user wants to update their status value
			log.Debug("User ID [%d] is requesting to update their status value", c.UserID)
			c.onUpdateUserStatusValue(packetJson)
		case updateStatusText:
			log.Debug("User ID [%d] is requesting to update their status text", c.UserID)
		//case onlineStatuses:
		//	log.Debug("User ID [%d] is requesting online statuses of server members", c.userID)
		//	c.writeChan <- c.onMemberOnlineStatusesRequest(packetJson)
		case imageHostAddress:
			log.Debug("User ID [%d] is requesting address of image host server", c.UserID)
			imageHostJson, err := json.Marshal(ImageHost)
			if err != nil {
				log.FatalError(err.Error(), "Error serializing ImageHost [%s]", ImageHost)
			}
			c.WriteChan <- macros.PreparePacket(242, imageHostJson)

		default: // if unknown
			log.Hack("User ID [%d] sent invalid packet type: [%d]", c.UserID, packetType)
			c.WriteChan <- macros.RespondFailureReason("Packet type is invalid")
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
		log.WarnError(errMsg, "Error writing message to session ID [%d] as user ID [%d]", c.SessionToken, c.UserID)
	}

	for {
		select {
		case messageBytes := <-c.WriteChan:
			c.WsConn.SetWriteDeadline(time.Now().Add(timeoutWrite))
			if err := c.WsConn.WriteMessage(websocket.BinaryMessage, messageBytes); err != nil {
				errorWriting(err.Error())
				return
			}
			log.Trace("Wrote to user ID [%d]", c.UserID)
		case <-ticker.C:
			// log.Trace("Pinging:", c.userID)
			c.WsConn.SetWriteDeadline(time.Now().Add(timeoutWrite))
			if err := c.WsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				errorWriting(err.Error())
				return
			}
		case closed := <-c.CloseChan:
			if closed {
				log.Debug("Session ID [%d] as user ID [%d] received a signal to close writeMessages goroutine", c.SessionToken, c.UserID)
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
			case addChatMessage, deleteChatMessage: // chat messages
				for _, client := range Clients {
					if client.CurrentChannelID == broadcastData.AffectedChannel { // if client is in affected channel
						broadcastLog(broadcastData.Type, client.UserID)
						client.WriteChan <- broadcastData.MessageBytes
					}
				}
			//case addServer: // servers
			//	for _, client := range Clients {
			//		broadcastLog(broadcastData.Type, client.userID)
			//		client.writeChan <- broadcastData.MessageBytes
			//	}
			case addChannel, deleteChannel, addServerMember, deleteServerMember, updateOnline:
				for _, client := range Clients {
					if client.currentServerID == broadcastData.AffectedServers[0] { // if client is currently in that server
						broadcastLog(broadcastData.Type, client.UserID)
						client.WriteChan <- broadcastData.MessageBytes
					}
				}
			case updateUserData, updateProfilePic, updateStatus, updateStatusText, deleteServer: // user updating account stuff
				for _, client := range Clients {
					for i := 0; i < len(broadcastData.AffectedServers); i++ {
						if client.currentServerID == broadcastData.AffectedServers[i] { // if client is member of any affected server
							broadcastLog(broadcastData.Type, client.UserID)
							client.WriteChan <- broadcastData.MessageBytes
						}
					}
				}
			}
		}
	}
}
