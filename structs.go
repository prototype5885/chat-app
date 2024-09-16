package main

type ConfigFile struct {
	Sqlite        bool
	TLS           bool
	LocalhostOnly bool
	LogInFile     bool
	Username      string
	Password      string
	Address       string
	Port          uint32
	DatabaseName  string
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
	ServerID uint64
	OwnerID  uint64
	Name     string
	Picture  string
}

type Channel struct {
	ChannelID uint64
	ServerID  uint64
	Name      string
}

type ServerChatMessage struct {
	MessageID uint64
	ChannelID uint64
	UserID    uint64
	Username  string
	Message   string
}

type ServerForClient struct { // this is whats sent to the client when client requests server list
	ServerID uint64
	Name     string
	Picture  string
}
