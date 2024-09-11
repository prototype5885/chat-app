package main

import (
	"encoding/json"
	"log"
	"proto-chat/modules/snowflake"
	"strconv"
	"strings"
)

func onChatMessage(jsonBytes []byte, userID uint64, displayName string) []byte {
	type ClientChatMsg struct {
		ChanID string
		Msg    string
	}

	var clientChatMsg ClientChatMsg

	if jsonErr := json.Unmarshal(jsonBytes, &clientChatMsg); jsonErr != nil {
		log.Println("Error deserializing Msg json:", jsonErr)
		return nil
	}

	if strings.HasPrefix(clientChatMsg.Msg, "/getmessages") {
		return nil
	}

	//log.Println(clientChatMsg.ChannelId)
	//log.Println(clientChatMsg.ChatMsg)

	// parse channel id string as uint64
	chanID, parseErr := strconv.ParseUint(clientChatMsg.ChanID, 10, 64)
	if parseErr != nil {
		printWithID(userID, "Error parsing:"+parseErr.Error())
		return nil
	}

	var serverChatMsg = ServerChatMessage{
		MessageID: snowflake.Generate(),
		ChannelID: chanID,
		UserID:    userID,
		Username:  displayName,
		Message:   clientChatMsg.Msg,
	}

	result := database.AddChatMessage(serverChatMsg.MessageID, serverChatMsg.ChannelID, serverChatMsg.UserID, serverChatMsg.Message)
	if !result.Success {
		// there is fatal already in database func
	}

	jsonBytes, err := json.Marshal(serverChatMsg)
	if err != nil {
		log.Fatal("Error serializing json at onChatMessage:", err)
	}

	return preparePacket(1, jsonBytes)
}

func onChatHistoryRequest(packetJson []byte, userID uint64) []byte {
	type ChatHistoryRequest struct {
		ChannelID uint64
	}

	var chatHistoryRequest ChatHistoryRequest

	if jsonErr := json.Unmarshal(packetJson, &chatHistoryRequest); jsonErr != nil {
		log.Println("Error deserializing chatHistoryRequest json:", jsonErr)
		return nil
	}

	messages, result := database.getMessagesFromChannel(chatHistoryRequest.ChannelID)
	if !result.Success {
		log.Println(result.Message)
		return nil
	}
	messagesBytes, err := json.Marshal(messages)
	if err != nil {
		log.Fatalln("Error serializing json at onChatHistoryRequest:", err)
	}
	return preparePacket(11, messagesBytes)
}

func onAddServerRequest(packetJson []byte, userID uint64) []byte {
	type AddServerRequest struct {
		Name string
	}

	var addServerRequest = AddServerRequest{}

	if jsonErr := json.Unmarshal(packetJson, &addServerRequest); jsonErr != nil {
		log.Println("Error deserializing addServerRequest json:", jsonErr)
		return nil
	}

	var server = Server{
		ServerID:      snowflake.Generate(),
		ServerOwnerID: userID,
		ServerName:    addServerRequest.Name,
	}

	if result := database.AddServer(server); !result.Success {
		log.Println(result.Message)
		return nil
	}

	messagesBytes, err := json.Marshal(server)
	if err != nil {
		log.Fatalln("Error serializing json at onAddServerRequest:", err)
	}
	return preparePacket(21, messagesBytes)
}
