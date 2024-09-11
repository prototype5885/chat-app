package main

type ConfigFile struct {
	Sqlite       bool
	LogInFile    bool
	Username     string
	Password     string
	Address      string
	Port         uint
	DatabaseName string
}

type Result struct {
	Success bool
	Message string
}

type Token struct {
	Token      []byte
	UserID     uint64
	Expiration uint64
}

type Server struct {
	ServerID      uint64
	ServerOwnerID uint64
	ServerName    string
}

type Channel struct {
	ChannelID   uint64
	ServerID    uint64
	ChannelName string
}

type ServerChatMessage struct {
	MessageID uint64
	ChannelID uint64
	UserID    uint64
	Username  string
	Message   string
}

type ServerChatMessages struct {
	Messages []ServerChatMessage
}
