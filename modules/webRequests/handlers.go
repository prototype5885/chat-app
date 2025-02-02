package webRequests

import (
	"bytes"
	"chat-app/modules/attachments"
	"chat-app/modules/database"
	log "chat-app/modules/logging"
	"chat-app/modules/macros"
	"chat-app/modules/pictures"
	"chat-app/modules/snowflake"
	"chat-app/modules/token"
	"chat-app/modules/websocket"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
		redirect(w, r, "/")
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
		log.Trace("Someone is trying to access /chat.html without authorized token, redirecting to / ...")
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
	const errorSendingError = "Error while replying with error in registerRequestHandler"

	// reading POST request body as bytes
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err.Error(), "Error reading /register POST request body")
		_, err := w.Write([]byte("Error reading /register POST request body"))
		if err != nil {
			log.Error(errorSendingError)
		}
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
		Username  string
		Password  string
		InviteKey string
	}

	var req RegisterRequest

	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		log.WarnError(err.Error(), "Error deserializing /register body json")
		_, err := w.Write([]byte("Error deserializing the received /register body json"))
		if err != nil {
			log.Error(errorSendingError)
		}
		return
	}

	// check if invite key is valid
	if req.InviteKey == "" {
		log.Hack("Username [%s] wants to register without invite key", req.Username)
		_, err := w.Write([]byte("No invite key was provided"))
		if err != nil {
			log.Error(errorSendingError)
		}
		return
	}

	exists := database.ConfirmInviteKey(req.InviteKey)
	if exists {
		log.Trace("Invite key [%s] is being used", req.InviteKey)
	} else {
		response := fmt.Sprintf("Invite key [%s] is invalid", req.InviteKey)
		log.Hack("%s", response)
		_, err := w.Write([]byte(response))
		if err != nil {
			log.Error(errorSendingError)
		}
		return
	}

	// checking if deserialized values aren't empty
	if req.Username == "" || req.Password == "" {
		log.Hack("Someone sent a /register POST request without username and/or password")
		_, err := w.Write([]byte("Username or password is missing"))
		if err != nil {
			log.Error(errorSendingError)
		}
		return
	}

	if !macros.IsAscii(req.Username) {
		log.Hack("Username [%s] wants to register their name with non ASCII character", req.Username)
		_, err := w.Write([]byte("Non ASCII characters are not allowed"))
		if err != nil {
			log.Error(errorSendingError)
		}
		return
	}

	tooLong := macros.CheckUsernameLength(req.Username)
	if tooLong {
		response := fmt.Sprintf("Username [%s] is longer than max allowed", req.Username)
		log.Hack("%s", response)
		_, err := w.Write([]byte(response))
		if err != nil {
			log.Error(errorSendingError)
		}
		return
	}

	taken := database.CheckIfUsernameExists(req.Username)
	if taken {
		response := fmt.Sprintf("Username [%s] is already taken", req.Username)
		log.Trace("%s", response)
		_, err := w.Write([]byte(response))
		if err != nil {
			log.Error(errorSendingError)
		}
		return
	}

	// decode password from base64 string to byte array so bcrypt can hash it, password is in SHA512 format
	// so the server can't really know what the original password was
	passwordBytes, err := base64.StdEncoding.DecodeString(req.Password)
	if err != nil {
		response := "Failed decoding SHA512 password string into byte array"
		log.Error("%s", response)
		_, err := w.Write([]byte(response))
		if err != nil {
			log.Error(errorSendingError)
		}
		return
	}

	// check if received password is in proper format
	if len(passwordBytes) != 64 {
		response := "Password byte array length isn't 64 bytes"
		log.Error("%s", response)
		_, err := w.Write([]byte(response))
		if err != nil {
			log.Error(errorSendingError)
		}
		return
	} else if len(req.Username) > 16 {
		response := "Username is longer than 16 bytes"
		log.Error("%s", response)
		_, err := w.Write([]byte(response))
		if err != nil {
			log.Error(errorSendingError)
		}
		return
	}

	// hash the password using bcrypt
	var start int64 = time.Now().UnixMilli()
	passwordHash, err := bcrypt.GenerateFromPassword(passwordBytes, 10)
	if err != nil {
		log.FatalError(err.Error(), "Failed generating bcrypt password hash for username [%s]", req.Username)
	}
	macros.MeasureTime(start, "Password hashing for user "+req.Username)

	var userID uint64 = snowflake.Generate()

	success := database.RegisterUser(userID, req.Username, passwordHash)
	if !success {
		response := "Registration failed"
		log.Error("%s", response)
		_, err := w.Write([]byte(response))
		if err != nil {
			log.Error(errorSendingError)
		}
		return
	}

	cookie := token.NewTokenCookie(userID)
	http.SetCookie(w, &cookie)
	redirect(w, r, "/chat.html")

	log.Trace("Registered user [%s] using invite key [%s] successfully", req.Username, req.InviteKey)
	database.Delete(database.InviteKey{
		Key: req.InviteKey,
	})
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
			// parse the id part from link
			stringInviteID := parts[len(parts)-1]
			inviteID, err := strconv.ParseUint(stringInviteID, 10, 64)

			// if invite link id is unparsable
			if err != nil {
				respondText(w, "What kind of invite ID is that?")
				log.Hack("User ID [%d] sent a server invite http request where the ID can't be parsed [%s]", userID, stringInviteID)
				return
			}

			// check if ID exists in database
			serverID, singleUse, _ := database.ConfirmServerInviteID(inviteID)
			if serverID != 0 {
				// add user into the server
				err := database.Insert(database.ServerMemberShort{ServerID: serverID, UserID: userID})
				if err != nil {
					respondText(w, "Failed joining server")
					return
				}
				log.Trace("User ID [%d] successfully joined server ID [%d]", userID, serverID)
				redirect(w, r, "/chat.html")
				websocket.OnUserJoinedServer(userID, serverID)

				// also delete from database if one time use link was used successfully
				if singleUse {
					log.Trace("Invite ID [%d] was one time use only, deleting from database...", inviteID)
					database.Delete(database.ServerInviteDelete{InviteID: inviteID})
				}
			} else {
				respondText(w, "No invite exists with given invite ID")
				return
			}
		}
	}
}

func uploadBannerHandler(w http.ResponseWriter, r *http.Request) {
	userID := token.CheckIfTokenIsValid(w, r)
	if userID == 0 {
		log.Hack("Someone is trying to upload a server banner without token")
		http.Error(w, "Who are you?", http.StatusUnauthorized)
		return
	}

	log.Trace("User ID [%d] wants to change a server banner", userID)

	// limit size
	const maxSizeMb float32 = 1
	const maxSize int64 = int64(1024 * 1024 * maxSizeMb)
	if r.ContentLength > maxSize {
		log.Warn("User ID [%d] tries to upload a server banner larger than [%f] MB", userID, maxSizeMb)
		http.Error(w, fmt.Sprintf("Uploaded picture is larger than allowed %f MB", maxSizeMb), http.StatusBadRequest)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	// parse formfile
	picFormFile, _, err := r.FormFile("banner-pic")
	if err != nil {
		log.WarnError(err.Error(), "Error parsing picture formfile sent by user ID [%d]", userID)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	defer picFormFile.Close()

	// read bytes from received avatar pic
	imgBytes, err := io.ReadAll(picFormFile)
	if err != nil {
		log.WarnError(err.Error(), "Error reading picture formfile of server banner from user ID [%d]", userID)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	var extension string

	mimeType := http.DetectContentType(imgBytes)
	switch mimeType {
	case "image/jpeg", "image/jpg", "image/png":
		extension = ".jpg"
	case "image/gif":
		extension = ".gif"
	default:
		log.Hack("User ID [%d] tried to upload unsupported filetype as server banner", userID)
		http.Error(w, "Unsupported filetype", http.StatusBadRequest)
		return
	}

	// check if received avatar pic is in correct format and compress
	if extension == ".jpg" {
		issue := pictures.CheckBanner(&imgBytes, userID)
		if issue != "" {
			http.Error(w, issue, http.StatusBadRequest)
			return
		}
	}

	hash := sha256.Sum256(imgBytes)
	fileName := hex.EncodeToString(hash[:]) + extension
	var filePath = "./public/content/banners/" + fileName

	// check if avatar pic file exists already, otherwise save as new
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		log.Trace("Server banner doesn't exist yet, creating...", fileName)
		err = os.WriteFile(filePath, imgBytes, 0644)
		if err != nil {
			log.FatalError(err.Error(), "Error writing bytes to server banner file from user ID [%d]", userID)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		log.FatalError(err.Error(), "Error creating file for server banner from user ID [%d]", userID)
		http.Error(w, "", http.StatusInternalServerError)
		return
	} else {
		log.Trace("Server banner [%s] of same hash already exists, using that one...", fileName)
	}

	serverID, err := strconv.ParseUint(r.FormValue("serverID"), 10, 64)
	if err != nil {
		log.WarnError(err.Error(), "Error parsing serverID as uint64 while changing banner of server of user ID [%d]", userID)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	log.Trace("User ID [%d] wants to change banner of server ID [%d]", userID, serverID)
	success := database.ChangeServerBanner(userID, serverID, fileName)
	if !success {
		log.Hack("Failed updating picture of server ID [%d] requested by user ID [%d]", serverID, userID)
		http.Error(w, "Failed updating picture of server", http.StatusForbidden)
		return
	}

	websocket.OnServerBannerChanged(serverID, fileName)
}

func uploadAvatarHandler(w http.ResponseWriter, r *http.Request) {
	var picType string
	var folder string
	switch r.URL.Path {
	case "/upload-profile-pic":
		picType = "profile-pic"
		folder = "avatars"
	case "/upload-server-pic":
		picType = "server-pic"
		folder = "avatars"
	}

	userID := token.CheckIfTokenIsValid(w, r)
	if userID == 0 {
		log.Hack("Someone is trying to upload a [%s] without token", picType)
		http.Error(w, "Who are you?", http.StatusUnauthorized)
		return
	}

	log.Trace("User ID [%d] wants to change [%s]", userID, picType)

	// limit size
	const maxSizeMb float32 = 1
	const maxSize int64 = int64(1024 * 1024 * maxSizeMb)
	if r.ContentLength > maxSize {
		log.Warn("User ID [%d] tries to upload [%s] larger than [%f] MB", userID, picType, maxSizeMb)
		http.Error(w, fmt.Sprintf("Uploaded picture is larger than allowed %f MB", maxSizeMb), http.StatusBadRequest)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	// parse formfile
	picFormFile, _, err := r.FormFile(picType)
	if err != nil {
		log.WarnError(err.Error(), "Error parsing picture formfile sent by user ID [%d]", userID)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	defer picFormFile.Close()

	// read bytes from received avatar pic
	imgBytes, err := io.ReadAll(picFormFile)
	if err != nil {
		log.WarnError(err.Error(), "Error reading picture formfile of [%s] from user ID [%d]", picType, userID)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	var extension string

	mimeType := http.DetectContentType(imgBytes)
	switch mimeType {
	case "image/jpeg", "image/jpg", "image/png":
		extension = ".jpg"
	case "image/gif":
		extension = ".gif"
	default:
		log.Hack("User ID [%d] tried to upload unsupported filetype as [%s]", userID, picType)
		http.Error(w, "Unsupported filetype", http.StatusBadRequest)
		return
	}

	// check if received avatar pic is in correct format and compress
	if extension == ".jpg" {
		issue := pictures.CheckAvatarPic(&imgBytes, userID)
		if issue != "" {
			http.Error(w, issue, http.StatusBadRequest)
			return
		}
	} else if extension == ".gif" {
		// gifData, _ := gif.DecodeAll(bytes.NewReader(imgBytes))

		// const dimension = 128
		// var resizedImages []*image.Paletted
		// var delays []int

		// originalPalette := gifData.Image[0].Palette

		// for _, img := range gifData.Image {
		// 	// Create a new image with the new size
		// 	newImg := image.NewRGBA(image.Rect(0, 0, dimension, dimension))

		// 	// Resize the current frame (img) to 256x256 using the nearest neighbor method
		// 	draw.NearestNeighbor.Scale(newImg, newImg.Rect, img, img.Bounds(), draw.Over, nil)

		// 	// Convert the resized RGBA image back to a paletted image using the original palette
		// 	bounds := img.Bounds()
		// 	palettedImg := image.NewPaletted(bounds, originalPalette)
		// 	draw.Draw(palettedImg, bounds, img, image.Point{}, draw.Over)

		// 	// Append the resized image and delay to the result arrays
		// 	resizedImages = append(resizedImages, palettedImg)
		// 	delays = append(delays, gifData.Delay[0])
		// }

		// var buf bytes.Buffer
		// _ = gif.EncodeAll(&buf, &gif.GIF{
		// 	Image: resizedImages,
		// 	Delay: delays,
		// })
		// imgBytes = buf.Bytes()
	}

	hash := sha256.Sum256(imgBytes)
	fileName := hex.EncodeToString(hash[:]) + extension
	var filePath = "./public/content/avatars/" + fileName

	// check if avatar pic file exists already, otherwise save as new
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		log.Trace("Profile pic [%s] doesn't exist yet, creating...", fileName)
		err = os.WriteFile(filePath, imgBytes, 0644)
		if err != nil {
			log.FatalError(err.Error(), "Error writing bytes to [%s] file from user ID [%d]", picType, userID)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		log.FatalError(err.Error(), "Error creating file for [%s] from user ID [%d]", picType, userID)
		http.Error(w, "", http.StatusInternalServerError)
		return
	} else {
		log.Trace("[%s] [%s] of same hash already exists, using that one...", picType, fileName)
	}

	if picType == "profile-pic" {
		success := database.UpdateUserValue(userID, fileName, "picture")
		if !success {
			log.Warn("Failed updating profile picture of user ID [%d] in database", userID)
			return
		}
		websocket.OnProfilePicChanged(userID, fileName)
	} else if picType == "server-pic" {
		serverID, err := strconv.ParseUint(r.FormValue("serverID"), 10, 64)
		if err != nil {
			log.WarnError(err.Error(), "Error parsing serverID as uint64 while changing picture of server of user ID [%d]", userID)
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		log.Trace("User ID [%d] wants to change picture of server ID [%d]", userID, serverID)
		success := database.ChangeServerPic(userID, serverID, fileName)
		if !success {
			log.Hack("Failed updating picture of server ID [%d] requested by user ID [%d]", serverID, userID)
			http.Error(w, "Failed updating picture of server", http.StatusForbidden)
			return
		}

		websocket.OnServerPicChanged(serverID, fileName)
	}

}

func checkAttachmentHandler(w http.ResponseWriter, r *http.Request) {
	userID := token.CheckIfTokenIsValid(w, r)
	if userID == 0 {
		http.Error(w, "", http.StatusUnauthorized)
		log.Hack("Someone is trying to check if an attachment file exists without token")
		return
	}

	log.Trace("User ID [%d] is checking if an attachment file exists", userID)

	var hashes [][32]byte
	_ = json.NewDecoder(r.Body).Decode(&hashes)

	log.Trace("User ID [%d] sent hashes of [%d] attachment(s) to check if they exist", userID, len(hashes))

	var hashesThatExist [][32]byte

	for i := 0; i < len(hashes); i++ {
		exists := database.CheckIfAttachmentExists(hashes[i][:])
		if exists {
			hashesThatExist = append(hashesThatExist, hashes[i])
		}
	}

	jsonBytes, err := json.Marshal(hashesThatExist)
	if err != nil {
		log.FatalError(err.Error(), "Error serializing attachment hashes that were confirmed for user ID [%d]", userID)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		log.WarnError(err.Error(), "Error sending attachment hashes that were confirmed for user ID [%d]", userID)
	}
}

func uploadAttachmentHandler(w http.ResponseWriter, r *http.Request) {
	userID := token.CheckIfTokenIsValid(w, r)
	if userID == 0 {
		http.Error(w, "", http.StatusUnauthorized)
		log.Hack("Someone is trying to upload an attachment without token")
		return
	}

	//err := r.ParseMultipartForm(100 << 20) // 10 MB limit
	//if err != nil {
	//	http.Error(w, "Unable to parse form", http.StatusBadRequest)
	//	return
	//}

	log.Trace("User ID [%d] is uploading an attachment", userID)

	reader, err := r.MultipartReader()
	if err != nil {
		log.WarnError(err.Error(), "Error parsing attachment from form file sent by user ID [%d]", userID)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	var awaitingAttachments []attachments.UploadedAttachment

	for {
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				http.Error(w, "", http.StatusBadRequest)
				return
			}
		}

		//if part.FileName() == "" {
		//	log.Warn("An attachment sent by user ID [%d] doesn't have a filename", userID)
		//	continue
		//}

		log.Trace("Name: %s", part.FormName())

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, part)
		if err != nil {
			log.WarnError(err.Error(), "Error reading attachment sent by user ID [%d]", userID)
			continue
		}

		if part.FormName() == "a" { // attachment
			log.Trace("User ID [%d] uploaded an attachment, reading and saving it...", userID)
			awaitingAttachment := attachments.UploadedAttachment{
				Hash: sha256.Sum256(buf.Bytes()),
				Name: part.FileName(),
			}

			hashString := hex.EncodeToString(awaitingAttachment.Hash[:])
			extension := filepath.Ext(part.FileName())
			path := "./public/content/attachments/" + hashString + extension

			_, err = os.Stat(path)
			if err == nil {
				log.Trace("Attachment at path [%s] already exists", path)
			} else if os.IsNotExist(err) {
				log.Trace("Attachment at path [%s] doesn't exist, creating...", path)
				err = os.WriteFile(path, buf.Bytes(), 0644)
				if err != nil {
					log.FatalError(err.Error(), "Error writing to file [%s]", path)
					return
				}
			} else {
				log.FatalError(err.Error(), "Error checking file [%s]", path)
			}

			awaitingAttachments = append(awaitingAttachments, awaitingAttachment)
		} else if part.FormName() == "h" { //hash
			log.Trace("User ID [%d] sent name and hash of file that already exists on server", userID)

			type ExistingAttachment struct {
				Hash [32]byte
				Name string
			}

			var existingAttachment ExistingAttachment

			err := json.Unmarshal(buf.Bytes(), &existingAttachment)
			if err != nil {
				log.WarnError(err.Error(), "Error decoding hash from form file sent by user ID [%d]", userID)
				continue
			}

			log.Trace("Name: [%s]", existingAttachment.Name)
			log.Trace("Hash: [%s]", hex.EncodeToString(existingAttachment.Hash[:]))

			awaitingAttachment := attachments.UploadedAttachment{
				Hash: existingAttachment.Hash,
				Name: existingAttachment.Name,
			}
			awaitingAttachments = append(awaitingAttachments, awaitingAttachment)
		}

	}

	log.Trace("[%s] POST request response was sent", r.URL.Path)

	attachmentToken := attachments.OnAttachmentUploaded(userID, awaitingAttachments)
	encoded := base64.StdEncoding.EncodeToString(attachmentToken[:])

	log.Trace("Response for [%s] POST request: [%s]", r.URL.Path, encoded)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte(fmt.Sprintf(`{"AttToken":"%s"}`, encoded)))
	if err != nil {
		log.WarnError(err.Error(), "Error sending %s POST request response", r.URL.Path)
	}
}
