package database

import (
	log "proto-chat/modules/logging"
)

type ServerInvite struct {
	InviteID     uint64
	ServerID     uint64
	TargetUserID uint64
	SingleUse    bool
	Expiration   uint64
}

type ServerInviteDelete struct {
	InviteID uint64
}

const insertServerInviteQuery = "INSERT INTO server_invites (invite_id, server_id, target_user_id, single_use, expiration) VALUES (?, ?, ?, ?, ?)"
const deleteServerInviteQuery = "DELETE FROM server_invites WHERE invite_id = ?"

func CreateServerInvitesTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS server_invites (
		invite_id BIGINT UNSIGNED PRIMARY KEY NOT NULL,
		server_id BIGINT UNSIGNED NOT NULL,
		target_user_id BIGINT UNSIGNED NOT NULL,
		single_use BOOLEAN NOT NULL,
		expiration BIGINT UNSIGNED NOT NULL,
		FOREIGN KEY (server_id) REFERENCES servers(server_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating server_invites table")
	}
}

func ConfirmServerInviteID(inviteID uint64) (uint64, bool, uint64) {
	const query string = "SELECT server_id, single_use, expiration FROM server_invites WHERE invite_id = ?"
	log.Query(query, inviteID)

	var serverID uint64
	var singleUse bool
	var expiration uint64
	err := Conn.QueryRow(query, inviteID).Scan(&serverID, &singleUse, &expiration)
	DatabaseErrorCheck(err)

	if serverID == 0 {
		log.Debug("Invite ID [%d] was not found in database", inviteID)
	} else {
		log.Debug("Invite ID [%d] was found in database, it belongs to server ID [%d]", inviteID, serverID)
	}

	return serverID, singleUse, expiration
}
