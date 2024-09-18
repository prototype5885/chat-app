package main

type ConfigFile struct {
	Port          uint32
	Sqlite        bool
	TLS           bool
	LocalhostOnly bool
	LogInFile     bool
	Username      string
	Password      string
	Address       string
	DatabasePort  uint32
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
