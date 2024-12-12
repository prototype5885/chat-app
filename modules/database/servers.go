package database

import (
	"encoding/json"
	log "proto-chat/modules/logging"
	"proto-chat/modules/snowflake"
)

type Server struct {
	ServerID uint64
	UserID   uint64
	Name     string
	Picture  string
}

type ServerDeletion struct {
	ServerID uint64
	UserID   uint64
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
func GetServerList(userID uint64) []byte {
	log.Trace("Getting joined server list of user ID [%d]...", userID)

	const query = "SELECT s.* FROM servers s JOIN server_members m ON s.server_id = m.server_id WHERE m.user_id = ?"

	rows, err := Conn.Query(query, userID)
	DatabaseErrorCheck(err)
	var servers []Server
	for rows.Next() {
		var server Server
		err := rows.Scan(&server.ServerID, &server.UserID, &server.Name, &server.Picture)
		DatabaseErrorCheck(err)
		servers = append(servers, server)
	}

	if len(servers) == 0 {
		log.Trace("User ID [%d] is not in any servers", userID)
		return nullJson
	}

	jsonResult, _ := json.Marshal(servers)
	return jsonResult
}
func GetServerOwner(serverID uint64) uint64 {
	log.Trace("Getting owner of server ID [%d]...", serverID)

	const query = "SELECT user_id FROM servers WHERE server_id = ?"

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

func AddNewServer(userID uint64, name string, picture string) Server {
	var serverID uint64 = snowflake.Generate()

	var server = Server{
		ServerID: serverID,
		UserID:   userID,
		Name:     name,
		Picture:  picture,
	}

	// err := Conn.Transaction(func(tx *gorm.DB) error {
	// 	var err error
	// 	// create server
	// 	err = tx.Create(&server).Error
	// 	DatabaseErrorCheck(err)

	// 	// create channel
	// 	var channel = Channel{
	// 		ChannelID: snowflake.Generate(),
	// 		ServerID:  server.ServerID,
	// 		Name:      "Default Channel",
	// 	}

	// 	err = tx.Create(&channel).Error
	// 	DatabaseErrorCheck(err)

	// 	// add owner user to server
	// 	var member = ServerMember{
	// 		ServerID: server.ServerID,
	// 		UserID:   userID,
	// 	}

	// 	err = tx.Create(&member).Error

	// 	return err
	// })
	// DatabaseErrorCheck(err)

	return server
}
