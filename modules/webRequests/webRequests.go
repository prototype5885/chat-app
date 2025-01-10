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
		if extension != "" && extension != ".html" {
			//userID := CheckIfTokenIsValid(w, r)
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
			http.ServeFile(w, r, getHtmlFilePath("/index.html"))
			return
		case "/chat.html":
			chatHandler(w, r)
			return
		case "/wss", "/ws":
			websocketHandler(w, r)
			return
		case "/login-register.html":
			loginRegisterHandler(w, r)
			return
		}

		// if accepting invite
		if strings.HasPrefix(r.URL.Path, "/invite") {
			inviteHandler(w, r)
			return
		}

	} else if r.Method == "POST" {
		switch r.URL.Path {
		case "/login":
			loginRequestHandler(w, r)
		case "/register":
			registerRequestHandler(w, r)
		// case "/check-profile-pic", "/check-server-pic":
		case "/upload-profile-pic", "/upload-server-pic":
			uploadAvatarHandler(w, r)
		case "/upload-attachment":
			uploadAttachmentHandler(w, r)
		case "/check-attachment":
			checkAttachmentHandler(w, r)
		}
	}
}
