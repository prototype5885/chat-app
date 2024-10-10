package database

import (
	"database/sql"
	log "proto-chat/modules/logging"
)

type Channel struct {
	ChannelID uint64
	ServerID  uint64
	Name      string
}

const (
	insertChannelQuery = "INSERT INTO channels (channel_id, server_id, name) VALUES (?, ?, ?)"
	deleteChannelQuery = "DELETE FROM channels WHERE channel_id = ?"
)

func CreateChannelsTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS channels (
		channel_id BIGINT UNSIGNED PRIMARY KEY NOT NULL,
		server_id BIGINT UNSIGNED NOT NULL,
		name TEXT NOT NULL,
		FOREIGN KEY (server_id) REFERENCES servers(server_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating channels table")
	}
}

func GetChannelList(serverID uint64) []byte {
	log.Debug("Getting channel list of server ID [%d] from database...", serverID)

	const query string = `
		SELECT JSON_ARRAYAGG(JSON_OBJECT(
            'ChannelID', CAST(channel_id AS CHAR),
            'Name', name
        )) AS json_result
        FROM channels
        WHERE server_id = ?
	`

	var jsonResult []byte
	err := db.QueryRow(query, serverID).Scan(&jsonResult)
	if err != nil {
		log.FatalError(err.Error(), "Error getting channel list of server ID [%d]", serverID)
	}

	if len(jsonResult) == 0 {
		return nullJson
	}

	return jsonResult
}

func GetServerOfChannel(channelID uint64) uint64 {
	var serverID uint64
	log.Trace("Getting which server channel ID [%d] belongs to", channelID)
	err := db.QueryRow("SELECT server_id FROM channels WHERE channel_id = ?", channelID).Scan(&serverID)
	if err != nil {
		log.Error(err.Error())
		if err == sql.ErrNoRows {
			log.Warn("Channel ID [%d] doesn't belong to any server", channelID)
			return 0
		}
		log.Fatal("Error getting which server channel ID [%d] belongs to", channelID)
	}
	return serverID
}
