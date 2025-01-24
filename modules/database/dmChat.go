package database

import (
	log "chat-app/modules/logging"
	"encoding/json"
)

const insertDmChatQuery = "INSERT INTO dm_chats (user1_id, user2_id, dm_id) VALUES (?, ?, ?)"

type DmChat struct {
	UserID1 uint64
	UserID2 uint64
	ChatID  uint64
}

func CreateDmChatTable() {
	// if the chat is just between 2 people, the owner will be 0, otherwise the creator's user id
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS dm_chats (
   			user1_id BIGINT UNSIGNED NOT NULL,
			user2_id BIGINT UNSIGNED NOT NULL,
			dm_id BIGINT UNSIGNED NOT NULL UNIQUE,
			FOREIGN KEY (user1_id) REFERENCES users(user_id) ON DELETE CASCADE,
			FOREIGN KEY (user2_id) REFERENCES users(user_id) ON DELETE CASCADE,
			PRIMARY KEY (user1_id, user2_id),
			CHECK (user1_id != user2_id)
			)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating direct message chats table")
	}
}

func GetDmListOfUser(userID uint64) []byte {
	const query string = "SELECT dm_id from dm_chats WHERE user1_id = ? OR user2_id = ?"
	log.Query(query, userID, userID)

	rows, err := Conn.Query(query, userID, userID)
	DatabaseErrorCheck(err)

	var dmChatIDs []uint64

	for rows.Next() {
		var dmChatID uint64
		DatabaseErrorCheck(rows.Scan(&dmChatID))
		dmChatIDs = append(dmChatIDs, dmChatID)
	}

	if len(dmChatIDs) == 0 {
		log.Trace("User ID [%d] doesn't have any direct messages", userID)
		return emptyArray
	} else {
		jsonBytes, err := json.Marshal(dmChatIDs)
		if err != nil {
			log.Fatal(err.Error(), "Error serializing direct messages of user ID [%d] into json", userID)
		}
		return jsonBytes
	}

}
