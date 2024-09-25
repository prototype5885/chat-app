package database

import (
	"database/sql"
	"encoding/hex"
	log "proto-chat/modules/logging"
)

type Token struct {
	Token      []byte
	UserID     uint64
	Expiration uint64
}

func (t *Tokens) CreateTokensTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS tokens (
		token BINARY(128) PRIMARY KEY,
		user_id BIGINT UNSIGNED,
		expiration BIGINT UNSIGNED,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating tokens table")
	}
}

func (t *Tokens) ConfirmToken(tokenBytes []byte) uint64 {
	log.Debug("Searching for token in database...")

	const query string = "SELECT user_id, expiration FROM tokens WHERE token = ?"

	var userID uint64
	var expiration uint64

	err := db.QueryRow(query, tokenBytes).Scan(&userID, &expiration)
	if err != nil {
		log.Error(err.Error())
		if err == sql.ErrNoRows { // token was not found
			log.Debug("Token was not found in database: [%s]", hex.EncodeToString(tokenBytes))
			return 0
		}
		log.Fatal("Error retrieving token [%s] from database", hex.EncodeToString(tokenBytes))
	}
	log.Debug("Given token was successfully found in database, it belongs to user ID [%d]", userID)
	return userID
}
