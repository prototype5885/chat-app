package database

import (
	"database/sql"
	log "proto-chat/modules/logging"
)

type Server struct {
	ServerID uint64
	OwnerID  uint64
	Name     string
	Picture  string
}

type ServerDeletion struct {
	ServerID uint64
	UserID   uint64
}

func (s *Servers) CreateServersTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS servers (
		server_id BIGINT UNSIGNED PRIMARY KEY NOT NULL,
		owner_id BIGINT UNSIGNED NOT NULL,
		name TEXT NOT NULL,
		picture TEXT NOT NULL,
		FOREIGN KEY (owner_id) REFERENCES users(user_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating servers table")
	}
}

func (s *Servers) GetServerList(userID uint64) []Server {
	log.Debug("Getting server list of user ID [%d]...", userID)
	const query string = "SELECT * FROM servers"

	rows, err := db.Query(query)
	if err != nil {
		log.FatalError(err.Error(), "Error searching for server list of user ID [%d]", userID)
	}

	var servers []Server

	var counter int = 0
	for rows.Next() {
		counter++
		var server = Server{}
		err := rows.Scan(&server.ServerID, &server.OwnerID, &server.Name, &server.Picture)
		if err != nil {
			log.FatalError(err.Error(), "Error scanning server row into struct for user ID [%d]:", userID)
		}
		servers = append(servers, server)
	}

	if counter == 0 {
		log.Debug("User ID [%d] is not in any servers", userID)
		return servers
	}

	log.Debug("Servers for user ID [%d] were retrieved successfully", userID)
	return servers
}

func (s *Servers) GetServerOwner(serverID uint64) uint64 {
	log.Debug("Getting owner of server ID [%d]...", serverID)
	const query string = "SELECT owner_ID FROM servers WHERE server_id = ?"

	var ownerID uint64

	err := db.QueryRow(query, serverID).Scan(&ownerID)
	if err != nil {
		log.Error(err.Error())
		if err == sql.ErrNoRows { // token was not found
			log.Debug("Given server ID does not exist: [%d]", serverID)
			return 0
		}
		log.Fatal("Error getting owner of server ID [%d]", serverID)
	}
	log.Debug("Owner of server ID [%d] is: [%d]", serverID, ownerID)

	return ownerID
}
