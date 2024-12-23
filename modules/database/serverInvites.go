package database

import (
	log "proto-chat/modules/logging"
	"time"
)

type ServerInvite struct {
	InviteID   uint64 `gorm:"primaryKey;not null"`
	ServerID   uint64 `gorm:"not null"`
	SingleUse  bool   `gorm:"not null"`
	Expiration uint64 `gorm:"not null"`
}

const insertServerInviteQuery = "INSERT INTO server_invites (invite_id, server_id, single_use, expiration)"

func CreateServerInvitesTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS server_invites (
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
	start := time.Now().UnixMicro()
	const query string = "SELECT server_id FROM server_invites WHERE invite_id = ?"
	log.Query(query, inviteID)

	var serverID uint64
	err := Conn.QueryRow(query, inviteID).Scan(&serverID)
	DatabaseErrorCheck(err)

	if serverID == 0 {
		log.Debug("Invite ID [%d] was not found in database", inviteID)
	} else {
		log.Debug("Invite ID [%d] was found in database, it belongs to server ID [%d]", inviteID, serverID)
	}

	measureDbTime(start)
	return serverID
}
