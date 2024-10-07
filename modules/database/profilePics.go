package database

import log "proto-chat/modules/logging"

func (p *ProfilePics) CreateProfilePicsTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS profilepics (
		hash BINARY(32) PRIMARY KEY NOT NULL,
		file_name TEXT NOT NULL
	)`)
	if err != nil {
		log.FatalError(err.Error(), "Error creating profilepics table")
	}
}
