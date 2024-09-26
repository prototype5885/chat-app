package database

import (
	"database/sql"
	"fmt"
	"os"
	log "proto-chat/modules/logging"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

type Users struct{}
type Tokens struct{}
type Servers struct{}
type ServerMembers struct{}
type Channels struct{}
type ChatMessages struct{}
type ProfilePics struct{}

var db *sql.DB

type Info struct {
	ValueNames string
	ValueCount int
}

var (
	UsersTable         Users
	TokensTable        Tokens
	ServersTable       Servers
	ServerMembersTable ServerMembers
	ChannelsTable      Channels
	ChatMessagesTable  ChatMessages
	ProfilePicsTable   ProfilePics
)

func ConnectSqlite() {
	log.Info("Opening sqlite database...")

	//os.Remove("./database/database.db")

	if err := os.MkdirAll("database", os.ModePerm); err != nil {
		log.FatalError(err.Error(), "Error creating sqlite database folder")
	}

	var err error
	db, err = sql.Open("sqlite", "./database/database.db")
	if err != nil {
		log.FatalError(err.Error(), "Error opening sqlite file")
	}

	db.SetMaxOpenConns(1)
}

func ConnectMariadb(username string, password string, address string, port string, dbName string) {
	log.Info("Opening MySQL/MariaDB database...")

	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, address, port, dbName))
	if err != nil {
		log.FatalError(err.Error(), "Error opening mariadb connection")
	}

	db.SetMaxOpenConns(100)
}

func CloseDatabaseConnection() error {
	fmt.Println("Closing main db connection...")
	err := db.Close()
	fmt.Println("Closed main db connection successfully")
	return err
}

func CreateTables() {
	UsersTable.CreateUsersTable()
	TokensTable.CreateTokensTable()
	ServersTable.CreateServersTable()
	ServerMembersTable.CreateServerMembersTable()
	ChannelsTable.CreateChannelsTable()
	ChatMessagesTable.CreateChatMessagesTable()
	ProfilePicsTable.CreateProfilePicsTable()
}

func Insert(structo any) bool {
	start := time.Now().UnixMicro()
	// makeQuestionMarks := func(valuesCount int) string {
	// 	questionMarks := make([]string, valuesCount)
	// 	for i := 0; i < valuesCount; i++ {
	// 		questionMarks[i] = "?"
	// 	}
	// 	return strings.Join(questionMarks, ", ")
	// }

	var typeName string
	var tableName string
	var id uint64

	printInsertingMsg := func() {
		log.Debug("Inserting row into db table [%s]", tableName)
	}

	var err error
	switch s := structo.(type) {
	case Channel:
		typeName = "channel"
		tableName = typeName + "s"
		id = s.ChannelID
		printInsertingMsg()
		const query string = "INSERT INTO channels (channel_id, server_id, name) VALUES (?, ?, ?)"
		_, err = db.Exec(query, s.ChannelID, s.ServerID, s.Name)
	case ChatMessage:
		typeName = "message"
		tableName = typeName + "s"
		id = s.MessageID
		printInsertingMsg()
		const query string = "INSERT INTO messages (message_id, channel_id, user_id, message) VALUES (?, ?, ?, ?)"
		_, err = db.Exec(query, s.MessageID, s.ChannelID, s.UserID, s.Message)
	case Server:
		typeName = "server"
		tableName = typeName + "s"
		id = s.ServerID
		printInsertingMsg()
		const query string = "INSERT INTO servers (server_id, owner_id, name, picture) VALUES (?, ?, ?, ?)"
		_, err = db.Exec(query, s.ServerID, s.OwnerID, s.Name, s.Picture)
	case Token:
		typeName = "token"
		id = s.UserID
		const query string = "INSERT INTO tokens (token, user_id, expiration) VALUES (?, ?, ?)"
		_, err = db.Exec(query, s.Token, s.UserID, s.Expiration)
	case User:
		typeName = "user"
		tableName = typeName + "s"
		id = s.UserID
		printInsertingMsg()
		const query string = "INSERT INTO users (user_id, username, display_name, password, totp) VALUES (?, ?, ?, ?, ?)"
		_, err = db.Exec(query, s.UserID, s.Username, s.DisplayName, s.Password, s.Totp)
	default:
		log.Fatal("Unknown type in db Insert: %T", s)
	}

	tableName = typeName + "s"

	if err != nil {
		if strings.Contains(err.Error(), "Error 1452") {
			log.Warn(err.Error())
			log.Hack("Failed adding [%s] ID [%d] into db table [%s], it wouldn't have an owner", typeName, id, tableName)
			return false
		} else if strings.Contains(err.Error(), "Error 1062") {
			log.Warn(err.Error())
			return false
		}
		log.FatalError(err.Error(), "Error adding [%s] ID [%d] into db table [%s]", typeName, id, tableName)
		return false
	}
	var duration int64 = time.Now().UnixMicro() - start
	log.Time("Insert took [%d μs] or [%d ms]", duration, duration/1000)
	return true
}

func Delete(structo any) uint64 {
	start := time.Now().UnixMicro()

	var typeName string
	var tableName string
	var ownerID uint64
	var itemID uint64
	var toReturn uint64 = 1

	printDeletingMsg := func() {
		log.Debug("Deleting row from db table [%s]", tableName)
	}

	var err error
	var result sql.Result
	switch s := structo.(type) {
	case Channel:
	case ChatMessageDeletion:
		typeName = "message"
		tableName = typeName + "s"
		ownerID = s.UserID
		itemID = s.MessageID
		printDeletingMsg()
		const query string = "DELETE FROM messages WHERE message_id = ? AND user_id = ? RETURNING channel_id"
		err = db.QueryRow(query, s.MessageID, s.UserID).Scan(&toReturn)
	case ServerDeletion:
		typeName = "server"
		tableName = typeName + "s"
		ownerID = s.UserID
		itemID = s.ServerID
		printDeletingMsg()
		const query string = "DELETE FROM servers WHERE server_id = ? AND owner_id = ?"
		result, err = db.Exec(query, s.ServerID, s.UserID)
	case Token:
	case User:
	default:
		log.Fatal("Unknown type in db Insert: [%T]", s)
	}

	if err != nil {
		log.Error(err.Error())
		if err == sql.ErrNoRows {
			log.Hack("User ID [%d] doesn't own any [%s] with ID [%d]", ownerID, typeName, itemID)
			return 0
		}
		log.Fatal("Error deleting [%s] ID [%d] of user ID [%d]", typeName, itemID, ownerID)
		return 0
	}

	if result != nil {
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.FatalError(err.Error(), "Error getting rowsAffected while deleting %s ID [%d] of user ID [%d]", typeName, itemID, ownerID)
		}

		if rowsAffected == 0 {
			log.Hack("User ID [%d] doesn't own any [%s] with ID [%d]", ownerID, typeName, itemID)
			return 0
		} else if rowsAffected != 1 {
			// this is not supposed to happen at all since it's not possible to have 2 messages with same ID
			log.Fatal("Multiple [%s] with same server ID [%d] were found and deleted", typeName, itemID)
			return 0
		}
	}

	log.Debug("[%s] ID [%d] from user ID [%d] was deleted from database", typeName, itemID, ownerID)
	var duration int64 = time.Now().UnixMicro() - start
	log.Time("Delete took [%d μs] or [%d ms]", duration, duration/1000)
	return toReturn
}
