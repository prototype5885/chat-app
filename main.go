package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"proto-chat/modules/snowflake"
	"strconv"
	"syscall"
)

func main() {
	// handle termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received termination signal...")
		log.Println("Closing db connection...")
		err := database.CloseDatabaseConnection()
		if err != nil {
			log.Println("Error closing db connection")
		}
		os.Exit(0)
	}()

	log.Println("Starting server...")
	// log.SetOutput(ioutil.Discard)

	config := readConfigFile()

	setupLogging(config.LogInFile)

	if config.Sqlite {
		database.ConnectSqlite()
	} else {
		database.ConnectMariadb(config.Username, config.Password, config.Address, strconv.Itoa(int(config.DatabasePort)), config.DatabaseName)
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

	log.Printf("Listening on port %d", config.Port)
	if config.TLS {
		const certFile = "./sslcert/cert.crt"
		const keyFile = "./sslcert/key.key"
		if err := http.ListenAndServeTLS(address, certFile, keyFile, nil); err != nil {
			log.Panic("Error starting TLS server:", err)
		}
	} else {
		if err := http.ListenAndServe(address, nil); err != nil {
			log.Panic("Error starting server:", err)
		}
	}
}

func ExitFunc() {
	log.Println("ende")
}
