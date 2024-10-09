package database

import (
	"database/sql"
	log "proto-chat/modules/logging"
)

type ServerInvite struct {
	InviteID   uint64
	ServerID   uint64
	SingleUse  bool
	Expiration uint64
}

const (
	insertServerInviteQuery = "INSERT INTO server_invites (invite_id, server_id, single_use, expiration) VALUES (?, ?, ?, ?)"
	deleteServerInviteQuery = "DELETE FROM server_invites WHERE invite_id = ?"
)

func CreateServerInvitesTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS server_invites (
		invite_id BIGINT UNSIGNED PRIMARY KEY NOT NULL,
		server_id BIGINT UNSIGNED NOT NULL,
		single_use BOOLEAN NOT NULL,
		expiration BIGINT UNSIGNED NOT NULL,
		FOREIGN KEY (server_id) REFERENCES servers(server_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating server_invites table")
	}
}

func ConfirmServerInviteID(inviteID uint64) uint64 {
	log.Debug("Searching for server invite ID in database...")

	const query string = "SELECT server_id FROM server_invites WHERE invite_id = ?"

	var serverID uint64

	err := db.QueryRow(query, inviteID).Scan(&serverID)
	if err != nil {
		log.Error(err.Error())
		if err == sql.ErrNoRows { // invite id was not found
			log.Debug("Invite ID [%d] was not found in database", inviteID)
			return 0
		}
		log.Fatal("Error retrieving invite ID [%d] from database", inviteID)
		return 0
	}
	log.Debug("Invite ID [%d] was found in database, it belongs to server ID [%d]", inviteID, serverID)
	return serverID
}
