package database

import (
	"database/sql"
	log "proto-chat/modules/logging"
	"proto-chat/modules/snowflake"
	"proto-chat/modules/structs"
)

type Server struct {
	ServerID uint64
	OwnerID  uint64
	Name     string
	Picture  string
}

const insertServerQuery string = "INSERT INTO servers (server_id, owner_id, name, picture) VALUES (?, ?, ?, ?)"

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

func (s *Servers) GetServerList(userID uint64) []structs.ServerResponse {
	log.Debug("Getting server list of user ID [%d]...", userID)
	const query string = `
		SELECT s.*
		FROM servers s
		JOIN server_members m ON s.server_id = m.server_id 
		WHERE m.user_id = ?
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		log.FatalError(err.Error(), "Error searching for server list of user ID [%d]", userID)
	}

	var servers []structs.ServerResponse

	var counter int = 0
	for rows.Next() {
		counter++
		var server = structs.ServerResponse{}
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

func (s *Servers) AddNewServer(userID uint64, name string, picture string) Server {
	tx, err := db.Begin()
	if err != nil {

	}

	// insert server
	var serverID uint64 = snowflake.Generate()

	var server = Server{
		ServerID: serverID,
		OwnerID:  userID,
		Name:     name,
		Picture:  picture,
	}

	log.Debug("Inserting new server ID [%d] on the request of user ID [%d]", server.ServerID, userID)
	_, err = tx.Exec(insertServerQuery, server.ServerID, server.OwnerID, server.Name, server.Picture)
	if err != nil {
		log.Error(err.Error())
		tx.Rollback()

	}

	// insert channel
	var channel = Channel{
		ChannelID: snowflake.Generate(),
		ServerID:  server.ServerID,
		Name:      "Default Channel",
	}

	log.Debug("Inserting default channel for server ID [%d]", server.ServerID)
	_, err = tx.Exec(insertChannelQuery, channel.ChannelID, channel.ServerID, channel.Name)
	if err != nil {
		log.Error(err.Error())
		tx.Rollback()
	}

	// insert member
	var member = ServerMember{
		ServerID: server.ServerID,
		UserID:   userID,
	}

	log.Debug("Adding server owner ID [%d] into server ID [%d] as member", userID, server.ServerID)
	_, err = tx.Exec(insertServerMemberQuery, member.ServerID, member.UserID)
	if err != nil {
		log.Error(err.Error())
		tx.Rollback()
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.FatalError(err.Error(), "Error adding new server")
	}
	return server
}
