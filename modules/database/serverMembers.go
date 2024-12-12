package database

import (
	log "proto-chat/modules/logging"
)

type ServerMember struct {
	ServerID uint64
	UserID   uint64
}

type MemberInfo struct {
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
			UNIQUE (server_id, user_id)
		)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating server_members table")
	}
}

func GetServerMembersList(serverID uint64, userID uint64) []MemberInfo {
	log.Trace("Getting list of members of server ID [%d] for user ID [%d]...", serverID, userID)

	const query = "SELECT u.user_id, u.display_name, u.picture, u.status, u.status_text FROM users u JOIN server_members sm ON u.user_id = sm.user_id WHERE sm.server_id = ?"

	rows, err := Conn.Query(query, serverID)
	DatabaseErrorCheck(err)

	var members []MemberInfo
	for rows.Next() {
		var m MemberInfo

		err := rows.Scan(&m.UserID, &m.Name, &m.Pic, &m.Status, &m.StatusText)
		DatabaseErrorCheck(err)

		members = append(members, m)
	}

	if len(members) == 0 {
		log.Hack("Server ID [%d] has no members", serverID)
	}

	log.Trace("Members of server ID [%d] for user ID [%d] were retrieved successfully", serverID, userID)

	return members
}
func ConfirmServerMembership(userID uint64, serverID uint64) bool {
	log.Trace("Searching for user ID [%d] in server ID [%d]...", userID, serverID)

	const query string = "SELECT EXISTS (SELECT 1 FROM server_members WHERE server_id = ? AND user_id = ?)"

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
	log.Trace("Getting list of server IDs where user ID [%d] is joined", userID)

	const query string = `
		SELECT s.server_id
		FROM servers s
		JOIN server_members m ON s.server_id = m.server_id
		WHERE m.user_id = ?
		`

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
