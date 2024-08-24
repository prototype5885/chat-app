package main

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
)

var logInFile bool = false

type LoginData struct {
	Logreg   string
	Username string
	Password string
}

func SetUpLogging() {
	if logInFile {
		os.MkdirAll("logs", fs.FileMode(os.ModePerm))
		file, err := os.OpenFile("./logs/protochat.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(file)
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	var user LoginData
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		// return HTTP 400 bad request
	}
	log.Println(user.Logreg)
	log.Println(user.Username)
	log.Println(user.Password)
}

func main() {
	SetUpLogging()
	log.Println("Starting server...")

	// todo initialize database here

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	http.HandleFunc("/logging_in", handlePostRequest)

	log.Println("Listening on port 3000...")
	err := http.ListenAndServeTLS(":3000", "./sslcert/selfsigned.crt", "./sslcert/selfsigned.key", nil)
	if err != nil {
		log.Fatal(err)
	}
}
