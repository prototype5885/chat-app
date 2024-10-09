package database

import (
	"database/sql"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
)

type Token struct {
	Token      []byte
	UserID     uint64
	Expiration uint64
}

const (
	insertTokenQuery = "INSERT INTO tokens (token, user_id, expiration) VALUES (?, ?, ?)"
	deleteTokenQuery = "DELETE FROM tokens WHERE token = ?"
)

func CreateTokensTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS tokens (
		token BINARY(128) PRIMARY KEY NOT NULL,
		user_id BIGINT UNSIGNED NOT NULL,
		expiration BIGINT UNSIGNED NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating tokens table")
	}
}

func ConfirmToken(tokenBytes []byte) (uint64, uint64) {
	log.Debug("Searching for token in database...")

	const query string = "SELECT user_id, expiration FROM tokens WHERE token = ?"

	var userID uint64
	var expiration uint64

	err := db.QueryRow(query, tokenBytes).Scan(&userID, &expiration)
	if err != nil {
		log.Error(err.Error())
		if err == sql.ErrNoRows { // token was not found
			log.Debug("Token was not found in database: [%s]", macros.ShortenToken(tokenBytes))
			return 0, 0
		}
		log.Fatal("Error retrieving token [%s] from database", macros.ShortenToken(tokenBytes))
	}
	log.Debug("Given token was successfully found in database, it belongs to user ID [%d], expires at [%d]", userID, expiration)
	return userID, expiration
}

func RenewTokenExpiration(newExpiration uint64, tokenBytes []byte) {
	log.Debug("Updating expiration date for token [%s] as [%d]...", macros.ShortenToken(tokenBytes), newExpiration)

	const query string = "UPDATE tokens SET expiration = ? WHERE token = ?"

	result, err := db.Exec(query, newExpiration, tokenBytes)
	if err != nil {
		log.FatalError(err.Error(), "Couldn't update token expiration timestamp for token [%s] in database", macros.ShortenToken(tokenBytes))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.FatalError(err.Error(), "Error getting rowsAffected after updating token expiration timestamp for token [%s] in database", macros.ShortenToken(tokenBytes))
	}

	if rowsAffected == 1 {
		log.Debug("Updated expiration timestamp for token [%s] in database", macros.ShortenToken(tokenBytes))
	} else if rowsAffected == 0 {
		log.Debug("No changes were made for expiration timestamp for token [%s] in database", macros.ShortenToken(tokenBytes))
	} else {
		log.Impossible("Multiple expiration timestamps of token [%s] were updated, this is not supposed to be possible at all", macros.ShortenToken(tokenBytes))
	}
}
