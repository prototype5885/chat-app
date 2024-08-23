package main

import (
	"log"
	"net/http"
)

func SetUpLogging() {
	// file, err := os.OpenFile("protochat.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.SetOutput(file)

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func main() {
	SetUpLogging()
	log.Println("Starting server...")

	// todo initialize database here

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	log.Println("Listening on port 3000...")
	err := http.ListenAndServeTLS(":3000", "./sslcert/selfsigned.crt", "./sslcert/selfsigned.key", nil)
	if err != nil {
		log.Fatal(err)
	}
}
