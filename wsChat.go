package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"proto-chat/modules/snowflake"
	"strconv"
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

// when client sent a chat message
func onChatMessageRequest(jsonBytes []byte, userID uint64, displayName string) []byte {
	type ClientChatMsg struct {
		ChannelID string
		Message   string
	}

	var clientChatMsg ClientChatMsg

	if err := json.Unmarshal(jsonBytes, &clientChatMsg); err != nil {
		log.Printf("Error deserializing onChatMessageRequest json of user ID [%d], reason: %s\n", userID, err.Error())
		return nil
	}

	log.Printf("ChannelID: %s, Msg: %s", clientChatMsg.ChannelID, clientChatMsg.Message)

	// parse channel id string as uint64
	channelID, parseErr := strconv.ParseUint(clientChatMsg.ChannelID, 10, 64)
	if parseErr != nil {
		log.Printf("Error parsing uint64 in onChatMessageRequest from user ID [%d], reason: %s\n", userID, parseErr.Error())
		return nil
	}

	var serverChatMsg = ServerChatMessage{
		MessageID: snowflake.Generate(),
		ChannelID: channelID,
		UserID:    userID,
		Username:  displayName,
		Message:   clientChatMsg.Message,
	}

	database.AddChatMessage(serverChatMsg.MessageID, serverChatMsg.ChannelID, serverChatMsg.UserID, serverChatMsg.Message)

	jsonBytes, err := json.Marshal(serverChatMsg)
	if err != nil {
		log.Panic("Error serializing json at onChatMessage:", err)
	}

	return preparePacket(1, jsonBytes)
}

// when client wants to delete a message they own
func onDeleteChatMessageRequest(jsonBytes []byte, userID uint64) []byte {
	type MessageToDelete struct {
		MessageID string
	}

	var messageToDelete = MessageToDelete{}

	if err := json.Unmarshal(jsonBytes, &messageToDelete); err != nil {
		log.Printf("Error deserializing onDeleteChatMessageRequest json of user ID [%d], reason: %s\n", userID, err.Error())
		return nil
	}

	// parse message ID string as uint64
	messageID, parseErr := strconv.ParseUint(messageToDelete.MessageID, 10, 64)
	if parseErr != nil {
		log.Printf("Error parsing uint64 in onDeleteChatMessageRequest from user ID [%d], reason: %s\n", userID, parseErr.Error())
		return nil
	}

	ownerID, dbSuccess := database.GetChatMessageOwner(messageID)
	if !dbSuccess {
		return nil
	}

	if ownerID != userID {
		log.Printf("User ID [%d] is trying to delete someone else's message [%d], aborting\n", userID, messageID)
		return nil
	}

	database.DeleteChatMessage(messageID)

	messagesBytes, err := json.Marshal(messageToDelete)
	if err != nil {
		log.Panicf("Error serializing json at onDeleteChatMessageRequest for user ID [%d], reason: %s\n:", userID, err.Error())
	}
	return preparePacket(3, messagesBytes)
}

// when client is requesting chat history for a channel
func onChatHistoryRequest(packetJson []byte, userID uint64) []byte {
	type ChatHistoryRequest struct {
		ChannelID string
	}

	var chatHistoryRequest ChatHistoryRequest

	if err := json.Unmarshal(packetJson, &chatHistoryRequest); err != nil {
		log.Printf("Error deserializing onChatHistoryRequest json of user ID [%d], reason: %s\n", userID, err.Error())
		return nil
	}

	// parse channel id string as uint64
	channelID, parseErr := strconv.ParseUint(chatHistoryRequest.ChannelID, 10, 64)
	if parseErr != nil {
		printWithID(userID, "Error parsing uint64 in onChatHistoryRequest:"+parseErr.Error())
		return nil
	}

	type ServerChatMessages struct {
		Messages []ServerChatMessage
	}

	var messages = ServerChatMessages{
		Messages: database.GetMessagesFromChannel(channelID),
	}

	messagesBytes, err := json.Marshal(messages)
	if err != nil {
		log.Panicf("Error serializing json at onChatHistoryRequest for user ID [%d], reason: %s\n:", userID, err.Error())
	}
	return preparePacket(2, messagesBytes)
}

// when client is requesting to add a new server
func onAddServerRequest(packetJson []byte, userID uint64) []byte {
	type AddServerRequest struct {
		Name string
	}

	var addServerRequest = AddServerRequest{}

	if err := json.Unmarshal(packetJson, &addServerRequest); err != nil {
		log.Printf("Error deserializing addServerRequest json of user ID [%d], reason: %s\n", userID, err.Error())
		return nil
	}

	var server = Server{
		ServerID: snowflake.Generate(),
		OwnerID:  userID,
		Name:     addServerRequest.Name,
		Picture:  "nothing.jpg",
	}

	database.AddServer(server)

	messagesBytes, err := json.Marshal(server)
	if err != nil {
		log.Panicf("Error serializing json at onAddServerRequest for user ID [%d], reason: %s\n:", userID, err.Error())
	}
	return preparePacket(21, messagesBytes)
}

func onServerListRequest(userID uint64) []byte {
	type ServersForClient struct {
		Servers []ServerForClient
	}

	var servers = ServersForClient{
		Servers: database.GetServerList(userID),
	}

	messagesBytes, err := json.Marshal(servers)
	if err != nil {
		log.Panicf("Error serializing json at onServerListRequest for user ID [%d], reason: %s\n:", userID, err.Error())
	}
	return preparePacket(22, messagesBytes)
}
