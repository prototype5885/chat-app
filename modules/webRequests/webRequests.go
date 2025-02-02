package webRequests

import (
	log "chat-app/modules/logging"
	"net/http"
	"path/filepath"
	"strings"
)

const publicFolder string = "./public"

//func CheckCacheServer() bool {
//	tr := &http.Transport{
//		TLSClientConfig: &tls.Config{
//			InsecureSkipVerify: true, // Disable SSL certificate verification
//		},
//	}
//
//	client := &http.Client{
//		Transport: tr,
//	}
//
//	// Make a quick HEAD request to the cache server to check if it's up
//	resp, err := client.Head(websocket.ImageHost)
//	if err != nil {
//		log.WarnError(err.Error(), "Error checking if cache server is available")
//		return false
//	}
//	defer func(Body io.ReadCloser) {
//		err := Body.Close()
//		if err != nil {
//			log.WarnError(err.Error(), "Error closing body of request checking if cache server is available")
//		}
//	}(resp.Body)
//
//	return resp.StatusCode == http.StatusOK
//}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	printReceivedRequest(r)

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
			//if extension == ".js" && jsfilesmerger.DynamicMergedJsGeneration {
			//	jsfilesmerger.CheckForChanges()
			//}

			//switch extension {
			//case ".js", ".css", ".json":
			//	log.Trace("Serving file: [%s]", r.URL.Path)
			//	http.ServeFile(w, r, publicFolder+r.URL.Path)
			//default:
			//	if r.Header.Get("Cache-Server") == "1" {
			//		log.Trace("Serving file to cache server [%s]", r.URL.Path)
			//		http.ServeFile(w, r, publicFolder+r.URL.Path)
			//	} else {
			//		available := CheckCacheServer()
			//		if available {
			//			log.Trace("Cache server is available, to [%s] to serve file [%s]", websocket.ImageHost, r.URL.Path)
			//			http.Redirect(w, r, websocket.ImageHost+r.URL.Path, http.StatusFound)
			//		} else {
			//			log.Warn("Cache server is not available, serving file: [%s]", r.URL.Path)
			//			http.ServeFile(w, r, publicFolder+r.URL.Path)
			//		}
			//	}
			//}
			log.Trace("Serving file: [%s]", r.URL.Path)
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
		case "/upload-profile-pic", "/upload-server-pic":
			uploadAvatarHandler(w, r)
		case "/upload-banner-pic":
			uploadBannerHandler(w, r)
		case "/upload-attachment":
			uploadAttachmentHandler(w, r)
		case "/check-attachment":
			checkAttachmentHandler(w, r)
		}
	}
}
