package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

// type User struct {
// 	userID       uint64
// 	username     string
// 	passwordHash string
// 	totpSecret   string
// 	activeTokens string
// }

type Database struct {
	db *sql.DB
}

var database Database = Database{}

func (d *Database) ConnectSqlite() {
	log.Println("Opening sqlite database...")

	//os.Remove("./database/database.db")

	if err := os.MkdirAll("database", os.ModePerm); err != nil {
		log.Panic("Error creating sqlite database folder:", err)
	}

	var err error
	d.db, err = sql.Open("sqlite", "./database/database.db")
	if err != nil {
		log.Panic("Error opening sqlite file:", err)
	}

	d.db.SetMaxOpenConns(1)
	d.createTables()
}

func (d *Database) ConnectMariadb(username string, password string, address string, port string, dbName string) {
	log.Println("Opening MySQL/MariaDB database...")

	var err error
	d.db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, address, port, dbName))
	if err != nil {
		log.Println(err.Error())
		log.Panic("Error opening mariadb connection")
	}

	d.db.SetMaxOpenConns(100)
	d.createTables()
}

func (d *Database) createTables() {
	// users table
	var err error
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS users (
		user_id BIGINT UNSIGNED PRIMARY KEY,
		username TEXT,
		picture TEXT,
		password BINARY(60),
		totp CHAR(32)
	)`)
	if err != nil {
		log.Panic("Error creating users table in database:", err)
	}

	// servers table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS servers (
			server_id BIGINT UNSIGNED PRIMARY KEY,
			owner_id BIGINT UNSIGNED,
			name TEXT,
			picture TEXT,
			FOREIGN KEY (owner_id) REFERENCES users(user_id)
		)`)
	if err != nil {
		log.Println(err.Error())
		log.Panic("Error creating server table in database")
	}

	// channels table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS channels (
			channel_id BIGINT UNSIGNED PRIMARY KEY,
			server_id BIGINT UNSIGNED,
			name TEXT,
			FOREIGN KEY (server_id) REFERENCES servers(server_id)
		)`)
	if err != nil {
		log.Println(err.Error())
		log.Panic("Error creating channels table in database")
	}

	// messages table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		message_id BIGINT UNSIGNED PRIMARY KEY,
		channel_id BIGINT UNSIGNED,
		user_id BIGINT UNSIGNED,
		message TEXT,
		FOREIGN KEY (channel_id) REFERENCES channels(channel_id)
	)`)
	if err != nil {
		log.Println(err.Error())
		log.Panic("Error creating messages table in database")
	}

	// tokens table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS tokens (
		token BINARY(128) PRIMARY KEY,
		user_id BIGINT UNSIGNED,
		expiration BIGINT UNSIGNED,
		FOREIGN KEY (user_id) REFERENCES users(user_id)
	)`)
	if err != nil {
		log.Println(err.Error())
		log.Panic("Error creating tokens table in database")
	}

	// profile pictures table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS profilepics (
			hash BINARY(32) PRIMARY KEY,
			file_name TEXT
		)`)
	if err != nil {
		log.Println(err.Error())
		log.Panic("Error creating profilepics table in database")
	}
}

func (d *Database) AddChatMessage(messageID uint64, channelID uint64, userID uint64, message string) string {
	log.Printf("Adding chat message ID [%d] from user ID [%d] into database...\n", messageID, userID)
	const query string = "INSERT INTO messages (message_id, channel_id, user_id, Message) VALUES (?, ?, ?, ?)"
	_, err := d.db.Exec(query, messageID, channelID, userID, message)
	if err != nil {
		log.Println(err.Error())
		if strings.Contains(err.Error(), "Error 1452") {
			log.Printf("HACK: Failed adding message ID [%d] into database for channel ID [%d] from user ID [%d], there is no channel with given ID", messageID, channelID, userID)
			return fmt.Sprintf("Can't add message ID [%d] to channel ID [%d] because channel doesn't exist", messageID, channelID)
		}
		log.Println(err.Error())
		log.Panicf("Error adding chat message ID [%d] from user ID [%d] into database\n", messageID, userID)
	}
	log.Printf("Chat message ID [%d] from user ID [%d] has been added into database successfully\n", messageID, userID)
	return ""
}

func (d *Database) GetChatMessageOwner(messageID uint64) (uint64, bool) {
	log.Printf("Searching for owner of message ID [%d]...\n", messageID)
	const query string = "SELECT user_id FROM messages WHERE message_id = ?"
	var userID uint64
	err := d.db.QueryRow(query, messageID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows { // there is no user with this name
			log.Printf("No message found in messages with ID %d\n", messageID)
			return 0, false
		}
		log.Println(err.Error())
		log.Panicf("Error getting user ID of the owner of message ID [%d]\n", messageID)
	}
	log.Printf("Owner ID of message ID [%d] was confirmed to be: [%d]\n", messageID, userID)
	return userID, true
}

func (d *Database) DeleteChatMessage(messageID uint64) bool {
	log.Printf("Deleting message ID [%d] from db table [messages]...", messageID)

	stmt, err := d.db.Prepare("DELETE FROM messages where message_id = ?")
	if err != nil {
		log.Println(err.Error())
		log.Panicf("Error preparing statement in DeleteChatMessage for message ID [%d]\n", messageID)
	}
	defer stmt.Close()

	result, err := stmt.Exec(messageID)
	if err != nil {
		log.Println(err.Error())
		log.Panicf("Error executing statement in DeleteChatMessage for message ID [%d]\n", messageID)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println(err.Error())
		log.Panicf("Error getting rowsAffected in DeleteChatMessage\n")
	}

	if rowsAffected == 0 {
		log.Printf("Message ID [%d] that was to be deleted was nowhere to be found\n", messageID)
		return false
	} else if rowsAffected != 0 && rowsAffected != 1 {
		// this is not supposed to happen at all since it's not possible to have 2 messages with same ID
		log.Panicf("Multiple messages with same ID [%d] were found and deleted\n", messageID)
	}
	return true
}

func (d *Database) GetMessagesFromChannel(channelID uint64) []ServerChatMessage {
	const query string = "SELECT * FROM messages WHERE channel_id = ?"

	rows, err := d.db.Query(query, channelID)
	if err != nil {
		log.Println(err.Error())
		log.Panicf("Error searching for messages on channel ID [%d]\n", channelID)
	}

	var messages []ServerChatMessage

	var counter int = 0
	for rows.Next() {
		counter++
		var message = ServerChatMessage{
			Username: "test",
		}
		err := rows.Scan(&message.MessageID, &message.ChannelID, &message.UserID, &message.Message)
		if err != nil {
			log.Println(err.Error())
			log.Panicf("Error scanning message row into struct in channel ID [%d]\n:", channelID)
		}
		messages = append(messages, message)
	}

	if counter == 0 {
		log.Printf("No messages found on channel ID: [%d]\n", channelID)
		return messages
	}

	log.Printf("Messages from channel ID [%d] were retrieved successfully\n", channelID)
	return messages
}

func (d *Database) AddServer(serverID uint64, ownerID uint64, serverName string, picture string) {
	log.Printf("Adding server ID [%d] of owner ID [%d] into database...\n", serverID, ownerID)
	const query string = "INSERT INTO servers (server_id, owner_id, name, picture) VALUES (?, ?, ?, ?)"
	_, err := d.db.Exec(query, serverID, ownerID, serverName, picture)
	if err != nil {
		log.Println(err.Error())
		log.Panicf("Error adding server ID [%d] into database\n", serverID)
	}
	log.Printf("Successfully added server ID [%d] into database\n", serverID)
}

func (d *Database) GetServerList(userID uint64) []ServerResponse {
	log.Printf("Getting server list of user ID [%d]...\n", userID)
	const query string = "SELECT server_id, name, picture FROM servers"

	rows, err := d.db.Query(query)
	if err != nil {
		log.Println(err.Error())
		log.Panicf("Error searching for server list of user ID [%d]\n", userID)
	}

	var servers []ServerResponse

	var counter int = 0
	for rows.Next() {
		counter++
		var server = ServerResponse{}
		err := rows.Scan(&server.ServerID, &server.Name, &server.Picture)
		if err != nil {
			log.Println(err.Error())
			log.Panicf("Error scanning server row into struct for user ID [%d]\n:", userID)
		}
		servers = append(servers, server)
	}

	if counter == 0 {
		log.Printf("User ID [%d] is not in any servers\n", userID)
		return servers
	}

	log.Printf("Servers for user ID [%d] were retrieved successfully\n", userID)
	return servers
}

func (d *Database) AddChannel(channelID uint64, serverID uint64, channelName string) bool {
	log.Printf("Adding channel ID [%d] into database...\n", channelID)
	const query string = "INSERT INTO channels (channel_id, server_id, name) VALUES (?, ?, ?)"
	_, err := d.db.Exec(query, channelID, serverID, channelName)
	if err != nil {
		log.Println(err.Error())
		if strings.Contains(err.Error(), "Error 1452") {
			log.Printf("Failed adding channel ID [%d] into database for server ID [%d], there is no server with given ID", channelID, serverID)
			return false
		}
		log.Println(err.Error())
		log.Panicf("Error adding channel ID [%d] into database\n", channelID)
	}
	log.Printf("Successfully added channel ID [%d] into database\n", channelID)
	return true
}

func (d *Database) GetChannelList(serverID uint64) []ChannelResponse {
	log.Printf("Getting channel list of server ID [%d]...\n", serverID)
	const query string = "SELECT channel_id, name FROM channels WHERE server_id = ?"

	rows, err := d.db.Query(query, serverID)
	if err != nil {
		log.Println(err.Error())
		log.Panicf("Error searching for channels list of server ID [%d]\n", serverID)
	}

	var channels []ChannelResponse

	var counter int = 0
	for rows.Next() {
		counter++
		var channel = ChannelResponse{}
		err := rows.Scan(&channel.ChannelID, &channel.Name)
		if err != nil {
			log.Println(err.Error())
			log.Panicf("Error scanning channel row into struct from server ID [%d]\n:", serverID)
		}
		channels = append(channels, channel)
	}

	if counter == 0 {
		log.Printf("Server ID [%d] doesn't have any channels\n", serverID)
		return channels
	}

	log.Printf("Channels from server ID [%d] were retrieved successfully\n", serverID)
	return channels
}

func (d *Database) RegisterNewUser(userId uint64, username string, passwordHash []byte, totpSecret string) bool {
	log.Printf("Registering new username [%s] into database...\n", username)
	const query string = "INSERT INTO users (user_id, username, password, totp) VALUES (?, ?, ?, ?)"
	_, err := d.db.Exec(query, userId, username, passwordHash, totpSecret)
	if err != nil {
		log.Println(err.Error())
		log.Panicf("Error registering username [%s] into database\n", username)
	}
	log.Printf("Username [%s] was registered into database successfully with id [%d]\n", username, userId)
	return true
}

// func (d *Database) GetUserID(username string) uint64 {
// 	printWithName(username, "Searching for field [user_id] in database...")
// 	const query string = "SELECT user_id FROM users WHERE username = ?"
// 	var userID uint64
// 	err := d.db.QueryRow(query, username).Scan(&userID)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			log.Println("No user was found with username [%s]\n", username)
// 			return 0
// 		}
// 		log.Println(err.Error())
// 		log.Panicf("Error getting user ID of username [%s] from database\n", username)
// 	}
// 	log.Println("User ID of username [%s] was retrieved from database successfully\n", username)
// 	return userID
// }

func (d *Database) GetUsername(userID uint64) string {
	log.Printf("Searching for field [username] in database using user ID [%d]...", userID)
	const query string = "SELECT username FROM users WHERE user_id = ?"
	var userName string
	err := d.db.QueryRow(query, userID).Scan(&userName)
	if err != nil {
		log.Println(err.Error())
		if err == sql.ErrNoRows { // there is no user with this id
			log.Printf("No user was found with user ID [%d]\n", userID)
			return ""
		}
		log.Println(err.Error())
		log.Panicf("Error getting username of user ID [%d] from database\n", userID)
	}
	log.Printf("Username of user ID [%d] was retrieved from database successfully\n", userID)
	return userName
}

func (d *Database) GetPasswordAndID(username string) ([]byte, uint64) {
	log.Printf("Searching for password of user [%s] in database...\n", username)
	const query string = "SELECT password, user_id FROM users WHERE username = ?"
	var passwordHash []byte
	var userID uint64
	err := d.db.QueryRow(query, username).Scan(&passwordHash, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No user was found with user [%s]\n", username)
			return nil, 0
		}
		log.Println(err.Error())
		log.Panicf("Error getting password of user [%s] from database\n", username)
	}
	log.Printf("Password of user  [%s] was retreived from database successfully\n", username)
	return passwordHash, userID
}

func (d *Database) AddToken(token Token) {
	log.Printf("Adding new token of user ID [%d] into database...\n", token.UserID)
	const query string = "INSERT INTO tokens (token, user_id, expiration) VALUES (?, ?, ?)"
	_, err := d.db.Exec(query, token.Token, token.UserID, token.Expiration)
	if err != nil {
		log.Println(err.Error())
		log.Panicf("Error adding new token for user ID [%d] into database\n", token.UserID)
	}
	log.Printf("added a new token for user ID [%d] into database\n", token.UserID)
}

func (d *Database) GetToken(tokenBytes []byte) (Token, bool) {
	log.Println("Searching for token in database...")

	const query string = "SELECT * FROM tokens WHERE token = ?"

	var token Token
	var text uint64

	err := d.db.QueryRow(query, tokenBytes).Scan(&token.Token, &token.UserID, &text)
	if err != nil {
		if err == sql.ErrNoRows { // token was not found
			log.Printf("Given token was not found in database: [%s]\n", hex.EncodeToString(tokenBytes))
			return Token{}, false
		}
		log.Panic("Error retrieving given token from database: " + err.Error())
	}
	log.Printf("Given token was successfully found in database: [%s]\n", hex.EncodeToString(tokenBytes))
	return token, true
}

// func getUserRowDB(userIDArgs uint64) (uint64, string, string, string, string, Result) {
// 	const query string = "SELECT * FROM users WHERE user_id = ?"

// 	var (
// 		userID       uint64
// 		username     string
// 		passwordHash string
// 		totpSecret   string
// 		activeTokens string
// 	)

// 	err := db.QueryRow(query, userIDArgs).Scan(&userID, &username, &passwordHash, &totpSecret, &activeTokens)
// 	if err != nil {
// 		if err == sql.ErrNoRows { // there is no user with this name
// 			return 0, "", "", "", "", Result{
// 				Success: false,
// 				Message: noUserIdFoundText(userIDArgs),
// 			}
// 		} else {
// 			log.Panicf("%s: Error executing SELECT query: %s\n", userIDArgs, err)
// 			return 0, "", "", "", "", Result{
// 				Success: false,
// 				Message: "PANIC: Error searching for user in database",
// 			}
// 		}
// 	}
// 	return userID, username, passwordHash, totpSecret, activeTokens, Result{
// 		Success: true,
// 		Message: "User row was retrieved successfully",
// 	}
// }
