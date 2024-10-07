package database

import log "proto-chat/modules/logging"

type ServerMember struct {
	ServerID uint64
	UserID   uint64
}

func (sm *ServerMembers) CreateServerMembersTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS server_members (
		server_id BIGINT UNSIGNED PRIMARY KEY NOT NULL,
		user_id BIGINT UNSIGNED NOT NULL,
		FOREIGN KEY (server_id) REFERENCES servers(server_id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating server_members table")
	}
}

func (sm *ServerMembers) GetServerMembersList(serverID uint64) []uint64 {
	log.Debug("Getting list of members of server ID [%d]...", serverID)
	const query string = "SELECT user_id FROM server_members WHERE server_id = ?"

	rows, err := db.Query(query, serverID)
	if err != nil {
		log.FatalError(err.Error(), "Error searching for members in server ID [%d]", serverID)
	}
	var userIDs []uint64

	var counter int = 0
	for rows.Next() {
		counter++
		var userID uint64
		err := rows.Scan(&userID)
		if err != nil {
			log.FatalError(err.Error(), "Error scanning server member row into userID of server ID [%d]:", userID)
		}
		userIDs = append(userIDs, userID)
	}

	if counter == 0 {
		log.Debug("Server ID [%d] doesn't have any members", serverID)
		return userIDs
	}

	log.Debug("Members of server ID [%d] were retrieved successfully", serverID)
	return userIDs
}
