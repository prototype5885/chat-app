package main

import (
	"log"
	"net/http"
	"strconv"
)

func main() {
	log.Println("Starting server...")

	config := readConfigFile()

	setupLogging(config.LogInFile)

	// database
	if config.Sqlite {
		database.ConnectSqlite()
	} else {
		database.ConnectMariadb(config.Username, config.Password, config.Address, strconv.Itoa(int(config.Port)), config.DatabaseName)
	}

	// websocket
	go pingClients()
	http.HandleFunc("/wss", func(w http.ResponseWriter, r *http.Request) {
		wssHandler(w, r)
	})

	// http.HandleFunc("GET /wss", wssHandler)
	http.HandleFunc("GET /login-register.html", loginRegisterHandler)
	http.HandleFunc("GET /chat.html", chatHandler)

	http.HandleFunc("POST /login", postRequestHandler)
	http.HandleFunc("POST /register", postRequestHandler)

	http.HandleFunc("/", mainHandler)

	const certFile = "./sslcert/selfsigned.crt"
	const keyFile = "./sslcert/selfsigned.key"

	log.Println("Listening on port 3000")
	if err := http.ListenAndServeTLS(":3000", certFile, keyFile, nil); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
