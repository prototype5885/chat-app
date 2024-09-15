package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"

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
		log.Panic("Error opening mariadb connection:", err)
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
		profilepic TEXT,
		password BINARY(60),
		totp CHAR(32)
	)`)
	if err != nil {
		log.Panic("Error creating users table in database:", err)
		return
	}

	// servers table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS servers (
			server_id BIGINT UNSIGNED PRIMARY KEY,
			owner_id BIGINT UNSIGNED,
			name TEXT,
			FOREIGN KEY (owner_id) REFERENCES users(user_id)
		)`)
	if err != nil {
		log.Panic("Error creating server table in database:", err)
		return
	}

	// channels table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS channels (
			channel_id BIGINT UNSIGNED PRIMARY KEY,
			server_id BIGINT UNSIGNED,
			name TEXT,
			FOREIGN KEY (server_id) REFERENCES servers(server_id)
		)`)
	if err != nil {
		log.Panic("Error creating channels table in database:", err)
		return
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
		log.Panic("Error creating messages table in database:", err)
		return
	}

	// tokens table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS tokens (
		token BINARY(32) PRIMARY KEY,
		user_id BIGINT UNSIGNED,
		expiration BIGINT UNSIGNED,
		FOREIGN KEY (user_id) REFERENCES users(user_id)
	)`)
	if err != nil {
		log.Panic("Error creating tokens table in database:", err)
		return
	}

	// profile pictures table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS profilepics (
			hash BINARY(32) PRIMARY KEY,
			file_name TEXT
		)`)
	if err != nil {
		log.Panic("Error creating profilepics table in database:", err)
		return
	}
}

func (d *Database) AddChatMessage(messageID uint64, channelID uint64, userID uint64, message string) {
	printWithID(userID, "Adding chat message into database...")
	const query string = "INSERT INTO messages (message_id, channel_id, user_id, Message) VALUES (?, ?, ?, ?)"
	_, err := d.db.Exec(query, messageID, channelID, userID, message)
	if err != nil {
		panicWithID(userID, "Error adding chat message ID ["+strconv.FormatUint(messageID, 10)+"] into database:", err.Error())
	}
	successWithID(userID, "Added chat message ID ["+strconv.FormatUint(messageID, 10)+"] into database")
}

func (d *Database) GetChatMessageOwner(messageID uint64) (uint64, bool) {
	log.Printf("Searching for field [user_id] in db table [messages] with message ID [%d]...", messageID)
	const query string = "SELECT user_id FROM messages WHERE message_id = ?"
	var userID uint64
	err := d.db.QueryRow(query, messageID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows { // there is no user with this name
			log.Printf("No message found in messages with ID %d\n", messageID)
			return 0, false
		}
		log.Panicf("Error getting user ID of the owner of message ID [%d]:%s", messageID, err.Error())
	}
	log.Printf("Owner ID of message ID [%d] was confirmed to be: [%d]\n", messageID, userID)
	return userID, true
}

func (d *Database) DeleteChatMessage(messageID uint64) {
	log.Printf("Deleting message ID [%d] from db table [messages]...", messageID)

	stmt, err := d.db.Prepare("DELETE FROM messages where message_id = ?")
	if err != nil {
		log.Panicf("Error preparing statement in DeleteChatMessage for message ID [%d]\n", messageID)
	}
	defer stmt.Close()

	result, err := stmt.Exec(messageID)
	if err != nil {
		log.Panicf("Error executing statement in DeleteChatMessage for message ID [%d]\n", messageID)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Panicf("Error getting rowsAffected: %s\n", err.Error())
	}

	if rowsAffected == 0 {
		log.Panicf("Message ID [%d] that was to be deleted was nowhere to be found\n", messageID)
	} else if rowsAffected != 0 && rowsAffected != 1 {
		log.Panicf("Multiple messages with same ID [%d] were found and deleted\n", messageID)
	}
}

func (d *Database) GetMessagesFromChannel(channelID uint64) (ServerChatMessages, bool) {
	const query string = "SELECT * FROM messages WHERE channel_id = ?"

	rows, err := d.db.Query(query, channelID)
	if err != nil {
		// if err == sql.ErrNoRows { // there is no channel with given id
		// 	return ServerChatMessages{}, Result{
		// 		Success: false,
		// 		Message: fmt.Sprintf("No messages found on channel ID: %d\n", channelID),
		// 	}
		// }
		panicWithID(channelID, "Error searching for messages in channel ID", err.Error())
	}

	var messages = ServerChatMessages{}

	var counter int = 0
	for rows.Next() {
		counter++
		var message ServerChatMessage = ServerChatMessage{
			Username: "test",
		}
		err := rows.Scan(&message.MessageID, &message.ChannelID, &message.UserID, &message.Message)
		if err != nil {
			log.Panic("Error scanning message row into struct")
		}
		messages.Messages = append(messages.Messages, message)
	}

	if counter == 0 {
		log.Printf("No messages found on channel ID: [%d]\n", channelID)
		return messages, false
	}

	successWithID(channelID, "Messages from channel ID were retrieved")
	return messages, true
}

func (d *Database) AddServer(server Server) {
	printWithID(server.ServerID, "Adding server into database...")
	const query string = "INSERT INTO servers (server_id, owner_id, name) VALUES (?, ?, ?)"
	_, err := d.db.Exec(query, server.ServerID, server.ServerOwnerID, server.ServerName)
	if err != nil {
		panicWithID(server.ServerID, "Error adding server into database:", err.Error())
	}
	successWithID(server.ServerID, "Added server into database")
}

func (d *Database) GetServerList(server Server, userID uint64) {

}

func (d *Database) AddChannel(channelID uint64, serverID uint64) {
	printWithID(channelID, "Adding channel into database...")
	const query string = "INSERT INTO channels (channel_id, server_id) VALUES (?, ?)"
	_, err := d.db.Exec(query, channelID, serverID)
	if err != nil {
		panicWithID(channelID, "Error adding channel into database:", err.Error())
	}
	successWithID(channelID, "Added channel into database")
}

// func (d *Database) getChannel(channelID uint64) (Channel, Result) {
// 	const query string = "SELECT * FROM messages WHERE channel_id = ?"

// 	var channel = Channel{}

// 	err := d.db.QueryRow(query, channelID).Scan(&channel.channelID, &channel.serverID)
// 	if err != nil {
// 		if err == sql.ErrNoRows { // there is no user with this name
// 			return channel, Result{
// 				Success: false,
// 				Message: noUserIdFoundText(userIDArgs),
// 			}
// 		} else {
// 			log.Panicf("%s: Error executing SELECT query: %s\n", userIDArgs, err)
// 			return channel, Result{
// 				Success: false,
// 				Message: "FATAL: Error searching for channel in database",
// 			}
// 		}
// 	}
// 	successWithID(channelID, "Channel was retrieved from database")
// 	return channel, Result{
// 		Success: true,
// 		Message: "",
// 	}
// }

func (d *Database) RegisterNewUser(userId uint64, username string, passwordHash []byte, totpSecret string) Result {
	printWithName(username, "Registering new user into database...")
	const query string = "INSERT INTO users (user_id, username, password, totp) VALUES (?, ?, ?, ?)"
	_, err := d.db.Exec(query, userId, username, passwordHash, totpSecret)
	if err != nil {
		panicWithName(username, "Error registering new user into database", err.Error())
	}
	successWithName(username, "Registered user into database")
	return Result{
		Success: true,
		Message: "",
	}
}

func (d *Database) GetUserID(username string) (uint64, Result) {
	printWithName(username, "Searching for field [user_id] in database...")
	const query string = "SELECT user_id FROM users WHERE username = ?"
	var userID uint64
	err := d.db.QueryRow(query, username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows { // there is no user with this name
			return 0, Result{
				Success: false,
				Message: noUsernameFoundText(username),
			}
		}
		panicWithName(username, "Error getting user ID of username from database", err.Error())
	}
	successWithName(username, "User ID of username was retrieved from database")
	// log.Println(successText("User ID of user [" + username + "] was retrieved from database"))
	return userID, Result{
		Success: true,
		Message: "",
	}
}

func (d *Database) GetUsername(userID uint64) (string, Result) {
	log.Printf("Searching for field [username] in database using user ID [%d]...", userID)
	const query string = "SELECT username FROM users WHERE user_id = ?"
	var userName string
	err := d.db.QueryRow(query, userID).Scan(&userName)
	if err != nil {
		if err == sql.ErrNoRows { // there is no user with this id
			return "", Result{
				Success: false,
				Message: noUserIdFoundText(userID),
			}
		}
		log.Panicf("Error getting username of user ID [%d] from database, reason: %s\n", userID, err.Error())
	}
	log.Printf("Username of user ID [%d] was retrieved from database successfully", userID)
	return userName, Result{
		Success: true,
		Message: "",
	}
}

func (d *Database) GetPassword(userID uint64) ([]byte, Result) {
	printWithID(userID, "Searching for field [password] in database...")

	const query string = "SELECT password FROM users WHERE user_id = ?"

	var passwordHash []byte
	err := d.db.QueryRow(query, userID).Scan(&passwordHash)
	if err != nil {
		if err == sql.ErrNoRows { // there is no user with this id
			return nil, Result{
				Success: false,
				Message: noUserIdFoundText(userID),
			}
		}
		panicWithID(userID, "Error getting password of user ID from database", err.Error())
	}
	successWithID(userID, "Password of user ID was retrieved from database")
	return passwordHash, Result{
		Success: true,
		Message: "",
	}
}

func (d *Database) AddToken(token Token) {
	printWithID(token.UserID, "Adding new token into database...")

	const query string = "INSERT INTO tokens (token, user_id, expiration) VALUES (?, ?, ?)"

	_, err := d.db.Exec(query, token.Token, token.UserID, token.Expiration)
	if err != nil {
		panicWithID(token.UserID, "Error adding new token for user ID into database", err.Error())
	}
	successWithID(token.UserID, "Added a new token for user ID into database")

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
