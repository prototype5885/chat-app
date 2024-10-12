package database

import (
	"database/sql"
	"fmt"
	log "proto-chat/modules/logging"
)

type User struct {
	UserID      uint64
	Username    string
	DisplayName string
	Picture     string
	Password    []byte
	Totp        string
}

const insertUserQuery string = "INSERT INTO users (user_id, username, display_name, picture, password, totp) VALUES (?, ?, ?, ?, ?, ?)"

func CreateUsersTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		user_id BIGINT UNSIGNED PRIMARY KEY NOT NULL,
		username VARCHAR(32) NOT NULL,
		display_name VARCHAR(64) NOT NULL,
		picture VARCHAR(255) NOT NULL,
		password BINARY(60) NOT NULL,
		totp CHAR(32),
		UNIQUE(username)
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating users table")
	}
}

func GetUsername(userID uint64) string {
	log.Debug("Searching for field [username] in database using user ID [%d]...", userID)
	const query string = "SELECT username FROM users WHERE user_id = ?"
	var userName string
	err := db.QueryRow(query, userID).Scan(&userName)
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

func GetPasswordAndID(username string) ([]byte, uint64) {
	log.Debug("Searching for password of user [%s] in database...", username)
	const query string = "SELECT user_id, password FROM users WHERE username = ?"
	var passwordHash []byte
	var userID uint64
	err := db.QueryRow(query, username).Scan(&userID, &passwordHash)
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

func GetUserInfo(userID uint64) (string, string) {
	log.Debug("Searching for fields display_name and picture in database of user ID [%d]...", userID)

	const query string = "SELECT display_name, picture FROM users WHERE user_id = ?"

	var displayName string
	var profilePic string

	err := db.QueryRow(query, userID).Scan(&displayName, &profilePic)
	if err != nil {
		log.Error(err.Error())
		if err == sql.ErrNoRows { // there is no user with this id
			log.Debug("No user was found with user ID [%d]", userID)
			return "", ""
		}
		log.Fatal("Error getting fields display_name and picture of user ID [%d] from database", userID)
	}
	log.Debug("Display name and picture of user ID [%d] were retrieved from database successfully", userID)
	return displayName, profilePic
}

func ChangeDisplayName(userID uint64, newDisplayName string) bool {
	var info string = fmt.Sprintf("Updating field display_name of user ID [%d] with [%s]", userID, newDisplayName)
	log.Debug(info)

	const query string = "UPDATE users SET display_name = ? WHERE user_id = ?"
	result, err := db.Exec(query, newDisplayName, userID)
	if err != nil {
		log.FatalError(err.Error(), "Fatal Error: "+info)
		return false
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.FatalError(err.Error(), "Fatal Error: Getting rowsAffected in ChangeDisplayName for user ID [%d]", userID)
		return false
	}

	if rowsAffected == 1 {
		log.Debug("Display name of user ID [%d] was successfully changed to [%s]", userID, newDisplayName)
		return true
	} else if rowsAffected == 0 {
		log.Hack("User ID [%d] tried to change their display name to same as their current one", userID)
		return false
	} else {
		log.Impossible("For some reason rowsAffected value is [%d] in ChangeDisplayName for user ID [%d], it should be only 1", rowsAffected, userID)
		return false
	}
}
