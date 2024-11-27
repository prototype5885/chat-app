package webRequests

import (
	"net/http"
	"path/filepath"
	jsfilesmerger "proto-chat/modules/jsFilesMerger"
	log "proto-chat/modules/logging"
	"strings"
)

const publicFolder string = "./public"

func MainHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)

	if r.Method == "GET" {
		var extension string = filepath.Ext(r.URL.Path)

		// check if client is requesting a file
		// continue if not
		if extension != "" {
			//userID := checkIfTokenIsValid(w, r)
			//if userID == 0 {
			//	respondText(w, "Not authorized")
			//	log.Hack("Someone is trying to request a file without token")
			//	return
			//}

			// look for js file changes then update script.js if there is before sending
			if extension == ".js" && jsfilesmerger.DynamicMergedJsGeneration {
				jsfilesmerger.CheckForChanges()
			}

			log.Debug("Serving file: %s", r.URL.Path)
			http.ServeFile(w, r, publicFolder+r.URL.Path)
			return
		}

		// if a normal http request
		switch r.URL.Path {
		case "/":
			http.ServeFile(w, r, getHtmlFilePath("/index"))
		case "/chat.html":
			chatHandler(w, r)
		case "/wss", "/ws":
			websocketHandler(w, r)
		case "/login-register.html":
			loginRegisterHandler(w, r)
		}

		// if accepting invite
		if strings.HasPrefix(r.URL.Path, "/invite") {
			inviteHandler(w, r)
		}

		// http.FileServer(http.Dir(publicFolder)).ServeHTTP(w, r) // serve static files
	} else if r.Method == "POST" {
		switch r.URL.Path {
		case "/login", "/register":
			loginRequestHandler(w, r)
		case "/upload-pfp":
			uploadProfilePicHandler(w, r)
			//case "/channel":
			//	log.Debug("Channel changed POST request")
		}
	}
}
