package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	log "proto-chat/modules/logging"
	"proto-chat/modules/snowflake"
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

		// log.CloseChan <- 0
		// time.Sleep(1 * time.Second)
		os.Exit(0)
	}()

	config := readConfigFile()
	log.SetupLogging("TRACE", config.LogConsole, config.LogFile)

	// start := getTimestamp()
	// for i := 0; i < 10000; i++ {
	// 	log.Info("Starting server...")
	// }

	// measureTime(start, "test")

	if config.Sqlite {
		database.ConnectSqlite()
	} else {
		database.ConnectMariadb(config.DatabaseUsername, config.DatabasePassword, config.DatabaseAddress, strconv.Itoa(int(config.DatabasePort)), config.DatabaseName)
	}

	snowflake.SetSnowflakeServerID(0)

	// websocket
	// this will allow sending messages to multiple clients
	go broadCastChannel()

	var wsType string
	if config.TLS {
		wsType = "/wss"
	} else {
		wsType = "/ws"
	}
	http.HandleFunc(wsType, func(w http.ResponseWriter, r *http.Request) {
		wssHandler(w, r)
	})

	http.HandleFunc("GET /login-register.html", loginRegisterHandler)
	http.HandleFunc("GET /chat.html", chatHandler)

	http.HandleFunc("POST /login", postRequestHandler)
	http.HandleFunc("POST /register", postRequestHandler)

	http.HandleFunc("/", mainHandler)

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
