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

func (c *Channels) CreateChannelsTable() {
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

func (c *Channels) GetChatMessages(channelID uint64) *sql.Rows {
	const query string = "SELECT message_id, user_id, message FROM messages WHERE channel_id = ?"

	rows, err := db.Query(query, channelID)
	if err != nil {
		log.FatalError(err.Error(), "Error searching for messages on channel ID [%d]", channelID)
	}
	return rows
}

func (c *Channels) GetChannelList(serverID uint64) *sql.Rows {
	log.Debug("Getting channel list of server ID [%d]...", serverID)
	const query string = "SELECT channel_id, name FROM channels WHERE server_id = ?"

	rows, err := db.Query(query, serverID)
	if err != nil {
		log.FatalError(err.Error(), "Error searching for channels list of server ID [%d]", serverID)
	}
	return rows
}
