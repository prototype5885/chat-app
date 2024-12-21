package database

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"os"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

var Conn *sql.DB

var nullJson = []byte("null")

var sqlite bool = false

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
	case ServerMember:
		log.Query(insertServerMemberQuery, s.ServerID, s.UserID)
		_, err = Conn.Exec(insertServerMemberQuery, s.ServerID, s.UserID)
	case ServerInvite:
		log.Query(insertServerInviteQuery, s.InviteID, s.ServerID, s.SingleUse, s.Expiration)
		_, err = Conn.Exec(insertServerInviteQuery, s.InviteID, s.ServerID, s.SingleUse, s.Expiration)
	default:
		log.Fatal("Unknown type in database Insert: %T", s)
	}

	if err != nil {
		if sqlite { // sqlite
			if strings.Contains(err.Error(), "1555") { // duplicate primary key
				log.Error("%s", err.Error())
				return false
			} else if strings.Contains(err.Error(), "2067") { // duplicate unique value
				log.Error("%s", err.Error())
				return false
			} else if strings.Contains(err.Error(), "787") { // no foreign key, no owner
				log.Error("%s", err.Error())
				return false
			} else { // unknown error
				log.FatalError(err.Error(), "")
				return false
			}
		} else { // mariadb or mysql
			if strings.Contains(err.Error(), "1452") { // no foreign key, no owner
				log.Error("%s", err.Error())
				return false
			} else if strings.Contains(err.Error(), "1062") { // duplicate primary key or unique value
				log.Error("%s", err.Error())
				return false
			} else { // unknown error
				log.Error("%s", err.Error())
				return false
			}
		}

	}
	measureTime(start)
	return true
}

func Delete(structo any) bool {
	start := time.Now().UnixMicro()

	var deletionID uint32 = rand.Uint32()
	strID := fmt.Sprintf("ID: [%d], ", deletionID)

	var err error
	var result sql.Result
	switch s := structo.(type) {
	case Channel:
	case Server:
		log.Query(strID+deleteServerQuery, s.ServerID, s.UserID)
		result, err = Conn.Exec(deleteServerQuery, s.ServerID, s.UserID)
	case Token:
		log.Query(strID+deleteTokenQuery, macros.ShortenToken(s.Token), s.UserID)
		result, err = Conn.Exec(deleteTokenQuery, s.Token, s.UserID)
	case User:
	case ServerMember:
		log.Query(strID+deleteServerMemberQuery, s.ServerID, s.UserID)
		result, err = Conn.Exec(deleteServerMemberQuery, s.ServerID, s.UserID)
	default:
		log.Fatal("Unknown type in database deletion ID [%d]: [%T]", deletionID, s)
	}

	if err != nil {
		log.FatalError(err.Error(), "Fatal error at deletion ID [%d]", deletionID)
		// if sqlite {

		// } else {

		// }
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.FatalError(err.Error(), "Error getting rowsAffected at deletion ID [%d]", deletionID)
	}

	measureTime(start)

	if rowsAffected == 1 {
		return true
	} else if rowsAffected == 0 {
		log.Trace("No rows were deleted at deletion ID [%d]", deletionID)
		return false
	} else {
		log.Impossible("Multiple rows were found and deleted at deletion ID [%d]", deletionID)
		return false
	}
}
