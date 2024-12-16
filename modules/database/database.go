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

var Conn *sql.DB

var nullJson = []byte("null")

func ConnectSqlite() {
	if err := os.MkdirAll("database", os.ModePerm); err != nil {
		log.FatalError(err.Error(), "Error creating sqlite database folder")
	}

	var err error
	Conn, err = sql.Open("sqlite", "./database/sqlite.db")
	if err != nil {
		log.FatalError(err.Error(), "Error connecting to sqlite database")
	}

	Conn.SetMaxOpenConns(1)

	_, err = Conn.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.FatalError(err.Error(), "Error enabling foreign keys for sqlite")
	}

	log.Info("Connected to sqlite")
}

func ConnectMariadb(username string, password string, address string, port string, dbName string) {
	var err error
	Conn, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, address, port, dbName))
	if err != nil {
		log.FatalError(err.Error(), "Error opening mariadb connection")
	}

	Conn.SetMaxOpenConns(100)
	log.Info("Connected to mysql/mariadb")
}

func CloseDatabaseConnection() error {
	fmt.Println("Closing main Conn connection...")
	err := Conn.Close()
	return err
}

func CreateTables() {
	log.Trace("Creating database tables...")
	start := time.Now().UnixMicro()
	CreateUsersTable()
	CreateTokensTable()
	CreateServersTable()
	CreateServerMembersTable()
	CreateChannelsTable()
	CreateChatMessagesTable()
	//CreateAttachmentsTable()
	CreateServerInvitesTable()
	measureTime(start)
}

func DatabaseErrorCheck(err error) {
	if err != nil {
		if err == sql.ErrNoRows {
			log.WarnError(err.Error(), "No row was returned")
		} else {
			log.FatalError(err.Error(), "Fatal error in database")
		}
	}
}

func transactionErrorCheck(err error) {
	if err != nil {
		log.FatalError(err.Error(), "Fatal error executing database transaction")
	}
}

func measureTime(start int64) {
	duration := time.Now().UnixMicro() - start
	durationMs := duration / 1000
	log.Time("Database statement took [%d μs] [%d ms]", duration, durationMs)
}

func Insert(structs any) bool {
	start := time.Now().UnixMicro()

	var typeName string
	var tableName string
	var insertedItemID uint64

	var err error
	switch s := structs.(type) {
	case Channel:
		typeName = "channel"
		insertedItemID = s.ChannelID
		log.Query(insertChannelQuery, s.ChannelID, s.ServerID, s.Name)
		_, err = Conn.Exec(insertChannelQuery, s.ChannelID, s.ServerID, s.Name)
	case Message:
		typeName = "message"
		insertedItemID = s.MessageID
		log.Query(insertChatMessageQuery, s.MessageID, s.ChannelID, s.UserID, s.Message, s.Attachments)
		_, err = Conn.Exec(insertChatMessageQuery, s.MessageID, s.ChannelID, s.UserID, s.Message, s.Attachments)
	case Attachment:
		typeName = "attachment"
		insertedItemID = s.MessageID
		log.Query(insertAttachmentQuery, s.FileName, s.FileExtension, s.MessageID)
		_, err = Conn.Exec(insertAttachmentQuery, s.FileName, s.FileExtension, s.MessageID)
	case Server:
		typeName = "server"
		insertedItemID = s.ServerID
		log.Query(insertServerQuery, s.ServerID, s.UserID, s.Name, s.Picture)
		_, err = Conn.Exec(insertServerQuery, s.ServerID, s.UserID, s.Name, s.Picture)
	case Token:
		typeName = "token"
		insertedItemID = s.UserID
		log.Query(insertTokenQuery, s.Token, s.UserID, s.Expiration)
		_, err = Conn.Exec(insertTokenQuery, s.Token, s.UserID, s.Expiration)
	case User:
		typeName = "user"
		insertedItemID = s.UserID
		log.Query(insertUserQuery, s.UserID, s.Username, s.Username, s.Password)
		_, err = Conn.Exec(insertUserQuery, s.UserID, s.Username, s.Username, s.Password)
	case ServerMember:
		typeName = "server_member"
		insertedItemID = s.UserID
		log.Query(insertServerMemberQuery, s.ServerID, s.UserID)
		_, err = Conn.Exec(insertServerMemberQuery, s.ServerID, s.UserID)
	case ServerInvite:
		typeName = "server_invite"
		insertedItemID = s.ServerID
		log.Query(insertServerInviteQuery, s.InviteID, s.ServerID, s.SingleUse, s.Expiration)
		_, err = Conn.Exec(insertServerInviteQuery, s.InviteID, s.ServerID, s.SingleUse, s.Expiration)
	default:
		log.Fatal("Unknown type in Conn Insert: %T", s)
	}

	if err != nil {
		if strings.Contains(err.Error(), "Error 1452") {
			// Error 1452: Cannot add or update a child row: a foreign key constraint fails
			log.WarnError(err.Error(), "Failed adding [%s] ID [%d] into Conn table [%s], it wouldn't have an owner", typeName, insertedItemID, tableName)
			return false
		} else if strings.Contains(err.Error(), "Error 1062") {
			// Error 1062: Duplicate entry for key
			log.WarnError(err.Error(), "Trying to insert duplicate key into database")
			return false
		} else {
			// unknown error
			log.FatalError(err.Error(), "Error adding [%s] ID [%d] into Conn table [%s]", typeName, insertedItemID, tableName)
			return false
		}
	}
	measureTime(start)
	return true
}

func Delete(structo any) bool {
	start := time.Now().UnixMicro()

	var typeName string
	var tableName string
	var deletedItemOwnerID uint64 // deleted item's owner
	var deletedItemID uint64      // deleted item

	printDeletingMsg := func() {
		tableName = typeName + "s"
		log.Trace("Deleting row from Conn table [%s]", tableName)
	}

	var err error
	var result sql.Result
	switch s := structo.(type) {
	case Channel:
	case ServerDeletion:
		typeName = "server"
		deletedItemID = s.ServerID
		deletedItemOwnerID = s.UserID
		printDeletingMsg()
		result, err = Conn.Exec(deleteServerQuery, s.ServerID, s.UserID)
	case Token:
	case User:
	case ServerMember:
		typeName = "user"
		deletedItemID = s.UserID
		deletedItemOwnerID = s.ServerID
		printDeletingMsg()
		result, err = Conn.Exec(deleteServerMemberQuery, s.ServerID, s.UserID)
	default:
		log.Fatal("Unknown type in Conn Insert: [%T]", s)
	}

	if err != nil {
		log.FatalError(err.Error(), "Error deleting [%s] ID [%d] requested by user ID [%d]", typeName, deletedItemID, deletedItemOwnerID)
	}

	measureTime(start)

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.FatalError(err.Error(), "Error getting rowsAffected while deleting %s ID [%d] requested by user ID [%d]", typeName, deletedItemID, deletedItemOwnerID)
	}

	if rowsAffected == 1 {
		log.Trace("[%s] ID [%d] owned by ID [%d] was deleted from database", typeName, deletedItemID, deletedItemOwnerID)
		return true
	} else if rowsAffected == 0 {
		log.Hack("User ID [%d] doesn't own any [%s] with ID [%d]", deletedItemOwnerID, typeName, deletedItemID)
		return false
	} else {
		log.Impossible("Multiple [%s] were found and deleted", typeName)
		return false
	}
}
