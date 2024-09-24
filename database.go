package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	log "proto-chat/modules/logging"
	"strings"
	"time"

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
	log.Info("Opening sqlite database...")

	//os.Remove("./database/database.db")

	if err := os.MkdirAll("database", os.ModePerm); err != nil {
		log.FatalError(err.Error(), "Error creating sqlite database folder")
	}

	var err error
	d.db, err = sql.Open("sqlite", "./database/database.db")
	if err != nil {
		log.FatalError(err.Error(), "Error opening sqlite file")
	}

	d.db.SetMaxOpenConns(1)
	d.createTables()
}

func (d *Database) ConnectMariadb(username string, password string, address string, port string, dbName string) {
	log.Info("Opening MySQL/MariaDB database...")

	var err error
	d.db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, address, port, dbName))
	if err != nil {
		log.FatalError(err.Error(), "Error opening mariadb connection")
	}

	d.db.SetMaxOpenConns(100)
	d.createTables()
}

func (d *Database) CloseDatabaseConnection() error {
	err := d.db.Close()
	fmt.Println("Closed main db connection...")
	return err
}

func (d *Database) createTables() {
	errorCreatingTable := func(s string, err error) {
		log.FatalError(err.Error(), "Error creating [%s] table in database", s)
	}

	var err error

	// snoflake IDs table
	// _, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS snowflakes (
	// 	id BIGINT UNSIGNED PRIMARY KEY
	// )`)
	// if err != nil {
	// 	errorCreatingTable("snowflakes", err)
	// }

	// users table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS users (
		user_id BIGINT UNSIGNED PRIMARY KEY,
		username VARCHAR(32) NOT NULL,
		display_name VARCHAR(64) NOT NULL,
		picture VARCHAR(255),
		password BINARY(60) NOT NULL,
		totp CHAR(32),
		UNIQUE(username)
	)`)
	if err != nil {
		errorCreatingTable("users", err)
	}

	// servers table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS servers (
			server_id BIGINT UNSIGNED PRIMARY KEY,
			owner_id BIGINT UNSIGNED,
			name TEXT,
			picture TEXT,
			FOREIGN KEY (owner_id) REFERENCES users(user_id) ON DELETE CASCADE
		)`)
	if err != nil {
		errorCreatingTable("servers", err)
	}

	// channels table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS channels (
			channel_id BIGINT UNSIGNED PRIMARY KEY,
			server_id BIGINT UNSIGNED,
			name TEXT,
			FOREIGN KEY (server_id) REFERENCES servers(server_id) ON DELETE CASCADE
		)`)
	if err != nil {
		errorCreatingTable("channels", err)
	}

	// server members table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS server_members (
		server_id BIGINT UNSIGNED,
		user_id BIGINT UNSIGNED,
		FOREIGN KEY (server_id) REFERENCES servers(server_id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	)`)
	if err != nil {
		errorCreatingTable("server_members", err)
	}

	// messages table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		message_id BIGINT UNSIGNED PRIMARY KEY,
		channel_id BIGINT UNSIGNED,
		user_id BIGINT UNSIGNED,
		message TEXT,
		FOREIGN KEY (channel_id) REFERENCES channels(channel_id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	)`)
	if err != nil {
		errorCreatingTable("messages", err)
	}

	// tokens table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS tokens (
		token BINARY(128) PRIMARY KEY,
		user_id BIGINT UNSIGNED,
		expiration BIGINT UNSIGNED,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	)`)
	if err != nil {
		errorCreatingTable("tokens", err)
	}

	// profile pictures table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS profilepics (
			hash BINARY(32) PRIMARY KEY,
			file_name TEXT
		)`)
	if err != nil {
		errorCreatingTable("profilepics", err)
	}
}

func (d *Database) AddChatMessage(messageID uint64, channelID uint64, userID uint64, message string) string {
	log.Debug("Adding chat message ID [%d] from user ID [%d] to channel ID [%d]...", messageID, userID, channelID)
	const query string = "INSERT INTO messages (message_id, channel_id, user_id, Message) VALUES (?, ?, ?, ?)"
	_, err := d.db.Exec(query, messageID, channelID, userID, message)
	if err != nil {
		log.Error(err.Error())
		if strings.Contains(err.Error(), "Error 1452") {
			log.Hack("Failed adding message ID [%d] into database for channel ID [%d] from user ID [%d], there is no channel with given ID", messageID, channelID, userID)
			return fmt.Sprintf("Can't add message ID [%d] to channel ID [%d]", messageID, channelID)
		}
		log.Fatal("Error adding chat message ID [%d] from user ID [%d] into database", messageID, userID)
	}
	log.Debug("Chat message ID [%d] from user ID [%d] has been added into database successfully", messageID, userID)
	return ""
}

func (d *Database) GetChatMessageOwner(messageID uint64) (uint64, bool) {
	log.Debug("Searching for owner of message ID [%d]...", messageID)
	const query string = "SELECT user_id FROM messages WHERE message_id = ?"
	var userID uint64
	err := d.db.QueryRow(query, messageID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows { // there is no user with this name
			log.Warn("No message found in messages with ID %d", messageID)
			return 0, false
		}
		log.FatalError(err.Error(), "Error getting user ID of the owner of message ID [%d]", messageID)
	}
	log.Debug("Owner ID of message ID [%d] was confirmed to be: [%d]", messageID, userID)
	return userID, true
}

func (d *Database) DeleteChatMessage(messageID uint64, userID uint64) uint64 {
	log.Debug("Deleting message ID [%d] from db table [messages]...", messageID)

	const query string = "DELETE FROM messages WHERE message_id = ? AND user_id = ? RETURNING channel_id"

	var channelID uint64

	err := d.db.QueryRow(query, messageID, userID).Scan(&channelID)
	if err != nil {
		log.Error(err.Error())
		if err == sql.ErrNoRows {
			log.Hack("User ID [%d] doesn't own any message with ID [%d]", userID, messageID)
			return 0
		}
		log.Fatal("Error deleting message ID [%d] of user ID [%d]", messageID, userID)
	}

	log.Debug("Message ID [%d] from user ID [%d] was deleted from database", messageID, userID)
	return channelID
}

func (d *Database) GetMessagesFromChannel(channelID uint64) []ChatMessageResponse {
	const query string = "SELECT message_id, user_id, message FROM messages WHERE channel_id = ?"

	rows, err := d.db.Query(query, channelID)
	if err != nil {
		log.FatalError(err.Error(), "Error searching for messages on channel ID [%d]", channelID)
	}

	var messages []ChatMessageResponse

	var counter int = 0
	for rows.Next() {
		counter++
		var message ChatMessageResponse
		err := rows.Scan(&message.IDm, &message.IDu, &message.Msg)
		if err != nil {
			log.FatalError(err.Error(), "Error scanning message row into struct in channel ID [%d]:", channelID)
		}
		messages = append(messages, message)
	}

	if counter == 0 {
		log.Debug("No messages found on channel ID: [%d]", channelID)
		return messages
	}

	log.Debug("Messages from channel ID [%d] were retrieved successfully", channelID)
	return messages
}

func (d *Database) AddServer(serverID uint64, ownerID uint64, serverName string, picture string) {
	log.Debug("Adding server ID [%d] of owner ID [%d] into database...", serverID, ownerID)
	const query string = "INSERT INTO servers (server_id, owner_id, name, picture) VALUES (?, ?, ?, ?)"
	_, err := d.db.Exec(query, serverID, ownerID, serverName, picture)
	if err != nil {
		log.FatalError(err.Error(), "Error adding server ID [%d] into database", serverID)
	}
	log.Debug("Successfully added server ID [%d] into database", serverID)
}

func (d *Database) GetServerList(userID uint64) []ServerResponse {
	log.Debug("Getting server list of user ID [%d]...", userID)
	const query string = "SELECT server_id, name, picture FROM servers"

	rows, err := d.db.Query(query)
	if err != nil {
		log.FatalError(err.Error(), "Error searching for server list of user ID [%d]", userID)
	}

	var servers []ServerResponse

	var counter int = 0
	for rows.Next() {
		counter++
		var server = ServerResponse{}
		err := rows.Scan(&server.ServerID, &server.Name, &server.Picture)
		if err != nil {
			log.FatalError(err.Error(), "Error scanning server row into struct for user ID [%d]:", userID)
		}
		servers = append(servers, server)
	}

	if counter == 0 {
		log.Debug("User ID [%d] is not in any servers", userID)
		return servers
	}

	log.Debug("Servers for user ID [%d] were retrieved successfully", userID)
	return servers
}

func (d *Database) DeleteServer(serverID uint64, userID uint64) bool {
	log.Debug("Deleting server ID [%d] of user ID [%d]...", serverID, userID)

	const query string = "DELETE FROM servers WHERE server_id = ? AND owner_id = ?"

	var start = time.Now().UnixMilli()
	result, err := d.db.Exec(query, serverID, userID)
	if err != nil {
		log.FatalError(err.Error(), "Error deleting server ID [%d] of user ID [%d]", serverID, userID)
	}
	log.Debug("Server deletion query took [%d ms]", time.Now().UnixMilli()-start)

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.FatalError(err.Error(), "Error getting rowsAffected while deleting server ID [%d] of user ID [%d]", serverID, userID)
	}

	if rowsAffected == 1 {
		log.Debug("Server ID [%d] of user ID [%d] was deleted successfully", serverID, userID)
		return true
	} else if rowsAffected == 0 {
		log.Hack("User ID [%d] doesn't own any server with ID [%d]", userID, serverID)
		return false
	} else {
		// this is not supposed to happen at all since it's not possible to have 2 messages with same ID
		log.Fatal("Multiple servers with same server ID [%d] were found and deleted", serverID)
		return false
	}
}

func (d *Database) AddChannel(channelID uint64, serverID uint64, channelName string) bool {
	log.Debug("Adding channel ID [%d] into database...", channelID)
	const query string = "INSERT INTO channels (channel_id, server_id, name) VALUES (?, ?, ?)"
	_, err := d.db.Exec(query, channelID, serverID, channelName)
	if err != nil {
		log.Error(err.Error())
		if strings.Contains(err.Error(), "Error 1452") {
			log.Hack("Failed adding channel ID [%d] into database for server ID [%d], there is no server with given ID", channelID, serverID)
			return false
		}
		log.Fatal("Error adding channel ID [%d] into database", channelID)
	}
	log.Debug("Successfully added channel ID [%d] into database", channelID)
	return true
}

func (d *Database) GetChannelList(serverID uint64) []ChannelResponse {
	log.Debug("Getting channel list of server ID [%d]...", serverID)
	const query string = "SELECT channel_id, name FROM channels WHERE server_id = ?"

	rows, err := d.db.Query(query, serverID)
	if err != nil {
		log.FatalError(err.Error(), "Error searching for channels list of server ID [%d]", serverID)
	}

	var channels []ChannelResponse

	var counter int = 0
	for rows.Next() {
		counter++
		var channel = ChannelResponse{}
		err := rows.Scan(&channel.ChannelID, &channel.Name)
		if err != nil {
			log.FatalError(err.Error(), "Error scanning channel row into struct from server ID [%d]:", serverID)
		}
		channels = append(channels, channel)
	}

	if counter == 0 {
		log.Debug("Server ID [%d] doesn't have any channels", serverID)
		return channels
	}

	log.Debug("Channels from server ID [%d] were retrieved successfully", serverID)
	return channels
}

func (d *Database) RegisterNewUser(userId uint64, username string, displayName string, passwordHash []byte, totpSecret string) bool {
	log.Debug("Registering new username [%s] into database...", username)
	const query string = "INSERT INTO users (user_id, username, display_name, password, totp) VALUES (?, ?, ?, ?, ?)"
	_, err := d.db.Exec(query, userId, username, displayName, passwordHash, totpSecret)
	if err != nil {
		log.Error(err.Error())
		if strings.Contains(err.Error(), "Error 1062") {
			log.Debug("Failed registering user [%s], username is already taken", username)
			return false
		}
		log.Fatal("Error registering username [%s] into database", username)
	}
	log.Debug("Username [%s] was registered into database successfully with id [%d]", username, userId)
	return true
}

// func (d *Database) GetUserID(username string) uint64 {
// 	printWithName(username, "Searching for field [user_id] in database...")
// 	const query string = "SELECT user_id FROM users WHERE username = ?"
// 	var userID uint64
// 	err := d.db.QueryRow(query, username).Scan(&userID)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			log.Println("No user was found with username [%s]", username)
// 			return 0
// 		}
// 		log.Println(err.Error())
// 		log.Panicf("Error getting user ID of username [%s] from database", username)
// 	}
// 	log.Println("User ID of username [%s] was retrieved from database successfully", username)
// 	return userID
// }

func (d *Database) GetUsername(userID uint64) string {
	log.Debug("Searching for field [username] in database using user ID [%d]...", userID)
	const query string = "SELECT username FROM users WHERE user_id = ?"
	var userName string
	err := d.db.QueryRow(query, userID).Scan(&userName)
	if err != nil {
		log.Error(err.Error())
		if err == sql.ErrNoRows { // there is no user with this id
			log.Debug("No user was found with user ID [%d]", userID)
			return ""
		}
		log.Fatal("Error getting username of user ID [%d] from database", userID)
	}
	log.Debug("Username of user ID [%d] was retrieved from database successfully", userID)
	return userName
}

func (d *Database) GetPasswordAndID(username string) ([]byte, uint64) {
	log.Debug("Searching for password of user [%s] in database...", username)
	const query string = "SELECT user_id, password FROM users WHERE username = ?"
	var passwordHash []byte
	var userID uint64
	err := d.db.QueryRow(query, username).Scan(&userID, &passwordHash)
	if err != nil {
		log.Error(err.Error())
		if err == sql.ErrNoRows {
			log.Debug("No user was found with user [%s]", username)
			return nil, 0
		}
		log.Fatal("Error getting password of user [%s] from database", username)
	}
	log.Debug("Password of user  [%s] was retreived from database successfully", username)
	return passwordHash, userID
}

func (d *Database) AddToken(token Token) {
	log.Debug("Adding new token of user ID [%d] into database...", token.UserID)
	const query string = "INSERT INTO tokens (token, user_id, expiration) VALUES (?, ?, ?)"
	_, err := d.db.Exec(query, token.Token, token.UserID, token.Expiration)
	if err != nil {
		log.FatalError(err.Error(), "Error adding new token for user ID [%d] into database", token.UserID)
	}
	log.Debug("Added a new token for user ID [%d] into database", token.UserID)
}

func (d *Database) ConfirmToken(tokenBytes []byte) uint64 {
	log.Debug("Searching for token in database...")

	const query string = "SELECT user_id, expiration FROM tokens WHERE token = ?"

	var userID uint64
	var expiration uint64

	err := d.db.QueryRow(query, tokenBytes).Scan(&userID, &expiration)
	if err != nil {
		log.Error(err.Error())
		if err == sql.ErrNoRows { // token was not found
			log.Debug("Token was not found in database: [%s]", hex.EncodeToString(tokenBytes))
			return 0
		}
		log.Fatal("Error retrieving token [%s] from database", hex.EncodeToString(tokenBytes))
	}
	log.Debug("Given token was successfully found in database, it belongs to user ID [%d]", userID)
	return userID
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
// 			log.Panicf("%s: Error executing SELECT query: %s", userIDArgs, err)
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
