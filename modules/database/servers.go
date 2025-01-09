package database

import (
	log "proto-chat/modules/logging"
	"proto-chat/modules/snowflake"
)

type Server struct {
	ServerID uint64
	UserID   uint64
	Name     string
	Picture  string
}

const insertServerQuery = "INSERT INTO servers (server_id, user_id, name, picture) VALUES (?, ?, ?, ?)"
const deleteServerQuery = "DELETE FROM servers WHERE server_id = ? AND user_id = ?"

func CreateServersTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS servers (
				server_id BIGINT UNSIGNED PRIMARY KEY NOT NULL,
				user_id BIGINT UNSIGNED NOT NULL,
				name TEXT NOT NULL,
				picture TEXT NOT NULL,
				FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
			)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating servers table")
	}
}

func GetServerOwner(serverID uint64) uint64 {
	const query = "SELECT user_id FROM servers WHERE server_id = ?"
	log.Query(query, serverID)

	var ownerID uint64
	err := Conn.QueryRow(query, serverID).Scan(&ownerID)
	DatabaseErrorCheck(err)

	if ownerID == 0 {
		log.Trace("Failed getting owner of server ID [%d]", serverID)
	} else {
		log.Trace("Owner of server ID [%d] is: [%d]", serverID, ownerID)
	}

	return ownerID
}

func AddNewServer(userID uint64, name string, picture string) uint64 {
	tx, err := Conn.Begin()
	transactionErrorCheck(err)

	defer tx.Rollback()

	// insert server
	var serverID uint64 = snowflake.Generate()
	log.Query(insertServerQuery, serverID, userID, name, picture)
	_, err = tx.Exec(insertServerQuery, serverID, userID, name, picture)
	transactionErrorCheck(err)

	// insert default channel
	var channelID uint64 = snowflake.Generate()
	log.Query(insertChannelQuery, channelID, serverID, defaultChannelName)
	_, err = tx.Exec(insertChannelQuery, channelID, serverID, defaultChannelName)
	transactionErrorCheck(err)

	// insert creator as server member
	log.Query(insertServerMemberQuery, serverID, userID)
	_, err = tx.Exec(insertServerMemberQuery, serverID, userID)
	transactionErrorCheck(err)

	err = tx.Commit()
	transactionErrorCheck(err)

	return serverID
}

func ChangeServerPic(userID uint64, serverID uint64, fileName string) bool {
	const query string = "UPDATE servers SET picture = ? WHERE user_id = ? AND server_id = ?"
	log.Query(query, fileName, userID, serverID)

	result, err := Conn.Exec(query, fileName, userID, serverID)
	DatabaseErrorCheck(err)

	rowsAffected, err := result.RowsAffected()
	DatabaseErrorCheck(err)

	if rowsAffected == 1 {
		log.Debug("Updated picture of server ID [%d] in database to [%s]", serverID, fileName)
		return true
	} else {
		log.Debug("Couldn't change picture of server ID [%d] in database to [%s]", serverID, fileName)
		return false
	}
}
