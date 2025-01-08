package webRequests

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"proto-chat/modules/attachments"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"proto-chat/modules/pictures"
	"proto-chat/modules/snowflake"
	"proto-chat/modules/token"
	"proto-chat/modules/websocket"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// on /wss or /ws
func websocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Someone is connecting to websocket...")

	// check if the user trying to connect to websocket has token
	userID := token.CheckIfTokenIsValid(w, r)
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
	userID := token.CheckIfTokenIsValid(w, r)
	if userID != 0 { // if user is trying to log in but has a token
		log.Trace("User is trying to access /login-register.html but already has authorized token, redirecting to /chat.html...")
		redirect(w, r, "/chat.html")
		return
	}

	// serve static files
	http.ServeFile(w, r, getHtmlFilePath(r.URL.Path))
}

// on /chat GET request
func chatHandler(w http.ResponseWriter, r *http.Request) {

	// check if user requesting login/registration already has a token
	userID := token.CheckIfTokenIsValid(w, r)
	if userID == 0 { // if user tries to use the chat but has no token or expired
		log.Trace("Someone is trying to access /chat without authorized token, redirecting to / ...")
		redirect(w, r, "/")
	} else {
		// serve static files
		http.ServeFile(w, r, getHtmlFilePath(r.URL.Path))
	}
}

// on /login POST request
func loginRequestHandler(w http.ResponseWriter, r *http.Request) {
	const serverError = "Error processing /login POST request"

	// reading POST request body as bytes
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err.Error(), "Error reading /login POST request body")
		w.Write([]byte(serverError))
		return
	}

	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.FatalError(err.Error(), "Unable to close body of /login POST request")
		}
	}()

	type LoginRequest struct {
		Username string
		Password string
	}

	var loginRequest LoginRequest

	err = json.Unmarshal(bodyBytes, &loginRequest)
	if err != nil {
		log.WarnError(err.Error(), "Error deserializing /login body json")
		w.Write([]byte(serverError))
		return
	}

	if loginRequest.Username == "" || loginRequest.Password == "" {
		log.Hack("Someone sent a login POST request without username and/or password")
		w.Write([]byte(serverError))
		return
	}

	if !macros.IsAscii(loginRequest.Username) {
		log.Trace("Username [%s] wants to login with non ASCII characters in username", loginRequest.Username)
		w.Write([]byte("Non ASCII characters are not allowed"))
		return
	}

	tooLong := macros.CheckUsernameLength(loginRequest.Username)
	if tooLong {
		w.Write([]byte("Username is longer than max allowed"))
		return
	}

	const userError = "Wrong username or password"

	// get the password hash from the database using username
	passwordHash, userID := database.GetPasswordAndID(loginRequest.Username)
	if passwordHash == nil || userID == 0 {
		log.Warn("There is no user with username [%s]", loginRequest.Username)
		w.Write([]byte(userError))
		return
	}

	// decode password from base64 string to byte array so bcrypt can hash it, password is in SHA512 format
	// so the server can't really know what the original password was
	passwordBytes, err := base64.StdEncoding.DecodeString(loginRequest.Password)
	if err != nil {
		log.Error("Failed decoding SHA512 password string into byte array")
		w.Write([]byte(serverError))
		return
	}

	// compare given password with the retrieved hash
	log.Debug("Comparing password hash and string for user [%s]...", loginRequest.Username)
	var start = time.Now().UnixMilli()
	if err := bcrypt.CompareHashAndPassword(passwordHash, passwordBytes); err != nil {
		log.Warn("User entered wrong password for username [%s]", loginRequest.Username)
		w.Write([]byte(userError))
		return
	}

	log.Trace("%s: password matches with hash, comparison took: %d ms", loginRequest.Username, time.Now().UnixMilli()-start)

	cookie := token.NewTokenCookie(userID)
	http.SetCookie(w, &cookie)
	redirect(w, r, "/chat.html")
}

func registerRequestHandler(w http.ResponseWriter, r *http.Request) {
	const serverError = "Error processing /register POST request"

	// reading POST request body as bytes
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err.Error(), "Error reading /register POST request body")
		w.Write([]byte(serverError))
		return
	}

	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.FatalError(err.Error(), "Unable to close body of /register POST request")
		}
	}()

	// deserializing json
	type RegisterRequest struct {
		Username string
		Password string
	}

	var registerRequest RegisterRequest

	err = json.Unmarshal(bodyBytes, &registerRequest)
	if err != nil {
		log.WarnError(err.Error(), "Error deserializing /register body json")
		w.Write([]byte(serverError))
		return
	}

	// checking if deserialized values arent empty
	if registerRequest.Username == "" || registerRequest.Password == "" {
		log.Hack("Someone sent a /register POST request without username and/or password")
		w.Write([]byte(serverError))
		return
	}

	if !macros.IsAscii(registerRequest.Username) {
		log.Hack("Username [%s] wants to register their name with non ASCII character", registerRequest.Username)
		w.Write([]byte("Non ASCII characters are not allowed"))
		return
	}

	tooLong := macros.CheckUsernameLength(registerRequest.Username)
	if tooLong {
		w.Write([]byte("Username is longer than max allowed"))
		return
	}

	taken := database.CheckIfUsernameExists(registerRequest.Username)
	if taken {
		response := fmt.Sprintf("Username [%s] is already taken", registerRequest.Username)
		log.Trace("%s", response)
		w.Write([]byte(response))
		return
	}

	// decode password from base64 string to byte array so bcrypt can hash it, password is in SHA512 format
	// so the server can't really know what the original password was
	passwordBytes, err := base64.StdEncoding.DecodeString(registerRequest.Password)
	if err != nil {
		log.Error("Failed decoding SHA512 password string into byte array")
		w.Write([]byte(serverError))
		return
	}

	// check if received password is in proper format
	if len(passwordBytes) != 64 {
		log.Error("Password byte array length isn't 64 bytes")
		w.Write([]byte(serverError))
		return
	} else if len(registerRequest.Username) > 16 {
		log.Error("Username is longer than 16 bytes")
		w.Write([]byte(serverError))
		return
	}

	// hash the password using bcrypt
	var start int64 = time.Now().UnixMilli()
	passwordHash, err := bcrypt.GenerateFromPassword(passwordBytes, 10)
	if err != nil {
		log.FatalError(err.Error(), "Failed generating bcrypt password hash for username [%s]", registerRequest.Username)
	}
	macros.MeasureTime(start, "Password hashing for user "+registerRequest.Username)

	var userID uint64 = snowflake.Generate()

	success := database.RegisterUser(userID, registerRequest.Username, passwordHash)
	if !success {
		w.Write([]byte(serverError))
		return
	}

	cookie := token.NewTokenCookie(userID)
	http.SetCookie(w, &cookie)
	redirect(w, r, "/chat.html")
}
func inviteHandler(w http.ResponseWriter, r *http.Request) {
	log.Trace("Received invite request")

	userID := token.CheckIfTokenIsValid(w, r)
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
				err := database.Insert(database.ServerMemberShort{ServerID: serverID, UserID: userID})
				if err != nil {
					respondText(w, "Failed joining server")
				}
				respondText(w, "Successfully joined server ID [%d]", serverID)
				log.Trace("User ID [%d] successfully joined server ID [%d]", userID, serverID)
				redirect(w, r, "/chat.html")
			} else {
				respondText(w, "No invite exists with this invite ID")
				return
			}
		}
	}
}

func uploadProfilePicHandler(w http.ResponseWriter, r *http.Request) {
	userID := token.CheckIfTokenIsValid(w, r)
	if userID == 0 {
		respondText(w, "Who are you?")
		log.Hack("Someone is trying to upload a profile picture without token")
		return
	}

	log.Trace("User ID [%d] wants to change their profile pic", userID)

	err := r.ParseMultipartForm(100 << 10)
	if err != nil {
		log.WarnError(err.Error(), "Received profile picture from user ID [%d] is too big in size", userID)
		http.Error(w, "Uploaded picture is too big", http.StatusBadRequest)
		return
	}

	// parse formfile
	formFile, _, err := r.FormFile("pfp")
	if err != nil {
		log.WarnError(err.Error(), "Error parsing multipart form 2")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	defer formFile.Close()

	// read bytes from received profile pic
	imgBytes, err := io.ReadAll(formFile)
	if err != nil {
		log.WarnError(err.Error(), "Error reading formFile of profile pic from user ID [%d]", userID)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	// check if received profile pic is in correct format
	var issue string
	imgBytes, issue = pictures.CheckProfilePic(imgBytes, userID)
	if issue != "" {
		http.Error(w, issue, http.StatusBadRequest)
		return
	}

	hash := sha256.Sum256(imgBytes)
	fileName := hex.EncodeToString(hash[:]) + ".jpg"
	var pfpPath = "./public/content/avatars/" + fileName

	// check if profile pic file exists already, otherwise save as new
	_, err = os.Stat(pfpPath)
	if os.IsNotExist(err) {
		log.Trace("Profile pic [%s] doesn't exist yet, creating...", fileName)
		err = os.WriteFile(pfpPath, imgBytes, 0644)
		if err != nil {
			log.FatalError(err.Error(), "Error writing bytes to profile pic file from user ID [%d]", userID)
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		log.FatalError(err.Error(), "Error creating file for profile pic from user ID [%d]", userID)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	} else {
		log.Trace("Profile pic [%s] of same hash already exists, using that one...", fileName)
	}

	success := database.UpdateUserValue(userID, fileName, "picture")
	if !success {
		log.Warn("Failed updating profile picture of user ID [%d] in database", userID)
		return
	}

	websocket.OnProfilePicChanged(userID, fileName)
}

func uploadAttachmentHandler(w http.ResponseWriter, r *http.Request) {
	userID := token.CheckIfTokenIsValid(w, r)
	if userID == 0 {
		_, err := fmt.Fprintf(w, "Who are you?")
		if err != nil {
			log.Error("%s", err.Error())
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

		const attachmentsPath string = "./public/content/attachments/"

		if part.FileName() != "" {
			extension := filepath.Ext(part.FileName())

			var fileBytes *bytes.Buffer = new(bytes.Buffer)
			var fileName string

			_, err := io.Copy(fileBytes, part)
			if err != nil {
				log.FatalError(err.Error(), "Error reading video sent by user ID [%d]", userID)
				continue
			}

			var hash [32]byte = sha256.Sum256(fileBytes.Bytes())

			log.Trace("Extension sent by user ID [%d] is [%s]", userID, extension)
			switch extension {

			case ".jpg", ".png", ".webp", ".jpeg", ".bmp":
				// var img image.Image
				// if extension == ".webp" {
				// 	img, err = webp.Decode(fileBytes)
				// } else if extension == ".bmp" {
				// 	img, err = bmp.Decode(fileBytes)
				// } else if extension == ".jpg" || extension == ".jpeg" {
				// 	img, err = jpeg.Decode(fileBytes)
				// } else if extension == ".png" {
				// 	img, err = png.Decode(fileBytes)
				// }
				// if err != nil {
				// 	log.Error("%s", err.Error())
				// 	log.Hack("Received attachment picture of extension [%s] from user ID [%d] could not be decoded", extension, userID)
				// 	continue
				// }

				// resizedImg := imaging.Resize(img, 2048, 0, imaging.Lanczos)

				// opt := jpeg.Options{Quality: 75}
				// err = jpeg.Encode(fileBytes, resizedImg, &opt)
				// if err != nil {
				// 	http.Error(w, "error", http.StatusInternalServerError)
				// 	log.FatalError(err.Error(), "Error encoding image sent by user ID [%d]", userID)
				// 	continue
				// }
				// fileName = hex.EncodeToString(hash[:]) + ".jpg"
				fileName = hex.EncodeToString(hash[:]) + extension
			case ".mp4", ".webm", ".mov":
				fileName = hex.EncodeToString(hash[:]) + extension
			default:
				log.Hack("User ID [%d] sent attachment of unknown extension [%s]", userID, extension)
				continue
			}

			fileNames = append(fileNames, fileName)

			path := attachmentsPath + fileName

			_, err = os.Stat(path)
			if err == nil {
				log.Trace("Attachment at path [%s] already exists", path)
			} else if os.IsNotExist(err) {
				log.Trace("Attachment at path [%s] doesn't exist, creating...", path)
				err = os.WriteFile(path, fileBytes.Bytes(), 0644)
				if err != nil {
					log.FatalError(err.Error(), "Error writing to file [%s]", path)
					return
				}
			} else {
				log.FatalError(err.Error(), "Error checking file [%s]", path)
			}

		}
	}

	log.Trace("[%s] POST request response was sent", r.URL.Path)

	attachmentToken := *attachments.OnAttachmentUploaded(userID, fileNames)
	encoded := base64.StdEncoding.EncodeToString(attachmentToken[:])

	log.Trace("Response for [%s] POST request: [%s]", r.URL.Path, encoded)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte(fmt.Sprintf(`{"AttToken":"%s"}`, encoded)))
	if err != nil {
		log.WarnError(err.Error(), "Error sending %s POST request response", r.URL.Path)
	}
}
