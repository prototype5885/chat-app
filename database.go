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
		log.Fatal("Error creating database folder:", err)
	}

	var err error
	d.db, err = sql.Open("sqlite", "./database/database.db")
	if err != nil {
		log.Fatal("Error opening sqlite file:", err)
	}

	d.db.SetMaxOpenConns(1)
	d.createTables()
}

func (d *Database) ConnectMariadb(username string, password string, address string, port string, dbName string) {
	log.Println("Opening MySQL/MariaDB database...")

	var err error
	d.db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, address, port, dbName))
	if err != nil {
		log.Fatal("Error opening mariadb connection:", err)
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
		password BINARY(60),
		totp CHAR(32)
	)`)
	if err != nil {
		log.Fatal("Error creating users table in database:", err)
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
		log.Fatal("Error creating server table in database:", err)
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
		log.Fatal("Error creating channels table in database:", err)
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
		log.Fatal("Error creating messages table in database:", err)
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
		log.Fatal("Error creating tokens table in database:", err)
		return
	}
}

func (d *Database) AddChatMessage(messageID uint64, channelID uint64, userID uint64, message string) Result {
	printWithID(userID, "Adding chat message into database...")
	const query string = "INSERT INTO messages (message_id, channel_id, user_id, Message) VALUES (?, ?, ?, ?)"
	_, err := d.db.Exec(query, messageID, channelID, userID, message)
	if err != nil {
		fatalWithID(userID, "Error adding chat message ID ["+strconv.FormatUint(messageID, 10)+"] into database:", err.Error())
	}
	successWithID(userID, "Added chat message ID ["+strconv.FormatUint(messageID, 10)+"] into database")
	return Result{
		Success: true,
		Message: "",
	}
}

func (d *Database) getMessagesFromChannel(channelID uint64) (ServerChatMessages, Result) {
	const query string = "SELECT * FROM messages WHERE channel_id = ?"

	rows, err := d.db.Query(query, channelID)
	if err != nil {
		if err == sql.ErrNoRows { // there is no channel with given id
			return ServerChatMessages{}, Result{
				Success: false,
				Message: fmt.Sprintf("No channel found with given ID: %d\n", channelID),
			}
		}
		fatalWithID(channelID, "Error searching for messages in channel ID", err.Error())
	}

	var messages = ServerChatMessages{}
	for rows.Next() {
		var message ServerChatMessage = ServerChatMessage{
			Username: "test",
		}
		err := rows.Scan(&message.MessageID, &message.ChannelID, &message.UserID, &message.Message)
		if err != nil {
			log.Fatal("Error scanning message row into struct")
		}
		messages.Messages = append(messages.Messages, message)
	}
	successWithID(channelID, "Messages from channel ID were retrieved")
	return messages, Result{
		Success: true,
		Message: "",
	}
}

func (d *Database) AddServer(server Server) Result {
	printWithID(server.ServerID, "Adding server into database...")
	const query string = "INSERT INTO servers (server_id, owner_id, name) VALUES (?, ?, ?)"
	_, err := d.db.Exec(query, server.ServerID, server.ServerOwnerID, server.ServerName)
	if err != nil {
		fatalWithID(server.ServerID, "Error adding server into database:", err.Error())
	}
	successWithID(server.ServerID, "Added server into database")
	return Result{
		Success: true,
		Message: "",
	}
}

func (d *Database) AddChannel(channelID uint64, serverID uint64) Result {
	printWithID(channelID, "Adding channel into database...")
	const query string = "INSERT INTO channels (channel_id, server_id) VALUES (?, ?)"
	_, err := d.db.Exec(query, channelID, serverID)
	if err != nil {
		fatalWithID(channelID, "Error adding channel into database:", err.Error())
	}
	successWithID(channelID, "Added channel into database")
	return Result{
		Success: true,
		Message: "",
	}
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
// 			log.Fatalf("%s: Error executing SELECT query: %s\n", userIDArgs, err)
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
		fatalWithName(username, "Error registering new user into database", err.Error())
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
		fatalWithName(username, "Error getting user ID of username from database", err.Error())
	}
	successWithName(username, "User ID of username was retrieved from database")
	// log.Println(successText("User ID of user [" + username + "] was retrieved from database"))
	return userID, Result{
		Success: true,
		Message: "",
	}
}

func (d *Database) GetUsername(userID uint64) (string, Result) {
	printWithID(userID, "Searching for field [username] in database...")
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
		fatalWithID(userID, "Error getting username of user ID from database", err.Error())
	}
	successWithID(userID, "Username of user ID was retrieved from database")
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
		fatalWithID(userID, "Error getting password of user ID from database", err.Error())
	}
	successWithID(userID, "Password of user ID was retrieved from database")
	return passwordHash, Result{
		Success: true,
		Message: "",
	}
}

func (d *Database) AddToken(token Token) Result {
	printWithID(token.UserID, "Adding new token into database...")

	const query string = "INSERT INTO tokens (token, user_id, expiration) VALUES (?, ?, ?)"

	_, err := d.db.Exec(query, token.Token, token.UserID, token.Expiration)
	if err != nil {
		fatalWithID(token.UserID, "Error adding new token for user ID into database", err.Error())
	}
	successWithID(token.UserID, "Added a new token for user ID into database")
	return Result{
		Success: true,
		Message: "",
	}
}

func (d *Database) GetToken(tokenBytes []byte) (Token, Result) {
	log.Println("Searching for token in database...")

	const query string = "SELECT * FROM tokens WHERE token = ?"

	var token Token
	var text uint64

	err := d.db.QueryRow(query, tokenBytes).Scan(&token.Token, &token.UserID, &text)
	if err != nil {
		if err == sql.ErrNoRows { // token was not found
			return Token{}, Result{
				Success: false,
				Message: "Token was not found in database",
			}
		}
		log.Fatal("Error retrieving given token from database: " + err.Error())
	}
	successWithName(hex.EncodeToString(tokenBytes), "Given token was found in database")
	return token, Result{
		Success: true,
		Message: "",
	}
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
// 			log.Fatalf("%s: Error executing SELECT query: %s\n", userIDArgs, err)
// 			return 0, "", "", "", "", Result{
// 				Success: false,
// 				Message: "FATAL: Error searching for user in database",
// 			}
// 		}
// 	}
// 	return userID, username, passwordHash, totpSecret, activeTokens, Result{
// 		Success: true,
// 		Message: "User row was retrieved successfully",
// 	}
// }
