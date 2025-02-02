package database

import (
	log "chat-app/modules/logging"
	"database/sql"
	"fmt"
)

type User struct {
	UserID      uint64
	Username    string
	DisplayName string
	Status      byte
	StatusText  string
	Picture     string
	Password    []byte
	Totp        string
}

type InitialData struct {
	UserID      uint64
	DisplayName string
	ProfilePic  string
	Pronouns    string
	StatusText  string
	Friends     []uint64
	Blocks      []uint64
	Servers     []JoinedServer
}

type UserData struct {
	UserID     uint64
	Name       string
	Pic        string
	Status     byte
	StatusText string
}

const insertUserQuery = "INSERT INTO users (user_id, username, display_name, password) VALUES (?, ?, ?, ?)"

func CreateUsersTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS users (
		user_id BIGINT UNSIGNED PRIMARY KEY,
		username VARCHAR(32) NOT NULL,
		display_name VARCHAR(32) NOT NULL,
		status TINYINT UNSIGNED	NOT NULL DEFAULT 1,
		status_text VARCHAR(32) NOT NULL DEFAULT '',
		pronouns VARCHAR(16) NOT NULL DEFAULT '',
		picture VARCHAR(255) NOT NULL DEFAULT '',
		password BINARY(60) NOT NULL,
		totp CHAR(32) NOT NULL DEFAULT '',
		UNIQUE(username)
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating users table")
	}
}

func RegisterUser(userID uint64, username string, passwordHash []byte) bool {
	var user = User{
		UserID:   userID,
		Username: username,
		Password: passwordHash,
	}

	err := Insert(user)
	if err != nil {
		log.Trace("Failed to register username [%s] into database", username)
		return false
	}

	log.Trace("Successfully registered username [%s] as user ID [%d] in database", username, userID)
	return true
}

func GetDisplayName(userID uint64) string {
	const query string = "SELECT display_name FROM users WHERE user_id = ?"
	log.Query(query, userID)

	var displayName string
	err := Conn.QueryRow(query, userID).Scan(&displayName)
	DatabaseErrorCheck(err)

	if displayName == "" {
		log.Trace("Failed to find user ID [%d] in database", userID)
	} else {
		log.Trace("Display name of user ID [%d] was retrieved from database successfully", userID)
	}

	return displayName
}

func GetUsername(userID uint64) string {
	const query string = "SELECT username FROM users WHERE user_id = ?"
	log.Query(query, userID)

	var username string
	err := Conn.QueryRow(query, userID).Scan(&username)
	DatabaseErrorCheck(err)

	if username == "" {
		log.Hack("Failed getting username of user ID [%d]", userID)
	} else {
		log.Trace("Username of user ID [%d] was retrieved from database successfully", userID)
	}

	return username
}

func GetUserStatus(userID uint64) byte {
	const query string = "SELECT status FROM users WHERE user_id = ?"
	log.Query(query, userID)

	var status byte = 0
	err := Conn.QueryRow(query, userID).Scan(&status)
	DatabaseErrorCheck(err)

	if status == 0 {
		log.Hack("Failed getting user status of user ID [%d]", userID)
	} else {
		log.Trace("Status value of user ID [%d] was retrieved from database successfully", userID)
	}

	return status
}
func GetPasswordAndID(username string) ([]byte, uint64) {
	const query = "SELECT password, user_id FROM users WHERE username = ?"
	log.Query(query, username)

	var password []byte = nil
	var userID uint64
	err := Conn.QueryRow(query, username).Scan(&password, &userID)
	DatabaseErrorCheck(err)

	if userID == 0 || password == nil {
		log.Trace("Failed to find username [%s] in database", username)
	} else {
		log.Trace("Password and user ID of username [%s] was retrieved from database successfully", username)
	}

	return password, userID

}

func CheckIfUsernameExists(username string) bool {
	const query string = "SELECT EXISTS (SELECT 1 FROM users WHERE username = ?)"
	log.Query(query, username)

	var taken bool = false
	err := Conn.QueryRow(query, username).Scan(&taken)
	DatabaseErrorCheck(err)

	if taken {
		log.Trace("Username [%s] is already taken", username)
	} else {
		log.Hack("Username [%s] is free", username)
	}

	return taken
}

func GetInitialData(userID uint64) (*InitialData, bool) {
	tx, err := Conn.Begin()
	transactionErrorCheck(err)

	defer tx.Rollback()
	//if err != nil {
	//	log.FatalError(err.Error(), "Error rolling back transaction in GetInitialData")
	//}

	initData := InitialData{
		UserID:  userID,
		Blocks:  []uint64{},
		Friends: []uint64{},
		Servers: []JoinedServer{},
	}

	// get user data
	const query1 = "SELECT display_name, picture, status_text, pronouns FROM users WHERE user_id = ?"
	log.Query(query1, userID)

	err = tx.QueryRow(query1, userID).Scan(&initData.DisplayName, &initData.ProfilePic, &initData.StatusText, &initData.Pronouns)
	transactionErrorCheck(err)

	// get block list
	const query2 = "SELECT blocked_id FROM block_list WHERE user_id = ?"
	log.Query(query2, userID)

	rows2, err := tx.Query(query2, userID)
	DatabaseErrorCheck(err)
	for rows2.Next() {
		var blockedID uint64
		err := rows2.Scan(&blockedID)
		DatabaseErrorCheck(err)
		initData.Blocks = append(initData.Blocks, blockedID)
	}

	// get friends
	const query3 = `
		SELECT 
			CASE
				WHEN user1_id = ? THEN user2_id
				WHEN user2_id = ? THEN user1_id
			END AS friend_id
		FROM friendships
		WHERE user1_id = ? OR user2_id = ?
		`

	log.Query(query3, userID, userID, userID, userID)

	rows3, err := tx.Query(query3, userID, userID, userID, userID)
	DatabaseErrorCheck(err)
	for rows3.Next() {
		var friendID uint64
		err := rows3.Scan(&friendID)
		DatabaseErrorCheck(err)
		initData.Friends = append(initData.Friends, friendID)
	}

	// get servers
	const query4 string = "SELECT s.* FROM servers s JOIN server_members m ON s.server_id = m.server_id WHERE m.user_id = ?"
	log.Query(query4, userID)

	rows4, err := tx.Query(query4, userID)
	DatabaseErrorCheck(err)

	for rows4.Next() {
		var server JoinedServer
		var ownerID uint64
		err := rows4.Scan(&server.ServerID, &ownerID, &server.Name, &server.Picture, &server.Banner)
		DatabaseErrorCheck(err)
		log.Trace("Owner ID: ", ownerID, " User ID: ", userID, "")
		if ownerID == userID {
			server.Owned = true
		} else {
			server.Owned = false
		}

		initData.Servers = append(initData.Servers, server)
	}

	err = tx.Commit()
	transactionErrorCheck(err)

	return &initData, true
}

func UpdateUserValue(userID uint64, value string, column string) bool {
	var query = fmt.Sprintf("UPDATE users SET %s = ? WHERE user_id = ?", column)
	log.Query(query, value, userID)

	var result sql.Result
	var err error
	switch column {
	case "status":
		result, err = Conn.Exec(query, value[0], userID)
	default:
		result, err = Conn.Exec(query, value, userID)
	}
	DatabaseErrorCheck(err)

	rowsAffected, err := result.RowsAffected()
	DatabaseErrorCheck(err)

	if rowsAffected == 1 {
		log.Debug("Updated [%s] of user ID [%d] in database", column, userID)
		return true
	} else {
		log.Debug("No changes were made to [%s] of user ID [%d] in database", column, userID)
		return false
	}
}

func GetUserData(userID uint64) ServerMember {
	const query = "SELECT display_name, picture, status, status_text FROM users WHERE user_id = ?"
	log.Query(query, userID)

	serverMember := ServerMember{
		UserID: userID,
	}

	err := Conn.QueryRow(query, userID).Scan(&serverMember.Name, &serverMember.Pic, &serverMember.Status, &serverMember.StatusText)
	DatabaseErrorCheck(err)

	return serverMember
}
