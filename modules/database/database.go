package database

import (
	"database/sql"
	"fmt"
	"os"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-sqlite3"
)

var Conn *sql.DB

var nullJson = []byte("null")

var sqlite bool = false

func ConnectSqlite() {
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

	log.Info("Connected to sqlite")
	sqlite = true
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
	CreateFriendshipsTable()
	CreateBlockListTable()
	CreateServerInvitesTable()
	measureDbTime(start)
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

func measureDbTime(start int64) {
	macros.MeasureTime(start, "Database statement")
}

func Insert(structs any) error {
	start := time.Now().UnixMicro()

	var err error
	switch s := structs.(type) {
	case Channel:
		log.Query(insertChannelQuery, s.ChannelID, s.ServerID, s.Name)
		_, err = Conn.Exec(insertChannelQuery, s.ChannelID, s.ServerID, s.Name)
	case Message:
		log.Query(insertChatMessageQuery, s.MessageID, s.ChannelID, s.UserID, s.Message, s.Attachments)
		_, err = Conn.Exec(insertChatMessageQuery, s.MessageID, s.ChannelID, s.UserID, s.Message, s.Attachments)
	case Attachment:
		log.Query(insertAttachmentQuery, s.FileName, s.FileExtension, s.MessageID)
		_, err = Conn.Exec(insertAttachmentQuery, s.FileName, s.FileExtension, s.MessageID)
	case Server:
		log.Query(insertServerQuery, s.ServerID, s.UserID, s.Name, s.Picture)
		_, err = Conn.Exec(insertServerQuery, s.ServerID, s.UserID, s.Name, s.Picture)
	case Token:
		log.Query(insertTokenQuery, s.Token, s.UserID, s.Expiration)
		_, err = Conn.Exec(insertTokenQuery, s.Token, s.UserID, s.Expiration)
	case User:
		log.Query(insertUserQuery, s.UserID, s.Username, s.Username, s.Password)
		_, err = Conn.Exec(insertUserQuery, s.UserID, s.Username, s.Username, s.Password)
	case ServerMemberShort:
		log.Query(insertServerMemberQuery, s.ServerID, s.UserID)
		_, err = Conn.Exec(insertServerMemberQuery, s.ServerID, s.UserID)
	case ServerInvite:
		log.Query(insertServerInviteQuery, s.InviteID, s.ServerID, s.SingleUse, s.Expiration)
		_, err = Conn.Exec(insertServerInviteQuery, s.InviteID, s.ServerID, s.SingleUse, s.Expiration)
	case Friendship:
		log.Query(insertFriendshipQuery, s.FirstUserID, s.SecondUserID, s.FriendsSince)
		_, err = Conn.Exec(insertFriendshipQuery, s.FirstUserID, s.SecondUserID, s.FriendsSince)
	case BlockUser:
		log.Query(insertBlockListQuery, s.UserID, s.BlockedUserID)
		_, err = Conn.Exec(insertBlockListQuery, s.UserID, s.BlockedUserID)
	default:
		log.Fatal("Unknown struct type in database Insert: %T", s)
	}

	if err != nil {
		if sqlite { // sqlite
			fmt.Printf("SQLite Error Code: %d\n", err.(sqlite3.Error).Code)
			fmt.Printf("SQLite Error Message: %s\n", err.(sqlite3.Error).Error())
			return err
			// if strings.Contains(err.Error(), "275") { // constraint check failed, 2 or more values are duplicates
			// 	log.Error(err.Error(), "%s", "duplicate values where it's enforced to not have")
			// 	return err
			// } else if strings.Contains(err.Error(), "1555") { // duplicate primary key
			// 	log.Error(err.Error(), "%s", "duplicate primary key")
			// 	return err
			// } else if strings.Contains(err.Error(), "2067") { // duplicate unique value
			// 	log.Error(err.Error(), "%s", "duplicate unique value")
			// 	return err
			// } else if strings.Contains(err.Error(), "787") { // no foreign key, no owner
			// 	log.Error(err.Error(), "%s", "foreign key/owner doesn't exist")
			// 	return err
			// } else { // unknown error
			// 	log.FatalError(err.Error(), "%s", "fatal error")
			// 	return err
			// }
		} else { // mariadb or mysql
			// if strings.Contains(err.Error(), "4025") { // constraint check failed, 2 or more values are duplicates
			// 	log.Error(err.Error(), "%s", "duplicate values where it's enforced to not have")
			// 	return err
			// } else if strings.Contains(err.Error(), "1452") { // no foreign key, no owner
			// 	log.Error(err.Error(), "%s", "foreign key/owner doesn't exist")
			// 	return err
			// } else if strings.Contains(err.Error(), "1062") { // duplicate primary key or unique value
			// 	log.Error(err.Error(), "%s", "duplicate primary key or unique value")
			// 	return err
			// } else { // unknown error
			// 	log.Error(err.Error(), "%s", "fatal error")
			// 	return err
			// }
			return err
		}

	}
	measureDbTime(start)
	return nil
}

func Delete(structo any) bool {
	start := time.Now().UnixMicro()

	var err error
	var result sql.Result
	switch s := structo.(type) {
	case Channel:
	case Server:
		log.Query(deleteServerQuery, s.ServerID, s.UserID)
		result, err = Conn.Exec(deleteServerQuery, s.ServerID, s.UserID)
	case Token:
		log.Query(deleteTokenQuery, macros.ShortenToken(s.Token), s.UserID)
		result, err = Conn.Exec(deleteTokenQuery, s.Token, s.UserID)
	case User:
	case ServerMemberShort:
		log.Query(deleteServerMemberQuery, s.ServerID, s.UserID)
		result, err = Conn.Exec(deleteServerMemberQuery, s.ServerID, s.UserID)
	case Friendship:
		log.Query(deleteFriendshipQuery, s.FirstUserID, s.SecondUserID)
		result, err = Conn.Exec(deleteFriendshipQuery, s.FirstUserID, s.SecondUserID)
	case BlockUser:
		log.Query(deleteBlockListQuery, s.UserID, s.BlockedUserID)
		result, err = Conn.Exec(deleteBlockListQuery, s.UserID, s.BlockedUserID)
	default:
		log.Fatal("Unknown type in database [%T]", s)
	}

	if err != nil {
		log.FatalError(err.Error(), "%s", "fatal error")
		// if sqlite {

		// } else {

		// }
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.FatalError(err.Error(), "%s", "error getting RowsAffected")
	}

	measureDbTime(start)

	if rowsAffected == 1 {
		return true
	} else if rowsAffected == 0 {
		log.Trace("%s", "no rows were deleted")
		return false
	} else {
		log.Impossible("%s", "multiple rows were deleted")
		return false
	}
}
