package database

import (
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"time"
)

type Token struct {
	Token      []byte
	UserID     uint64
	Expiration int64
}

const insertTokenQuery = "INSERT INTO tokens (token, user_id, expiration) VALUES (?, ?, ?)"
const deleteTokenQuery = "DELETE FROM tokens WHERE token = ? AND user_id = ?"

func CreateTokensTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS tokens (
			token BINARY(128) PRIMARY KEY NOT NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			expiration BIGINT UNSIGNED NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
		)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating tokens table")
	}
}
func ConfirmToken(tokenBytes []byte) Token {
	const query string = "SELECT user_id, expiration FROM tokens WHERE token = ?"
	log.Query(query, macros.ShortenToken(tokenBytes))

	token := Token{
		Token: tokenBytes,
	}

	err := Conn.QueryRow(query, tokenBytes).Scan(&token.UserID, &token.Expiration)
	DatabaseErrorCheck(err)

	if token.UserID == 0 || token.Expiration == 0 {
		log.Debug("Failed getting token [%s] in database", macros.ShortenToken(tokenBytes))
	} else {
		et := time.Unix(token.Expiration, 0)
		formattedDate := et.Format("2006-01-02 15:04:05")
		log.Debug("Token [%s] was found in database, it belongs to user ID [%d], expires at [%s]", macros.ShortenToken(tokenBytes), token.UserID, formattedDate)
	}

	return token
}

func RenewTokenExpiration(newExpiration int64, tokenBytes []byte) bool {
	const query string = "UPDATE tokens SET expiration = ? WHERE token = ?"
	log.Query(query, newExpiration, macros.ShortenToken(tokenBytes))

	result, err := Conn.Exec(query, newExpiration, tokenBytes)
	DatabaseErrorCheck(err)

	rowsAffected, err := result.RowsAffected()
	DatabaseErrorCheck(err)

	if rowsAffected == 1 {
		log.Debug("Updated expiration timestamp for token [%s] in database", macros.ShortenToken(tokenBytes))
		return true
	} else {
		log.Debug("No changes were made for expiration timestamp for token [%s] in database", macros.ShortenToken(tokenBytes))
		return false
	}
}
