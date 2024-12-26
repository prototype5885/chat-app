package database

import (
	log "proto-chat/modules/logging"
	"time"
)

type Friendship struct {
	FirstUserID  uint64
	SecondUserID uint64
	FriendsSince int64
}

// const insertFriendshipQuery string = `
// 		INSERT INTO friendships (user1_id, user2_id, friends_since)
// 		SELECT ?, ?, NOW()
// 		WHERE NOT EXISTS (
// 			SELECT 1
// 			FROM friendships
// 			WHERE (user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)
// 		);`

const insertFriendshipQuery string = "INSERT INTO friendships (user1_id, user2_id, friends_since) VALUES (?, ?, ?)"
const deleteFriendshipQuery string = "DELETE FROM friendships WHERE MIN(user1_id, user2_id) = ? AND MAX(user1_id, user2_id) = ?"

func CreateFriendshipsTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS friendships (
			user1_id BIGINT UNSIGNED NOT NULL,
			user2_id BIGINT UNSIGNED NOT NULL,
			pending BOOLEAN NOT NULL DEFAULT FALSE,
			friends_since DATE NOT NULL DEFAULT 0,
			FOREIGN KEY (user1_id) REFERENCES users (user_id) ON DELETE CASCADE,
			FOREIGN KEY (user2_id) REFERENCES users (user_id) ON DELETE CASCADE,
			PRIMARY KEY (user1_id, user2_id),
			CHECK (user1_id != user2_id)
		)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating friendships table")
	}
}

func CheckIfFriends(userID uint64, targetUserID uint64) bool {
	start := time.Now().UnixMicro()
	const query string = `
	SELECT EXISTS (
		SELECT 1
		FROM friendships 
		WHERE (user1_id = ? AND user2_id = ?) 
		OR (user1_id = ? AND user2_id = ?)
	)`
	log.Query(query, userID, targetUserID, targetUserID, userID)

	var areFriends bool
	err := Conn.QueryRow(query, userID, targetUserID, userID, targetUserID).Scan(&areFriends)
	DatabaseErrorCheck(err)

	if !areFriends {
		log.Debug("User ID [%d] is not friends with user ID [%d]", targetUserID, userID)
	} else {
		log.Debug("User ID [%d] is friends with user ID [%d]", targetUserID, userID)
	}

	measureDbTime(start)
	return areFriends
}
