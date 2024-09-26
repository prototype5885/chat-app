package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"proto-chat/modules/database"
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
			log.Error(err.Error())
			log.Fatal("Error opening config file")

		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Error(err.Error())
				log.Fatal("Error closing config file")
			}
		}(file)

		var config ConfigFile
		err = json.NewDecoder(file).Decode(&config)
		if err != nil {
			log.Error(err.Error())
			log.Fatal("Error decoding config file")
		}
		return config
	}

	config := readConfigFile()
	log.SetupLogging("TRACE", config.LogConsole, config.LogFile)

	// database
	if config.Sqlite {
		database.ConnectSqlite()
	} else {
		database.ConnectMariadb(config.DatabaseUsername, config.DatabasePassword, config.DatabaseAddress, strconv.Itoa(int(config.DatabasePort)), config.DatabaseName)
	}
	database.CreateTables()

	// snowflake
	snowflake.SetSnowflakeServerID(0)

	// database.Insert(database.Channel{
	// 	ChannelID: snowflake.Generate(),
	// 	ServerID:  snowflake.Generate(),
	// 	Name:      "test channel name",
	// })

	// start := macros.GetTimestamp()
	// for i := 0; i < 1000; i++ {
	// 	// log.Info("Starting server...")
	// 	database.Insert(database.ChatMessage{
	// 		MessageID: snowflake.Generate(),
	// 		ChannelID: 1811203793171251200,
	// 		UserID:    1810997960123613184,
	// 		Message:   "test message",
	// 	})

	// }
	// macros.MeasureTime(start, "test")

	// database.Insert(database.ChatMessage{
	// 	MessageID: snowflake.Generate(),
	// 	ChannelID: 1811203793171251200,
	// 	UserID:    1810997960123613184,
	// 	Message:   "test message",
	// })
	// return

	// websocket
	websocket.Init()

	var wsType string
	if config.TLS {
		wsType = "/wss"
	} else {
		wsType = "/ws"
	}
	http.HandleFunc(wsType, func(w http.ResponseWriter, r *http.Request) {
		webRequests.WssHandler(w, r)
	})

	// http requests

	http.HandleFunc("GET /login-register.html", webRequests.LoginRegisterHandler)
	http.HandleFunc("GET /chat.html", webRequests.ChatHandler)

	http.HandleFunc("POST /login", webRequests.PostRequestHandler)
	http.HandleFunc("POST /register", webRequests.PostRequestHandler)

	http.HandleFunc("/", webRequests.MainHandler)

	var address string

	if config.LocalhostOnly {
		address = fmt.Sprintf("%s:%d", "127.0.0.1", config.Port)
	} else {
		address = fmt.Sprintf("%s:%d", "0.0.0.0", config.Port)
	}

	log.Info("Listening on port %d", config.Port)
	if config.TLS {
		const certFile = "./sslcert/cert.crt"
		const keyFile = "./sslcert/key.key"
		if err := http.ListenAndServeTLS(address, certFile, keyFile, nil); err != nil {
			log.Error(err.Error())
			log.Fatal("Error starting TLS server")
		}
	} else {
		if err := http.ListenAndServe(address, nil); err != nil {
			log.Error(err.Error())
			log.Fatal("Error starting server")
		}
	}
}
