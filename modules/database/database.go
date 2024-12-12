package database

import (
	"database/sql"
	"os"
	log "proto-chat/modules/logging"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var Conn *sql.DB

var nullJson = []byte("null")

func ConnectSqlite() {
	//os.Remove("./database/database.Conn")
	if err := os.MkdirAll("database", os.ModePerm); err != nil {
		log.FatalError(err.Error(), "Error creating sqlite database folder")
	}

	var err error
	Conn, err = sql.Open("sqlite3", "./database/sqlite.db")
	if err != nil {
		log.FatalError(err.Error(), "Error connecting to sqlite database")
	}

	Conn.SetMaxOpenConns(1)

	_, err = Conn.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.FatalError(err.Error(), "Error enabling foreign keys for sqlite")
	}

	log.Info("Connection to Sqlite database opened")
}

func ConnectMariadb(username string, password string, address string, port string, dbName string) {
	// var err error
	// Conn, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, address, port, dbName)), &gorm.Config{Logger: newLogger})
	// if err != nil {
	// 	log.FatalError(err.Error(), "Error connecting to mariadb database")
	// }

	// sqlDB, _ := Conn.DB()
	// sqlDB.SetMaxOpenConns(100)

	//var err error
	//Conn, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, address, port, dbName))
	//if err != nil {
	//	log.FatalError(err.Error(), "Error opening mariadb connection")
	//}
	//
	//Conn.SetMaxOpenConns(100)
	log.Info("Connection to MySQL/MariaDB database opened")
}

func CloseDatabaseConnection() error {
	//fmt.Println("Closing main Conn connection...")
	//err := Conn.Close()
	//return err
	return nil
}

func CreateTables() {
	//Conn.AutoMigrate(&User{}, &Token{}, &Server{}, &ServerMember{}, &Channel{}, &message{}, &ServerInvite{})
	CreateUsersTable()
	CreateTokensTable()
	CreateServersTable()
	CreateServerMembersTable()
	CreateChannelsTable()
	CreateChatMessagesTable()
	//CreateAttachmentsTable()
	CreateServerInvitesTable()
}

func DatabaseErrorCheck(err error) {
	if err != nil {
		log.FatalError(err.Error(), "Fatal error in database")
	}
}

func Insert(structs any) bool {
	start := time.Now().UnixMicro()

	var typeName string
	var tableName string
	var insertedItemID uint64

	printInsertingMsg := func() {
		tableName = typeName + "s"
		log.Trace("Inserting row into Conn table [%s]", tableName)
	}

	var err error
	switch s := structs.(type) {
	case Channel:
		typeName = "channel"
		insertedItemID = s.ChannelID
		printInsertingMsg()
		_, err = Conn.Exec(insertChannelQuery, s.ChannelID, s.ServerID, s.Name)
	case Message:
		typeName = "message"
		insertedItemID = s.MessageID
		printInsertingMsg()
		_, err = Conn.Exec(insertChatMessageQuery, s.MessageID, s.ChannelID, s.UserID, s.Message, s.Attachments)
	case Attachment:
		typeName = "attachment"
		insertedItemID = s.MessageID
		printInsertingMsg()
		_, err = Conn.Exec(insertAttachmentQuery, s.FileName, s.FileExtension, s.MessageID)
	case Server:
		typeName = "server"
		insertedItemID = s.ServerID
		printInsertingMsg()
		_, err = Conn.Exec(insertServerQuery, s.ServerID, s.UserID, s.Name, s.Picture)
	case Token:
		typeName = "token"
		insertedItemID = s.UserID
		_, err = Conn.Exec(insertTokenQuery, s.Token, s.UserID, s.Expiration)
	case User:
		typeName = "user"
		insertedItemID = s.UserID
		printInsertingMsg()
		_, err = Conn.Exec(insertUserQuery, s.UserID, s.Username, s.DisplayName, s.Picture, s.Password, s.Totp)
	case ServerMember:
		typeName = "server_member"
		insertedItemID = s.UserID
		printInsertingMsg()
		_, err = Conn.Exec(insertServerMemberQuery, s.ServerID, s.UserID)
	case ServerInvite:
		typeName = "server_invite"
		insertedItemID = s.ServerID
		printInsertingMsg()
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
	var duration int64 = time.Now().UnixMicro() - start
	log.Time("Insert took [%d μs] or [%d ms]", duration, duration/1000)
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

	// first check if there are errors executing the query
	if err != nil {
		log.FatalError(err.Error(), "Error deleting [%s] ID [%d] requested by user ID [%d]", typeName, deletedItemID, deletedItemOwnerID)
	}

	// print time it took
	var duration int64 = time.Now().UnixMicro() - start
	log.Time("Deletion took [%d μs] [%d ms]", duration, duration/1000)

	// get how many rows were affected
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
