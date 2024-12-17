package database

import (
	log "proto-chat/modules/logging"
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

const insertUserQuery = "INSERT INTO users (user_id, username, display_name, password) VALUES (?, ?, ?, ?)"

func CreateUsersTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS users (
		user_id BIGINT UNSIGNED PRIMARY KEY NOT NULL,
		username VARCHAR(32) NOT NULL,
		display_name VARCHAR(64) NOT NULL,
		status TINYINT UNSIGNED	NOT NULL DEFAULT 1,
		status_text VARCHAR(32) NOT NULL DEFAULT '',
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

	success := Insert(user)
	if !success {
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

func GetUserData(userID uint64) (string, string) {
	const query = "SELECT display_name, picture FROM users WHERE user_id = ?"
	log.Query(query, userID)

	var displayName string
	var picture string
	err := Conn.QueryRow(query, userID).Scan(&displayName, &picture)
	DatabaseErrorCheck(err)

	if displayName == "" || picture == "" {
		log.Trace("Failed to find username [%d] in database", userID)
	} else {
		log.Trace("Successfully retrieved display name and profile pic of user ID [%d]", userID)
	}

	return displayName, picture
}
