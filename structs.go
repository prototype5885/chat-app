package main

type ConfigFile struct {
	LocalhostOnly    bool
	Port             uint32
	TLS              bool
	LogConsole       bool
	LogFile          bool
	Sqlite           bool
	DatabaseAddress  string
	DatabasePort     uint32
	DatabaseUsername string
	DatabasePassword string
	DatabaseName     string
}

type Result struct {
	Success bool
	Message string
}

type BroadcastData struct {
	MessageBytes []byte
	Type         byte
	ID           uint64
}

type Token struct {
	Token      []byte
	UserID     uint64
	Expiration uint64
}
