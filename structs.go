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
	Token      string
	UserID     uint64
	Expiration uint64
}
