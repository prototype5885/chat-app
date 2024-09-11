// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/binary"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second // Time allowed to write a Message to the peer.
	pongWait       = 60 * time.Second // Time allowed to read the next pong Message from the peer.
	pingPeriod     = 15 * time.Second // Send pings to peer with this period. Must be less than pongWait.
	maxMessageSize = 8192             // Maximum Message size allowed from peer.
)

// var (
// 	newline = []byte{'\n'}
// 	space   = []byte{' '}
// )

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	userID uint64

	displayName string

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		log.Println("Closing connection")
		c.hub.unregister <- c
		c.conn.Close()
	}()
	// c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, receivedBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println(err)
			}
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

		var responseBytes []byte
		switch packetType {
		case 1: // client sent a chat message
			responseBytes = onChatMessage(packetJson, c.userID, c.displayName)
		case 11: // chat history request
			responseBytes = onChatHistoryRequest(packetJson, c.userID)
		case 21: // client added a server
			responseBytes = onAddServerRequest(packetJson, c.userID)
		}

		if responseBytes == nil {
			printWithID(c.userID, "User sent a websocket message with unprocessable packet type")
			return
		}

		var responseEndIndex uint32 = binary.LittleEndian.Uint32(responseBytes[:4])
		var responsePacketType byte = responseBytes[4]
		var responsePacketJson []byte = responseBytes[5:endIndex]

		log.Println("prepared packet:", responseEndIndex, responsePacketType, string(responsePacketJson))

		// printWithID(c.userID, string(responseBytes))
		c.hub.broadcast <- responseBytes
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				log.Println("The hub closed the channel")
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				log.Println(err)
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket Message.
			log.Println("Add queued chat messages to the current websocket Message")
			n := len(c.send)
			for i := 0; i < n; i++ {
				// w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				log.Println(err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			// printWithID(c.userID, "Pinging")
			err := c.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func acceptWsClient(userID uint64, hub *Hub, w http.ResponseWriter, r *http.Request) {
	printWithID(userID, "Accepting websocket connection")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error accepting websocket connection: ", err)
		return
	}
	username, nameResult := database.GetUsername(userID)
	if !nameResult.Success { // this is not supposed to happen
		fatalWithID(userID, "For some reason no username was associated with the given user id", "FATAL")
	}

	client := &Client{
		hub:         hub,
		userID:      userID,
		displayName: username,
		conn:        conn,
		send:        make(chan []byte, 256),
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
	printWithID(userID, "Client has connected to the websocket")
}
