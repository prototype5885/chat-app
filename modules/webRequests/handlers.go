package webRequests

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"proto-chat/modules/attachments"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/websocket"
	"strconv"
	"strings"
)

// on /wss or /ws
func websocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Someone is connecting to websocket...")

	// check if the user trying to connect to websocket has token
	userID := CheckIfTokenIsValid(w, r)
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

// on /login-register GET request
func loginRegisterHandler(w http.ResponseWriter, r *http.Request) {
	// check if user requesting login/registration already has a token
	userID := CheckIfTokenIsValid(w, r)
	if userID != 0 { // if user is trying to log in but has a token
		log.Trace("User is trying to access /login-register but already has authorized token, redirecting to /chat.html...")
		redirect(w, r, "/chat.html")
		return
	}

	// serve static files
	http.ServeFile(w, r, getHtmlFilePath(r.URL.Path))
}

// on /chat GET request
func chatHandler(w http.ResponseWriter, r *http.Request) {

	// check if user requesting login/registration already has a token
	userID := CheckIfTokenIsValid(w, r)
	if userID == 0 { // if user tries to use the chat but has no token
		log.Trace("Someone is trying to access /chat without authorized token, redirecting to / ...")
		redirect(w, r, "/")
	} else {
		// serve static files
		http.ServeFile(w, r, getHtmlFilePath(r.URL.Path))
	}
}

// on /login or /register POST requests
func loginRequestHandler(w http.ResponseWriter, r *http.Request) {

	// reading POST request body as bytes
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}

	// print received json
	log.Trace("Received json from post request: %s", string(bodyBytes))

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

		log.Trace("Response for log/reg request: %s", string(responseJsonBytes))
		i, err := w.Write(responseJsonBytes)
		if err != nil {
			log.WarnError(err.Error(), "Error sending %s POST request response", r.URL.Path)
		}
		log.Trace("%s POST request response was sent: %d", r.URL.Path, i)
	}
}

func inviteHandler(w http.ResponseWriter, r *http.Request) {
	log.Trace("Received invite request")

	userID := CheckIfTokenIsValid(w, r)
	if userID == 0 { // if user has no valid token
		respondText(w, "Not logged in")
		log.Hack("Someone without authorized token clicked on an invite link")
		return
	} else {
		parts := strings.Split(r.URL.Path, "/invite/")
		if len(parts) > 1 {
			inviteIDstr := parts[len(parts)-1]
			inviteID, err := strconv.ParseUint(inviteIDstr, 10, 64)
			if err != nil {
				respondText(w, "What kind of invite ID is that?")
				log.Hack("User ID [%d] sent a server invite http request where the ID can't be parsed [%s]", userID, inviteIDstr)
				return
			}
			serverID := database.ConfirmServerInviteID(inviteID)
			if serverID != 0 {
				success := database.Insert(database.ServerMember{ServerID: serverID, UserID: userID})
				if success {
					respondText(w, "Successfully joined server ID [%d]", serverID)
					log.Trace("User ID [%d] successfully joined server ID [%d]", userID, serverID)
					redirect(w, r, "/chat.html")
					return
				} else {
					respondText(w, "Failed joining server")
				}
			} else {
				respondText(w, "No invite exists with this invite ID")
				return
			}
		}
	}
}

func uploadProfilePicHandler(w http.ResponseWriter, r *http.Request) {
	userID := CheckIfTokenIsValid(w, r)
	if userID == 0 {
		respondText(w, "Who are you?")
		log.Hack("Someone is trying to upload a profile picture without token")
		return
	}

	log.Trace("User ID [%d] wants to change their profile pic", userID)

	err := r.ParseMultipartForm(100 << 10)
	if err != nil {
		log.WarnError(err.Error(), "Received profile picture from user ID [%d] is too big in size", userID)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	formFile, handler, err := r.FormFile("pfp")
	if err != nil {
		log.WarnError(err.Error(), "Error parsing multipart form 2")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer formFile.Close()

	var pfpPath = "./public/content/avatars/" + handler.Filename

	pfp, err := os.Create(pfpPath)
	if err != nil {
		log.WarnError(err.Error(), "Error creating formFile of profile pic from user ID [%d]", userID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer pfp.Close()

	if _, err := io.Copy(pfp, formFile); err != nil {
		log.WarnError(err.Error(), "Error copying profile pic to avatars folder from user ID [%d]", userID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}

	// success := database.UpdateUserRow(database.User{Picture: handler.Filename}, userID)
	// if !success {
	// 	log.Warn("Failed updating profile picture of user ID [%d]", userID)
	// 	return
	// }

	websocket.OnProfilePicChanged(userID, handler.Filename)
}

func uploadAttachmentHandler(w http.ResponseWriter, r *http.Request) {
	userID := CheckIfTokenIsValid(w, r)
	if userID == 0 {
		_, err := fmt.Fprintf(w, "Who are you?")
		if err != nil {
			log.Error(err.Error())
		}
		log.Hack("Someone is trying to upload an attachment without token")
		return
	}

	log.Trace("User ID [%d] is uploading an attachment", userID)

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var fileNames []string

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		}

		if part.FileName() != "" {
			img, _, err := image.Decode(part)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			buf := new(bytes.Buffer)
			opt := jpeg.Options{Quality: 75}
			err = jpeg.Encode(buf, img, &opt)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.FatalError(err.Error(), "Error encoding image sent by user ID [%d]", userID)
			}

			hash := sha256.Sum256(buf.Bytes())

			fileName := hex.EncodeToString(hash[:]) + ".jpg"
			fileNames = append(fileNames, fileName)

			var filePath = "./public/content/attachments/" + fileName

			_, err = os.Stat(filePath)
			if err == nil {
				log.Trace("Attachment at path [%s] already exists", filePath)
			} else if os.IsNotExist(err) {
				log.Trace("Attachment at path [%s] doesn't exist, creating...", filePath)
				err = os.WriteFile(filePath, buf.Bytes(), 0644)
				if err != nil {
					log.FatalError(err.Error(), "Error writing to file:")
					return
				}
			} else {
				fmt.Println("Error checking file:", err)
			}
		}
	}

	log.Trace("[%s] POST request response was sent", r.URL.Path)

	attachmentToken := attachments.OnAttachmentUploaded(userID, fileNames)
	encoded := base64.StdEncoding.EncodeToString(attachmentToken)

	log.Trace("Response for [%s] POST request: [%s]", r.URL.Path, encoded)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte(fmt.Sprintf(`{"AttToken":"%s"}`, encoded)))
	if err != nil {
		log.WarnError(err.Error(), "Error sending %s POST request response", r.URL.Path)
	}
}
