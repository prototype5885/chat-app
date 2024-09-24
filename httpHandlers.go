package main

import (
	"encoding/json"
	"io"
	"net/http"
	log "proto-chat/modules/logging"
)

func printReceivedRequest(url string, method string) {
	log.Debug("Received %s %s request", url, method)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)
	// serve static files
	http.FileServer(http.Dir("./public")).ServeHTTP(w, r)
}

func wssHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)
	log.Info("Someone is connecting to websocket...")

	// check if the user trying to connect to websocket has token
	userID := checkIfTokenIsValid(r)
	if userID != 0 {
		acceptWsClient(userID, w, r)
		return
	} else {
		// someone is trying to connect to websocket directly without authorized token
		// this is not supposed to happen normally, as the .js file that connects to the websocket
		// is only sent if user was already authorized
		log.Hack("Someone is trying to connect to websocket directly without token")
	}
}

func loginRegisterHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)

	// check if user requesting login/registration already has a token
	userID := checkIfTokenIsValid(r)
	if userID != 0 { // if user is trying to login but has a token
		log.Debug("User is trying to access /login-register.html but already has authorized token, redirecting to /chat.html...")
		http.Redirect(w, r, "/chat.html", http.StatusMovedPermanently)
		return
	}

	// serve static files
	http.FileServer(http.Dir("./public")).ServeHTTP(w, r)
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)

	// check if user requesting login/registration already has a token
	userID := checkIfTokenIsValid(r)
	if userID == 0 { // if user tries to use the chat but has no token
		log.Debug("Someone is trying to access /chat.html without authorized token, redirecting to / ...")
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
			log.Fatal("Unable to close body: %s", err)
		}
	}()

	// print received json
	log.Debug("Received json: %s", string(bodyBytes))

	// handle different POST requests
	if r.URL.Path == "/login" || r.URL.Path == "/register" {
		cookie, result := loginOrRegister(bodyBytes, r.URL.Path)
		if result.Success {
			http.SetCookie(w, &cookie) // sets the token as cookie on the client side
		}

		// serialize the response into json
		responseJsonBytes, jsonErr := json.Marshal(result)
		if jsonErr != nil {
			log.Error(jsonErr.Error())
			log.Fatal("Error serializing log/reg POST request response")
		}

		log.Debug("Response for log/reg request: %s", string(responseJsonBytes))
		i, err := w.Write(responseJsonBytes)
		if err != nil {
			log.Error(err.Error())
			log.Warn("Error sending %s POST request response", r.URL.Path)
		}
		log.Debug("%s POST request response was sent: %d", r.URL.Path, i)
	}
}
