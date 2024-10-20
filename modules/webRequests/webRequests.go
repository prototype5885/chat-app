package webRequests

import (
	"net/http"
	"path/filepath"
	jsfilesmerger "proto-chat/modules/jsFilesMerger"
	log "proto-chat/modules/logging"
)

const publicFolder string = "./public"
const picsFolder string = publicFolder + "/pics"

const jsFolder string = publicFolder + "/js"
const uiFolder string = publicFolder + "/ui"

// on any requests
func MainHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r.URL.Path, r.Method)

	if r.Method == "GET" {
		var extension string = filepath.Ext(r.URL.Path)
		// check if client is requesting a file
		// continue if not
		if extension != "" {
			switch extension {
			case ".webp", ".jpg", ".png":
				http.ServeFile(w, r, picsFolder+r.URL.Path)
			case ".svg":
				http.ServeFile(w, r, uiFolder+r.URL.Path)
			case ".css":
				http.ServeFile(w, r, publicFolder+r.URL.Path)
			case ".js":
				// w.Header().Set("Content-Type", "application/javascript")
				// _, err := io.WriteString(w, jsfilesmerger.MergeJsFiles())
				// if err != nil {
				// 	log.WarnError(err.Error(), "Error sending javascript string to client")
				// }
				jsfilesmerger.CheckForChanges()
				http.ServeFile(w, r, jsFolder+r.URL.Path)
			}
			// don't continue if request was confirmed to be a file
			return
		}

		// if a normal http request
		switch r.URL.Path {
		case "/":
			http.ServeFile(w, r, getHtmlFilePath("/index"))
		case "/chat":
			chatHandler(w, r)
		case "/wss", "/ws":
			websocketHandler(w, r)
		case "/invite":
			inviteHandler(w, r)
		case "/login-register":
			loginRegisterHandler(w, r)
		}

		// http.FileServer(http.Dir(publicFolder)).ServeHTTP(w, r) // serve static files
	} else if r.Method == "POST" {
		switch r.URL.Path {
		case "/login", "/register":
			loginRequestHandler(w, r)
		case "/upload":
			log.Debug("Uploading")
		}
	}
}
