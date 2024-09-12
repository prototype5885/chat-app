package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)
	// serve static files
	http.FileServer(http.Dir("./public")).ServeHTTP(w, r)
}

func wssHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)

	// check if the user trying to connect to websocket has token
	userID, result := checkIfTokenIsValid(r)
	if result.Success {
		acceptWsClient(userID, w, r)
		return
	}
	log.Println(result.Message)

	// someone is trying to connect to websocket directly without authorized token
	// this is not supposed to happen normally, as the .js file that connects to the websocket
	// is only sent if user was already authorized
	log.Println("Someone is trying to connect to websocket directly without token")
	log.Println("Redirecting to / ...")
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func loginRegisterHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)

	// check if user requesting login/registration already has a token
	_, result := checkIfTokenIsValid(r)
	if result.Success { // if user is trying to login but has a token
		log.Println("User is trying to access /login-register.html but already has authorized token")
		log.Println("Redirecting user to /chat.html ...")
		http.Redirect(w, r, "/chat.html", http.StatusMovedPermanently)
		return
	}

	// serve static files
	http.FileServer(http.Dir("./public")).ServeHTTP(w, r)
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)

	// check if user requesting login/registration already has a token
	_, result := checkIfTokenIsValid(r)
	if !result.Success { // if user tries to use the chat but has no token
		log.Println("Someone is trying to access /chat.html without authorized token")
		log.Println("Redirecting to / ...")
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	// serve static files
	http.FileServer(http.Dir("./public")).ServeHTTP(w, r)
}

func postRequestHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)

	// reading POST request body as bytes
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}

	// will close body on return
	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.Printf("Unable to close body: %s\n", err)
		}
	}()

	// print received json
	log.Println(string(bodyBytes))

	// handle different POST requests
	if r.URL.Path == "/login" || r.URL.Path == "/register" {
		cookie, result := loginOrRegister(bodyBytes, r.URL.Path)
		if result.Success {
			http.SetCookie(w, &cookie) // sets the token as cookie on the client side
		}

		// serialize the response into json
		responseJsonBytes, jsonErr := json.Marshal(result)
		if jsonErr != nil {
			log.Fatalln("Error serializing log/reg POST request response:", jsonErr.Error())
		}

		log.Println(string(responseJsonBytes))
		i, err := w.Write(responseJsonBytes)
		if err != nil {
			log.Printf("Error sending %s POST request response: %s", r.URL.Path, err)
		}
		log.Println(r.URL.Path, "POST request response was sent successfully:", i)
	}
}
