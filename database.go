package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"

	//_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

// type User struct {
// 	userID       uint64
// 	username     string
// 	passwordHash string
// 	totpSecret   string
// 	activeTokens string
// }

var db *sql.DB

func ConnectSqlite() {
	log.Println("Opening sqlite database...")

	//os.Remove("./database/database.db")

	if err := os.MkdirAll("database", os.ModePerm); err != nil {
		log.Fatal("Error creating database folder:", err)
	}

	const dbSource string = "./database/database.db"

	var err error
	db, err = sql.Open("sqlite", dbSource)
	if err != nil {
		log.Fatal("Error opening sqlite file:", err)
	}

	db.SetMaxOpenConns(1)
	createTablesDB()
}

func ConnectMariadb(username string, password string, address string, port string, dbName string) {
	log.Println("Opening MySQL/MariaDB database...")

	var dbSource string = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, address, port, dbName)

	var err error
	db, err = sql.Open("mysql", dbSource)
	if err != nil {
		log.Fatal("Error opening mariadb connection:", err)
	}

	db.SetMaxOpenConns(100)
	createTablesDB()
}

func createTablesDB() {
	// users table
	var err error
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		userid BIGINT UNSIGNED PRIMARY KEY,
		username TEXT,
		password BINARY(60),
		totp CHAR(32)
	)`)
	if err != nil {
		log.Fatal("Error executing CREATE TABLE users query:", err)
		return
	}

	// messages table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		messageid BIGINT UNSIGNED PRIMARY KEY,
		channelid BIGINT UNSIGNED,
		userid BIGINT UNSIGNED,
		message TEXT
	)`)
	if err != nil {
		log.Fatal("Error executing CREATE TABLE messages query:", err)
		return
	}

	// tokens table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tokens (
		token TEXT,
		userid BIGINT UNSIGNED,
		expiration BIGINT UNSIGNED
	)`)
	if err != nil {
		log.Fatal("Error executing CREATE TABLE tokens query:", err)
		return
	}
}

func addChatMessageDB(messageID uint64, channelID uint64, userID uint64, message string) Result {
	//log.Println("Adding Message into database...")
	const query string = "INSERT INTO messages (messageid, channelid, userid, Message) VALUES (?, ?, ?, ?)"
	_, err := db.Exec(query, messageID, channelID, userID, message)
	if err != nil {
		log.Fatal("Error executing INSERT query:", err)
		//return Result{
		//	Success: false,
		//	Message: "FATAL: Error adding chat Message to database",
		//}
	}
	return Result{
		Success: true,
		Message: "Chat Message added to database",
	}
}

func addNewUserToDB(userId uint64, username string, passwordHash []byte, totpSecret string) Result {
	printWithName(username, "Adding new user into database...")
	const query string = "INSERT INTO users (userid, username, password, totp) VALUES (?, ?, ?, ?)"
	_, err := db.Exec(query, userId, username, passwordHash, totpSecret)
	if err != nil {
		log.Fatalf("%s: Error executing INSERT query: %s\n", username, err)
		//return Result{
		//	Success: false,
		//	Message: "FATAL: Error adding new user to database",
		//}
	}
	return Result{
		Success: true,
		Message: "User added to database",
	}
}

func getUserIdFromDB(username string) (uint64, Result) {
	printWithName(username, "Searching for field [userid] in database...")
	const query string = "SELECT userid FROM users WHERE username = ?"
	var userID uint64
	err := db.QueryRow(query, username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows { // there is no user with this name
			return 0, Result{
				Success: false,
				Message: noUsernameFoundText(username),
			}
		} else {
			fatalWithName(username, "Error executing SELECT query", err.Error())
		}
	}
	return userID, Result{
		Success: true,
		Message: "Field [userid] retrieved from database",
	}
}

func getUserNameFromDB(userID uint64) (string, Result) {
	printWithID(userID, "Searching for field [username] in database...")
	const query string = "SELECT username FROM users WHERE userid = ?"
	var userName string
	err := db.QueryRow(query, userID).Scan(&userName)
	if err != nil {
		if err == sql.ErrNoRows { // there is no user with this id
			return "", Result{
				Success: false,
				Message: noUserIdFoundText(userID),
			}
		} else {
			fatalWithID(userID, "Error executing SELECT username query", err.Error())
		}
	}
	return userName, Result{
		Success: true,
		Message: "Field [username] retrieved from database",
	}
}

func getPasswordFromDB(userID uint64) ([]byte, Result) {
	printWithID(userID, "Searching for field [password] in database...")
	const query string = "SELECT password FROM users WHERE userid = ?"
	var passwordHash []byte
	err := db.QueryRow(query, userID).Scan(&passwordHash)
	if err != nil {
		if err == sql.ErrNoRows { // there is no user with this id
			return nil, Result{
				Success: false,
				Message: noUserIdFoundText(userID),
			}
		} else {
			fatalWithID(userID, "Error executing SELECT password query", err.Error())
		}
	}
	return passwordHash, Result{
		Success: true,
		Message: "Field [password] retrieved from database",
	}
}

func addTokenDB(tokenRow Token) Result {
	printWithID(tokenRow.UserID, "Adding new token into database...")

	const query string = "INSERT INTO tokens (token, userid, expiration) VALUES (?, ?, ?)"
	_, err := db.Exec(query, tokenRow.Token, tokenRow.UserID, tokenRow.Expiration)
	if err != nil {
		fatalWithID(tokenRow.UserID, "Error executing INSERT token query", err.Error())
		//return Result{
		//	Success: false,
		//	Message: "FATAL: Error adding new user to database",
		//}
	}
	return Result{
		Success: true,
		Message: "Token added to database",
	}
}

func getTokenFromDB(tokenString string) (Token, Result) {
	log.Println("Searching for token in database...")

	const query string = "SELECT * FROM tokens WHERE token = ?"

	var token Token
	var text uint64

	err := db.QueryRow(query, tokenString).Scan(&token.Token, &token.UserID, &text)
	if err != nil {
		if err == sql.ErrNoRows { // token was not found
			return Token{}, Result{
				Success: false,
				Message: "Failure: Token was not found in database",
			}
		} else {
			log.Fatal("Error executing SELECT token query: " + err.Error())
		}
	}
	return token, Result{
		Success: true,
		Message: "Success: Token was found in the database",
	}
}

// func getUserRowDB(userIDArgs uint64) (uint64, string, string, string, string, Result) {
// 	const query string = "SELECT * FROM users WHERE userid = ?"

// 	var (
// 		userID       uint64
// 		username     string
// 		passwordHash string
// 		totpSecret   string
// 		activeTokens string
// 	)

// 	err := db.QueryRow(query, userIDArgs).Scan(&userID, &username, &passwordHash, &totpSecret, &activeTokens)
// 	if err != nil {
// 		if err == sql.ErrNoRows { // there is no user with this name
// 			return 0, "", "", "", "", Result{
// 				Success: false,
// 				Message: noUserIdFoundText(userIDArgs),
// 			}
// 		} else {
// 			log.Fatalf("%s: Error executing SELECT query: %s\n", userIDArgs, err)
// 			return 0, "", "", "", "", Result{
// 				Success: false,
// 				Message: "FATAL: Error searching for user in database",
// 			}
// 		}
// 	}
// 	return userID, username, passwordHash, totpSecret, activeTokens, Result{
// 		Success: true,
// 		Message: "User row was retrieved successfully",
// 	}
// }
