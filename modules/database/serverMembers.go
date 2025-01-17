package database

import (
	log "proto-chat/modules/logging"
)

type ServerMemberShort struct {
	ServerID uint64
	UserID   uint64
}

type ServerMember struct {
	UserID     uint64
	Name       string
	Pic        string
	Online     bool
	Status     byte
	StatusText string
}

const insertServerMemberQuery = "INSERT INTO server_members (server_id, user_id) VALUES (?, ?)"
const deleteServerMemberQuery = "DELETE FROM server_members WHERE server_id = ? AND user_id = ?"

func CreateServerMembersTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS server_members (
			server_id BIGINT UNSIGNED NOT NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			FOREIGN KEY (server_id) REFERENCES servers(server_id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
			PRIMARY KEY (server_id, user_id)
		)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating server_members table")
	}
}

func GetServerMembersList(serverID uint64) []ServerMember {
	const query = "SELECT u.user_id, u.display_name, u.picture, u.status, u.status_text FROM users u JOIN server_members sm ON u.user_id = sm.user_id WHERE sm.server_id = ?"
	log.Query(query, serverID)

	rows, err := Conn.Query(query, serverID)
	DatabaseErrorCheck(err)

	var members []ServerMember
	for rows.Next() {
		var m ServerMember

		err := rows.Scan(&m.UserID, &m.Name, &m.Pic, &m.Status, &m.StatusText)
		log.Trace("test: ", m.StatusText)
		DatabaseErrorCheck(err)

		members = append(members, m)
	}
	DatabaseErrorCheck(rows.Err())

	if len(members) == 0 {
		log.Hack("Server ID [%d] has no members", serverID)
	}

	log.Trace("Members of server ID [%d] were retrieved successfully", serverID)
	return members
}
func ConfirmServerMembership(userID uint64, serverID uint64) bool {
	const query string = "SELECT EXISTS (SELECT 1 FROM server_members WHERE server_id = ? AND user_id = ?)"
	log.Query(query, serverID, userID)

	var isMember bool = false
	err := Conn.QueryRow(query, serverID, userID).Scan(&isMember)
	DatabaseErrorCheck(err)

	if isMember {
		log.Trace("User ID [%d] is a member of server ID [%d]", userID, serverID)
	} else {
		log.Hack("User ID [%d] is not a member of server ID [%d]", userID, serverID)
	}

	return isMember
}

func GetJoinedServersList(userID uint64) []uint64 {
	// get what servers are the user part of, so message will broadcast to members of these servers
	// this makes sure users who don't have visual on the user who receive the changes
	const query string = "SELECT s.server_id FROM servers s JOIN server_members m ON s.server_id = m.server_id WHERE m.user_id = ?"
	log.Query(query, userID)

	rows, err := Conn.Query(query, userID)
	DatabaseErrorCheck(err)

	var serverIDs []uint64
	for rows.Next() {
		var serverID uint64
		err := rows.Scan(&serverID)
		DatabaseErrorCheck(err)
		serverIDs = append(serverIDs, serverID)
	}

	if len(serverIDs) == 0 {
		log.Hack("User ID [%d] is not in any servers", userID)
	} else {
		log.Trace("Successfully retrieved list of servers where user ID [%d] is joined", userID)
	}
	return serverIDs
}
