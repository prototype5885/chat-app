package database

import (
	"encoding/json"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
)

type Channel struct {
	ChannelID uint64
	ServerID  uint64
	Name      string
}

const insertChannelQuery = "INSERT INTO channels (channel_id, server_id, name) VALUES (?, ?, ?)"

func CreateChannelsTable() {
	_, err := Conn.Exec(`CREATE TABLE IF NOT EXISTS channels (
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
	log.Trace("Getting channels of server ID [%d] from database...", serverID)

	const query string = "SELECT * FROM channels WHERE server_id = ?"

	rows, err := Conn.Query(query, serverID)
	DatabaseErrorCheck(err)

	var channels []Channel
	for rows.Next() {
		var channel Channel
		err := rows.Scan(&channel.ChannelID, &channel.ServerID, &channel.Name)
		DatabaseErrorCheck(err)
		channels = append(channels, channel)
	}

	if len(channels) == 0 {
		log.Trace("Server ID [%d] does't have any channels")
		return nullJson
	}

	jsonResult, err := json.Marshal(channels)
	if err != nil {
		macros.ErrorSerializing(err.Error(), "channel list", serverID)
	}

	return jsonResult
}

func GetServerIdOfChannel(channelID uint64) uint64 {
	log.Trace("Getting which server channel ID [%d] belongs to", channelID)

	const query = "SELECT server_id FROM channels WHERE channel_id = ?"

	var serverID uint64
	err := Conn.QueryRow(query, channelID).Scan(&serverID)
	DatabaseErrorCheck(err)

	if serverID == 0 {
		log.Trace("Channel ID [%d] does not belong to any server", channelID)
	} else {
		log.Trace("Channel ID [%d] belongs to server ID [%d]", channelID, serverID)
	}

	return serverID
}
