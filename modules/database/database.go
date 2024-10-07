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
type ServerInvites struct{}

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
	ServerInvitesTable ServerInvites
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
	ServerInvitesTable.CreateServerInvitesTable()
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
		tableName = typeName + "s"
		log.Debug("Inserting row into db table [%s]", tableName)
	}

	var err error
	switch s := structo.(type) {
	case Channel:
		typeName = "channel"
		id = s.ChannelID
		printInsertingMsg()
		_, err = db.Exec(insertChannelQuery, s.ChannelID, s.ServerID, s.Name)
	case ChatMessage:
		typeName = "message"
		id = s.MessageID
		printInsertingMsg()
		_, err = db.Exec(insertChatMessageQuery, s.MessageID, s.ChannelID, s.UserID, s.Message)
	case Server:
		typeName = "server"
		id = s.ServerID
		printInsertingMsg()
		_, err = db.Exec(insertServerQuery, s.ServerID, s.OwnerID, s.Name, s.Picture)
	case Token:
		typeName = "token"
		id = s.UserID
		_, err = db.Exec(insertTokenQuery, s.Token, s.UserID, s.Expiration)
	case User:
		typeName = "user"
		id = s.UserID
		printInsertingMsg()
		_, err = db.Exec(insertUserQuery, s.UserID, s.Username, s.DisplayName, s.Picture, s.Password, s.Totp)
	case ServerMember:
		typeName = "server_member"
		id = s.UserID
		printInsertingMsg()
		_, err = db.Exec(insertServerMemberQuery, s.ServerID, s.UserID)
	case ServerInvite:
		typeName = "server_invite"
		id = s.ServerID
		printInsertingMsg()
		_, err = db.Exec(insertServerInviteQuery, s.InviteID, s.ServerID, s.SingleUse, s.Expiration)
	default:
		log.Fatal("Unknown type in db Insert: %T", s)
	}

	if err != nil {
		if strings.Contains(err.Error(), "Error 1452") {
			// Error 1452: Cannot add or update a child row: a foreign key constraint fails
			log.WarnError(err.Error(), "Failed adding [%s] ID [%d] into db table [%s], it wouldn't have an owner", typeName, id, tableName)
			return false
		} else if strings.Contains(err.Error(), "Error 1062") {
			// Error 1062: Duplicate entry for key
			log.WarnError(err.Error(), "Trying to insert duplicate key into database")
			return false
		} else {
			// unknown error
			log.FatalError(err.Error(), "Error adding [%s] ID [%d] into db table [%s]", typeName, id, tableName)
			return false
		}
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
		tableName = typeName + "s"
		log.Debug("Deleting row from db table [%s]", tableName)
	}

	var err error
	var result sql.Result
	switch s := structo.(type) {
	case Channel:
	case ChatMessageDeletion:
		typeName = "message"
		ownerID = s.UserID
		itemID = s.MessageID
		printDeletingMsg()
		const query string = "DELETE FROM messages WHERE message_id = ? AND user_id = ? RETURNING channel_id"
		err = db.QueryRow(query, s.MessageID, s.UserID).Scan(&toReturn)
	case ServerDeletion:
		typeName = "server"
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
			// what was to be deleted was nowhere to be found
			log.Hack("User ID [%d] doesn't own any [%s] with ID [%d]", ownerID, typeName, itemID)
			return 0
		} else {
			// unknown error
			log.Fatal("Error deleting [%s] ID [%d] of user ID [%d]", typeName, itemID, ownerID)
			return 0
		}

	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.FatalError(err.Error(), "Error getting rowsAffected while deleting %s ID [%d] of user ID [%d]", typeName, itemID, ownerID)
	}

	if rowsAffected != 1 {
		// this always should be 1, because if its 0 it should have returned already on sql.ErrNoRows,
		// and it can't be more than 1 either since it's not possible to have duplicate IDs
		log.Fatal("Multiple or none [%s] with server ID [%d] were found and deleted, it always should be 1", typeName, itemID)
		return 0
	}

	// if rowsAffected == 0 {
	// 	log.Hack("User ID [%d] doesn't own any [%s] with ID [%d]", ownerID, typeName, itemID)
	// 	return 0
	// } else if rowsAffected != 1 {
	// 	// this is not supposed to happen at all since it's not possible to have 2 messages with same ID
	// 	log.Fatal("Multiple [%s] with same server ID [%d] were found and deleted", typeName, itemID)
	// 	return 0
	// }

	log.Debug("[%s] ID [%d] from user ID [%d] was deleted from database", typeName, itemID, ownerID)
	var duration int64 = time.Now().UnixMicro() - start
	log.Time("Delete took [%d μs] or [%d ms]", duration, duration/1000)
	return toReturn
}
