package database

import (
	log "proto-chat/modules/logging"
	"proto-chat/modules/structs"
)

type Channel struct {
	ChannelID uint64
	ServerID  uint64
	Name      string
}

const insertChannelQuery string = "INSERT INTO channels (channel_id, server_id, name) VALUES (?, ?, ?)"

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

func (c *Channels) GetChannelList(serverID uint64) []structs.ChannelResponse {
	log.Debug("Getting channel list of server ID [%d] from database...", serverID)
	const query string = "SELECT channel_id, name FROM channels WHERE server_id = ?"

	rows, err := db.Query(query, serverID)
	if err != nil {
		log.FatalError(err.Error(), "Error searching for channels list of server ID [%d]", serverID)
	}

	var channels []structs.ChannelResponse

	var counter int = 0
	for rows.Next() {
		counter++
		var channel = structs.ChannelResponse{}
		err := rows.Scan(&channel.ChannelID, &channel.Name)
		if err != nil {
			log.FatalError(err.Error(), "Error scanning channel row into struct from server ID [%d]:", serverID)
		}
		channels = append(channels, channel)
	}

	if counter == 0 {
		log.Debug("Server ID [%d] doesn't have any channels", serverID)
	} else {
		log.Debug("Channels from server ID [%d] were retrieved successfully", serverID)
	}

	return channels
}
