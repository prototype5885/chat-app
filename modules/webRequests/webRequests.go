package webRequests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/websocket"
	"strconv"
	"strings"
)

func printReceivedRequest(url string, method string) {
	log.Trace("Received %s %s request", url, method)
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)
	http.FileServer(http.Dir("./public")).ServeHTTP(w, r) // serve static files
}

func WssHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)
	log.Info("Someone is connecting to websocket...")

	// check if the user trying to connect to websocket has token
	userID := checkIfTokenIsValid(w, r)
	if userID != 0 {
		websocket.AcceptWsClient(userID, w, r)
		return
	} else {
		// someone is trying to connect to websocket directly without authorized token
		// this is not supposed to happen normally, as the .js file that connects to the websocket
		// is only sent if user was already authorized
		log.Hack("Someone is trying to connect to websocket directly without token")
	}
}

func LoginRegisterHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)

	// check if user requesting login/registration already has a token
	userID := checkIfTokenIsValid(w, r)
	if userID != 0 { // if user is trying to login but has a token
		log.Debug("User is trying to access /login-register.html but already has authorized token, redirecting to /chat.html...")
		http.Redirect(w, r, "/chat.html", http.StatusMovedPermanently)
		return
	}

	// serve static files
	http.FileServer(http.Dir("./public")).ServeHTTP(w, r)
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)

	// check if user requesting login/registration already has a token
	userID := checkIfTokenIsValid(w, r)
	if userID == 0 { // if user tries to use the chat but has no token
		log.Debug("Someone is trying to access /chat.html without authorized token, redirecting to / ...")
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	// serve static files
	http.FileServer(http.Dir("./public")).ServeHTTP(w, r)
}

func PostRequestHandler(w http.ResponseWriter, r *http.Request) {
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
	log.Trace("Received json: %s", string(bodyBytes))

	// handle different POST requests
	if r.URL.Path == "/login" || r.URL.Path == "/register" {
		cookie, result := loginOrRegister(bodyBytes, r.URL.Path)
		if result.Success {
			http.SetCookie(w, &cookie) // sets the token as cookie on the client side
		}

		// serialize the response into json
		responseJsonBytes, jsonErr := json.Marshal(result)
		if jsonErr != nil {
			log.FatalError(jsonErr.Error(), "Error serializing log/reg POST request response")
		}

		log.Debug("Response for log/reg request: %s", string(responseJsonBytes))
		i, err := w.Write(responseJsonBytes)
		if err != nil {
			log.WarnError(err.Error(), "Error sending %s POST request response", r.URL.Path)
		}
		log.Debug("%s POST request response was sent: %d", r.URL.Path, i)
	}
}

func InviteHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)

	var userID uint64 = checkIfTokenIsValid(w, r)
	if userID == 0 { // if user has no valid token
		fmt.Fprintln(w, "Not logged in")
		log.Debug("Someone without authorized token clicked on an invite link")
		return
	} else {
		fmt.Fprintln(w, "Logged in")

		parts := strings.Split(r.URL.Path, "/invite/")
		if len(parts) > 1 {
			var inviteIDstring string = parts[len(parts)-1]
			inviteID, err := strconv.ParseUint(inviteIDstring, 10, 64)
			if err != nil {
				log.Warn("User ID [%d] sent a server invite http request where the ID can't be parsed [%s]", userID, inviteIDstring)
				return
			}
			log.Debug("Server invite ID is: [%d]", inviteID)
			var serverID uint64 = database.ServerInvitesTable.ConfirmServerInviteID(inviteID)
			if serverID != 0 {
				log.Debug("Invite ID [%d] belongs to server ID [%d]", inviteID, serverID)

				if !database.Insert(database.ServerMember{ServerID: serverID, UserID: userID}) {

				}
			}
		}
	}
}
