package database

import (
	log "proto-chat/modules/logging"
)

type ServerMember struct {
	ServerID uint64
	UserID   uint64
}

const (
	insertServerMemberQuery = "INSERT INTO server_members (server_id, user_id) VALUES (?, ?)"
	deleteServerMemberQuery = "DELETE FROM server_members WHERE server_id = ? AND user_id = ?"
)

func CreateServerMembersTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS server_members (
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

func GetServerMembersList(serverID uint64, userID uint64) []byte {
	log.Trace("Getting list of members of server ID [%d]...", serverID)

	//const query string = `
	//	SELECT u.user_id, u.display_name, u.picture
	//	FROM users u
	//	JOIN server_members sm ON u.user_id = sm.user_id
	//	WHERE sm.server_id = ?
	//`

	const query string = `
		SELECT JSON_ARRAYAGG(JSON_OBJECT(
			'UserID', CAST(user_id AS CHAR),
			'Name', display_name,
			'Pic', picture,
		    'Status', status,
		    'StatusText', status_text
		)) AS json_result
		FROM (
			SELECT u.user_id, u.display_name, u.picture, u.status, u.status_text
			FROM users u
			JOIN server_members sm ON u.user_id = sm.user_id 
			WHERE sm.server_id = ?
		) AS members_chunk;
	`

	var jsonResult []byte
	err := db.QueryRow(query, serverID).Scan(&jsonResult)
	if err != nil {
		log.FatalError(err.Error(), "Error getting server member list of server ID [%d] for user ID [%d]", serverID, userID)
	}

	if len(jsonResult) == 0 {
		log.Warn("Server ID [%d] does not have any members", serverID)
		return nullJson
	}

	return jsonResult

	//rows, err := db.Query(query, serverID)
	//if err != nil {
	//	log.FatalError(err.Error(), "Error searching for members in server ID [%d]", serverID)
	//}
	//var userInfos []structs.ServerMemberListResponse
	//
	//var counter int = 0
	//for rows.Next() {
	//	counter++
	//	var userInfo = structs.ServerMemberListResponse{Status: "custom status"}
	//	err := rows.Scan(&userInfo.UserID, &userInfo.Name, &userInfo.Picture)
	//	if err != nil {
	//		log.FatalError(err.Error(), "Error scanning server member row into userID of server ID [%d]:", serverID)
	//	}
	//	userInfos = append(userInfos, userInfo)
	//}
	//
	//if counter == 0 {
	//	log.Debug("Server ID [%d] doesn't have any members", serverID)
	//	return userInfos
	//}
	//
	//log.Trace("Members of server ID [%d] were retrieved successfully", serverID)
	//return userInfos
}

func ConfirmServerMembership(userID uint64, serverID uint64) bool {
	log.Trace("Searching for user ID [%d] in server ID [%d]...", userID, serverID)

	const query string = "SELECT EXISTS (SELECT 1 FROM server_members WHERE server_id = ? AND user_id = ?)"

	var isMember bool = false

	err := db.QueryRow(query, serverID, userID).Scan(&isMember)
	if err != nil {
		// sql.ErrNoRows won't happen here because it returns a bool
		log.FatalError(err.Error(), "Error checking if user ID [%d] is member of server ID [%d]", userID, serverID)
	}

	if isMember {
		log.Trace("User ID [%d] is a member of server ID [%d]", userID, serverID)
	} else {
		log.Hack("User ID [%d] is not a member of server ID [%d]", userID, serverID)
	}
	return isMember
}

func GetJoinedServersList(userID uint64) ([]byte, bool) {
	log.Trace("Getting list of server IDs where user ID [%d] is joined", userID)

	const query string = `
		SELECT JSON_ARRAYAGG(s.server_id) AS json_result
		FROM servers s
		JOIN server_members m ON s.server_id = m.server_id 
		WHERE m.user_id = ?
		`

	var jsonResult []byte
	err := db.QueryRow(query, userID).Scan(&jsonResult)
	if err != nil {
		log.FatalError(err.Error(), "Error getting server list of user ID [%d]", userID)
		return nil, false
	}

	if len(jsonResult) == 0 {
		return nil, true
	}

	return jsonResult, false
}
