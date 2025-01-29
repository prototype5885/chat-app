package database

import (
	log "chat-app/modules/logging"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/sijms/go-ora/v2"
)

var Conn *sql.DB

var emptyArray = []byte("[]")

var sqlite bool = false

func ConnectSqlite() {
	// if err := os.MkdirAll("database", os.ModePerm); err != nil {
	// 	log.FatalError(err.Error(), "Error creating sqlite database folder")
	// }

	var err error
	Conn, err = sql.Open("sqlite3", "./database.sqlite")
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

//func ConnectOracle() {
//	connectionString := "oracle://" + "chatapp" + ":" + "quLRYL~(NOh8TPI_" + "@" + "adb.eu-zurich-1.oraclecloud.com" + ":" + "1522" + "/" + "g62216a4ff7f865_j7xoz0c10y6k73gq_high.adb.oraclecloud.com"
//
//	connectionString += "?TRACE FILE=trace.log&SSL=enable&SSL Verify=false&WALLET=Wallet_J7XOZ0C10Y6K73GQ"
//
//	var err error
//	Conn, err = sql.Open("oracle", connectionString)
//	if err != nil {
//		panic(fmt.Errorf("error in sql.Open: %w", err))
//	}
//}

func ConnectToMySQL(username string, password string, address string, port string, dbName string) {
	var err error
	Conn, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?clientFoundRows=true", username, password, address, port, dbName))
	if err != nil {
		log.FatalError(err.Error(), "Error opening mysql/mariadb connection")
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
	log.Trace("Creating tables in database...")
	CreateUsersTable()
	CreateTokensTable()
	CreateServersTable()
	CreateServerMembersTable()
	CreateChannelsTable()
	CreateChatMessagesTable()
	CreateFriendshipsTable()
	CreateBlockListTable()
	CreateDmChatTable()
	CreateServerInvitesTable()
	CreateAttachmentsTable()
	CreateInviteKeysTable()
	CreateBotTable()
}

func DatabaseErrorCheck(err error) {
	if err != nil {
		log.WarnError(err.Error(), "Error in db")
		//if sqlite {
		//	if err == sql.ErrNoRows {
		//		log.WarnError(err.Error(), "No row was returned")
		//	} else {
		//		log.FatalError(err.Error(), "Fatal error in database")
		//	}
		//} else {
		//
		//}
	}
}

func transactionErrorCheck(err error) {
	if err != nil {
		log.FatalError(err.Error(), "Fatal error executing database transaction")
	}
}

func Insert(structs any) error {
	var err error
	switch s := structs.(type) {
	case Channel:
		log.Query(insertChannelQuery, s.ChannelID, s.ServerID, s.Name)
		_, err = Conn.Exec(insertChannelQuery, s.ChannelID, s.ServerID, s.Name)
	case Message:
		log.Query(insertChatMessageQuery, s.MessageID, s.ChannelID, s.UserID, s.Message, s.HasAttachments, 0)
		_, err = Conn.Exec(insertChatMessageQuery, s.MessageID, s.ChannelID, s.UserID, s.Message, s.HasAttachments, 0)
	case Attachment:
		log.Query(insertAttachmentQuery, s.Hash, s.MessageID, s.Name)
		_, err = Conn.Exec(insertAttachmentQuery, s.Hash, s.MessageID, s.Name)
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
		log.Query(insertServerInviteQuery, s.InviteID, s.ServerID, s.TargetUserID, s.SingleUse, s.Expiration)
		_, err = Conn.Exec(insertServerInviteQuery, s.InviteID, s.ServerID, s.TargetUserID, s.SingleUse, s.Expiration)
	case Friendship:
		log.Query(insertFriendshipQuery, s.FirstUserID, s.SecondUserID, s.FriendsSince)
		_, err = Conn.Exec(insertFriendshipQuery, s.FirstUserID, s.SecondUserID, s.FriendsSince)
	case BlockUser:
		log.Query(insertBlockListQuery, s.UserID, s.BlockedUserID)
		_, err = Conn.Exec(insertBlockListQuery, s.UserID, s.BlockedUserID)
	case DmChat:
		log.Query(insertDmChatQuery, s.UserID1, s.UserID2, s.ChatID)
		_, err = Conn.Exec(insertDmChatQuery, s.UserID1, s.UserID2, s.ChatID)
	default:
		log.Fatal("Unknown struct type in database Insert: %T", s)
	}
	if err != nil {
		if sqlite { // sqlite
			log.Error(err.Error())
			log.Warn("SQLite Error Code: %d\n", err.(sqlite3.Error).Code)
			log.Warn("SQLite Error Message: %s\n", err.(sqlite3.Error).Error())
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
			log.Error(err.Error())
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

	return nil
}

func Delete(structo any) bool {
	var err error
	var result sql.Result
	switch s := structo.(type) {
	case ChannelDelete:
		log.Query(deleteChannelQuery, s.ChannelID, s.ServerID)
		result, err = Conn.Exec(deleteChannelQuery, s.ChannelID, s.ServerID)
	case ServerDelete:
		log.Query(deleteServerQuery, s.ServerID, s.UserID)
		result, err = Conn.Exec(deleteServerQuery, s.ServerID, s.UserID)
	case ServerInviteDelete:
		log.Query(deleteServerInviteQuery, s.InviteID)
		result, err = Conn.Exec(deleteServerInviteQuery, s.InviteID)
	case Token:
		log.Query(deleteTokenQuery, s.Token, s.UserID)
		result, err = Conn.Exec(deleteTokenQuery, s.Token, s.UserID)
	case DeleteMessage:
		log.Query(deleteChatMessageQuery, s.MessageID, s.UserID)
		result, err = Conn.Exec(deleteChatMessageQuery, s.MessageID, s.UserID)
	case ServerMemberShort:
		log.Query(deleteServerMemberQuery, s.ServerID, s.UserID)
		result, err = Conn.Exec(deleteServerMemberQuery, s.ServerID, s.UserID)
	case FriendshipSimple:
		log.Query(deleteFriendshipQuery, s.UserID, s.ReceiverID)
		result, err = Conn.Exec(deleteFriendshipQuery, s.UserID, s.ReceiverID)
	case BlockUser:
		log.Query(deleteBlockListQuery, s.UserID, s.BlockedUserID)
		result, err = Conn.Exec(deleteBlockListQuery, s.UserID, s.BlockedUserID)
	case InviteKey:
		log.Query(deleteInviteKeyQuery, s.Key)
		result, err = Conn.Exec(deleteInviteKeyQuery, s.Key)
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

	if rowsAffected == 1 {
		return true
	} else if rowsAffected == 0 {
		log.Trace("%s", "no rows were deleted")
		return false
	} else {
		log.Impossible("%s", "multiple rows were deleted")
	}
	return false
}
