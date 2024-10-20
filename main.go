package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"proto-chat/modules/database"
	jsfilesmerger "proto-chat/modules/jsFilesMerger"
	log "proto-chat/modules/logging"
	"proto-chat/modules/snowflake"
	"proto-chat/modules/webRequests"
	"proto-chat/modules/websocket"
	"strconv"
	"syscall"
)

func main() {
	fmt.Println("Starting server...")

	// handle termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Received termination signal...")
		err := database.CloseDatabaseConnection()
		if err != nil {
			log.Error("Error closing db connection")
		}
		fmt.Println("Closed main db connection successfully")
		os.Exit(0)
	}()

	// reading config file

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

	readConfigFile := func() ConfigFile {
		configFile := "config.json"
		file, err := os.Open(configFile)
		if err != nil {
			log.FatalError(err.Error(), "Error opening config file")
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.FatalError(err.Error(), "Error closing config file")
			}
		}(file)

		var config ConfigFile
		err = json.NewDecoder(file).Decode(&config)
		if err != nil {
			log.FatalError(err.Error(), "Error decoding config file")
		}
		return config
	}

	config := readConfigFile()
	log.SetupLogging("DEBUG", config.LogConsole, config.LogFile)

	jsfilesmerger.Init()

	// database
	if config.Sqlite {
		database.ConnectSqlite()
	} else {
		database.ConnectMariadb(config.DatabaseUsername, config.DatabasePassword, config.DatabaseAddress, strconv.Itoa(int(config.DatabasePort)), config.DatabaseName)
	}
	database.CreateTables()

	// snowflake
	snowflake.SetSnowflakeWorkerID(0)

	// websocket
	websocket.Init()

	// handle http requests
	http.HandleFunc("/", webRequests.MainHandler)

	var address string

	if config.LocalhostOnly {
		address = fmt.Sprintf("%s:%d", "127.0.0.1", config.Port)
	} else {
		address = fmt.Sprintf("%s:%d", "0.0.0.0", config.Port)
	}

	if config.TLS {
		const certFile = "./sslcert/cert.crt"
		const keyFile = "./sslcert/key.key"

		log.Info("Listening on https://%s", address)
		if err := http.ListenAndServeTLS(address, certFile, keyFile, nil); err != nil {
			log.FatalError(err.Error(), "Error starting TLS server")
		}
	} else {
		log.Info("Listening on http://%s", address)
		if err := http.ListenAndServe(address, nil); err != nil {
			log.FatalError(err.Error(), "Error starting non-TLS server")
		}
	}
}
