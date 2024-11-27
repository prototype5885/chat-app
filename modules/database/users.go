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
		status TINYINT UNSIGNED	NOT NULL,
		status_text VARCHAR(32) NOT NULL,
		picture VARCHAR(255) NOT NULL,
		password BINARY(60) NOT NULL,
		totp CHAR(32),
		UNIQUE(username)
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating users table")
	}
}

//func GetDisplayname(userID uint64) string {
//	log.Debug("Searching for field [display_name] in database using user ID [%d]...", userID)
//	const query string = "SELECT display_name FROM users WHERE user_id = ?"
//	var displayName string
//	err := db.QueryRow(query, userID).Scan(&displayName)
//	if err != nil {
//		log.Error(err.Error())
//		if err == sql.ErrNoRows { // there is no user with this id
//			log.Debug("No user was found with user ID [%d]", userID)
//			return ""
//		}
//		log.Fatal("Error getting field [display_name] of user ID [%d] from database", userID)
//	}
//	log.Debug("Display name of user ID [%d] was retrieved from database successfully", userID)
//	return displayName
//}

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
		log.Fatal("Error getting field [username] of user ID [%d] from database", userID)
	}
	log.Debug("Username of user ID [%d] was retrieved from database successfully", userID)
	return userName
}

func GetUserStatus(userID uint64) byte {
	log.Debug("Searching for field [status] in database of user ID [%d]...", userID)
	const query string = "SELECT status FROM users WHERE user_id = ?"
	var status byte
	err := db.QueryRow(query, userID).Scan(&status)
	if err != nil {
		log.Error(err.Error())
		if err == sql.ErrNoRows { // there is no user with this id
			log.Debug("No user was found with user ID [%d]", userID)
			return 0
		}
		log.Fatal("Error getting field [status] of user ID [%d] from database", userID)
	}
	log.Debug("Status value of user ID [%d] was retrieved from database successfully", userID)
	return status
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

func GetUserData(userID uint64) (string, string) {
	log.Debug("Searching for fields [display_name] and [picture] in database of user ID [%d]...", userID)

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
		log.Fatal("Error getting fields [display_name] and picture of user ID [%d] from database", userID)
	}
	log.Debug("Display name and picture of user ID [%d] were retrieved from database successfully", userID)
	return displayName, profilePic
}

func UpdateUserRow(userID uint64, newValueStr string, newValueByte byte, fieldToUpdate string) bool {
	var info = fmt.Sprintf("Updating field [%s] of user ID [%d] with [%s]", fieldToUpdate, userID, newValueStr)
	log.Debug(info)

	var query = fmt.Sprintf("UPDATE users SET %s = ? WHERE user_id = ?", fieldToUpdate)

	var result sql.Result
	var err error

	if newValueStr != "" {
		result, err = db.Exec(query, newValueStr, userID)
	} else {
		result, err = db.Exec(query, newValueByte, userID)
	}

	if err != nil {
		log.FatalError(err.Error(), "Fatal Error: "+info)
		return false
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.FatalError(err.Error(), "Fatal Error: Getting rowsAffected in UpdateUserRow for user ID [%d]", userID)
		return false
	}

	log.Trace("Rows affected: [%d] while updating field [%s] of user ID [%d]", rowsAffected, fieldToUpdate, userID)

	if rowsAffected == 1 {
		if newValueStr != "" {
			log.Debug("Field [%s] of user ID [%d] was successfully changed to [%s]", fieldToUpdate, userID, newValueStr)
		} else {
			log.Debug("Field [%s] of user ID [%d] was successfully changed to [%d]", fieldToUpdate, userID, newValueByte)
		}

		return true
	} else if rowsAffected == 0 {
		log.Hack("User ID [%d] tried to change field [%s] to the same as before", userID, fieldToUpdate)
		return false
	} else {
		log.Impossible("For some reason rowsAffected value is [%d] while changing field [%d] for user ID [%d], it should be only 1", fieldToUpdate, rowsAffected, userID)
		return false
	}
}
