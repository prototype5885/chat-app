package database

import (
	log "proto-chat/modules/logging"
	"proto-chat/modules/structs"
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
		server_id BIGINT UNSIGNED PRIMARY KEY,
		owner_id BIGINT UNSIGNED,
		name TEXT,
		picture TEXT,
		FOREIGN KEY (owner_id) REFERENCES users(user_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating servers table")
	}
}

func (s *Servers) GetServerList(userID uint64) []structs.ServerResponse {
	log.Debug("Getting server list of user ID [%d]...", userID)
	const query string = "SELECT server_id, name, picture FROM servers"

	rows, err := db.Query(query)
	if err != nil {
		log.FatalError(err.Error(), "Error searching for server list of user ID [%d]", userID)
	}

	var servers []structs.ServerResponse

	var counter int = 0
	for rows.Next() {
		counter++
		var server = structs.ServerResponse{}
		err := rows.Scan(&server.ServerID, &server.Name, &server.Picture)
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
