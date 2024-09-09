package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"proto-chat/modules/snowflake"
	"strconv"
	"time"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

func setupLogging(logInFile bool) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	if logInFile {
		err := os.MkdirAll("logs", fs.FileMode(os.ModePerm))
		if err != nil {
			log.Fatal("Error creating log folder:", err)
		}
		file, err := os.OpenFile("./logs/protochat.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal("Error opening log file:", err)
		}
		log.SetOutput(file)
	}
}

func readConfigFile() ConfigFile {
	configFile := "config.json"
	file, err := os.Open(configFile)
	if err != nil {
		log.Panicln("Error opening config file:", err)
		return ConfigFile{}
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Panicln("Error closing config file:", err)
		}
	}(file)

	var config ConfigFile
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Panicln("Error decoding config file:", err)
		return ConfigFile{}
	}
	return config
}

func findCookie(cookies []*http.Cookie, cookieName string) (http.Cookie, Result) {
	log.Printf("Searching for cookie called: %s...\n", cookieName)

	for _, cookie := range cookies {
		// log.Printf("Cookie: %s=%s\n", cookie.Name, cookie.Value)
		if cookie.Name == cookieName {
			return *cookie, Result{
				Success: true,
				Message: "",
			}
		}
	}
	return http.Cookie{}, Result{
		Success: false,
		Message: "No cookie with the following name was found: " + cookieName,
	}
}

func loginOrRegister(bodyBytes []byte, pathURL string) (http.Cookie, Result) {
	// deserialize the body message into LoginData struct
	type LoginData struct {
		Username string
		Password string
	}
	var loginData LoginData
	jsonErr := json.Unmarshal(bodyBytes, &loginData)
	if jsonErr != nil {
		return http.Cookie{}, Result{
			Success: false,
			Message: "Error deserializing received loginData json from POST request",
		}
	}

	// decode password from base64 string to byte array so bcrypt can hash it, password is in SHA512 format
	// so the server can't really know what the original password was
	passwordBytes, err := base64.StdEncoding.DecodeString(loginData.Password)
	if err != nil {
		return http.Cookie{}, Result{
			Success: false,
			Message: "Error decoding base64 password to byte array",
		}
	}

	// the values received next will be stored in this
	var logRegResult Result
	var userID uint64

	// run depending on if its registration or login request
	if pathURL == "/register" {
		userID, logRegResult = registerUser(loginData.Username, passwordBytes)
	} else if pathURL == "/login" {
		userID, logRegResult = loginUser(loginData.Username, passwordBytes)
	} else {
		// this is not supposed to happen ever
		fatalWithName(loginData.Username, "Invalid path URL:"+pathURL, "")
	}

	// generate token if login or registration was success, otherwise it will remain empty
	var cookie http.Cookie
	if logRegResult.Success {
		token, tokenResult := newToken(userID)
		if !tokenResult.Success {
			fatalWithName(loginData.Username, "Error generating token", tokenResult.Message)
		} else {
			cookie = http.Cookie{
				Name:     "token",
				Value:    token.Token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				Secure:   true,
				Expires:  time.Unix(int64(token.Expiration), 0),
			}
		}
	}
	printWithID(userID, logRegResult.Message)
	return cookie, logRegResult
}

// Register user by adding it into the database
func registerUser(username string, passwordBytes []byte) (uint64, Result) {
	printWithName(username, "Starting registration...")

	// check if received password is in proper format
	if len(passwordBytes) != 64 {
		return 0, Result{
			Success: false,
			Message: "Password byte array length isn't 64 bytes",
		}
	} else if len(username) > 16 {
		return 0, Result{
			Success: false,
			Message: "Username is longer than 16 bytes",
		}
	}

	// hash the password using bcrypt
	var start int64 = time.Now().UnixMilli()
	// printWithName(username, "Hashing password...")
	passwordHash, err := bcrypt.GenerateFromPassword(passwordBytes, 10)
	if err != nil {
		var errMsg string
		if err.Error() == "bcrypt: password length exceeds 72 bytes" {
			errMsg = err.Error()
		} else {
			errMsg = "Error generating bcrypt hash"
		}
		return 0, Result{
			Success: false,
			Message: errMsg,
		}
	}
	printWithName(username, fmt.Sprintf("%s: password hashing took: %d ms", username, time.Now().UnixMilli()-start))

	// generate userID
	var userID uint64 = snowflake.Generate()

	// generate TOTP secret key
	//totpKey, totpResult := generateTOTP(userID)
	//if !totpResult.Success {
	//	return 0, totpResult
	//}
	//printWithName(username, totpResult.Message)

	// add the new user to database
	newUserResult := registerNewUserToDB(userID, username, passwordHash, "")
	if !newUserResult.Success {
		return 0, newUserResult
	}

	// return the Success
	return userID, Result{
		Success: true,
		Message: "Successful registration",
	}
}

// Login user, first checking if username exists in the database, then getting the password
// hash and checking if user entered the correct password, returns the user's ID.
func loginUser(username string, passwordBytes []byte) (uint64, Result) {
	printWithName(username, "Starting login...")

	// get the user id from the database
	userID, result := getUserIdFromDB(username)
	if !result.Success {
		return 0, result
	}
	printWithName(username, "Confirmed to be: "+strconv.FormatUint(userID, 10))

	// get the password hash from the database
	passwordHash, result := getPasswordFromDB(userID)
	if !result.Success {
		return 0, result
	}

	// compare given password with the retrieved hash
	printWithID(userID, "Comparing password hash and string...")
	var start = time.Now().UnixMilli()
	if err := bcrypt.CompareHashAndPassword(passwordHash, passwordBytes); err != nil {
		return 0, Result{
			Success: false,
			Message: "Wrong password",
		}
	}

	log.Printf("%s: password matches with hash, comparison took: %d ms\n", username, time.Now().UnixMilli()-start)

	// return the Success
	return userID, Result{
		Success: true,
		Message: "Successful login",
	}
}

func addChatMessage(jsonBytes []byte, userID uint64, displayName string) []byte {
	type ClientChatMsg struct {
		ChanID  uint64
		Name    string
		ChatMsg string
	}

	var clientChatMsg ClientChatMsg

	if err := json.Unmarshal(jsonBytes, &clientChatMsg); err != nil {
		log.Println("Error deserializing Msg json:", err)
		return []byte("Error deserializing Msg json")
	}

	//log.Println(clientChatMsg.ChannelId)
	//log.Println(clientChatMsg.ChatMsg)

	type ServerChatMsg struct {
		MsgID  uint64
		ChanID uint64
		UserID uint64
		Name   string
		Msg    string
	}

	var serverChatMsg = ServerChatMsg{
		MsgID:  snowflake.Generate(),
		ChanID: snowflake.Generate(),
		UserID: userID,
		Name:   displayName,
		Msg:    clientChatMsg.ChatMsg,
	}

	result := addChatMessageDB(serverChatMsg.MsgID, serverChatMsg.ChanID, serverChatMsg.UserID, serverChatMsg.Msg)
	if !result.Success {

	}

	jsonBytes, err := json.Marshal(serverChatMsg)
	if err != nil {
		log.Fatal("Error serializing ServerChatMsg json:", err)
		return []byte("Error serializing ServerChatMsg json")
	}

	var typeByte byte = 1
	var packet []byte = make([]byte, len(jsonBytes)+1)

	packet[0] = typeByte
	copy(packet[1:], jsonBytes)

	//for _, value := range packet {
	//	fmt.Printf("%d ", value)
	//}

	return packet
}

func generateTOTP(userID uint64) (string, Result) {
	printWithID(userID, "Generating TOTP secret key...")

	totpKey, err := totp.Generate(totp.GenerateOpts{
		AccountName: strconv.FormatUint(userID, 10),
		Issuer:      "ProToType",
	})
	if err != nil {
		log.Fatal(err)
		//return "", Result{
		//	Success: false,
		//	Message: "Error generating TOTP secret key",
		//}
	}
	return totpKey.Secret(), Result{
		Success: true,
		Message: "TOTP secret key generated",
	}
}

func newToken(userID uint64) (Token, Result) {
	printWithID(userID, "Generating new token...")

	// generate new token
	var tokenBytes []byte = make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, tokenBytes)
	if err != nil {
		return Token{}, Result{
			Success: false,
			Message: err.Error(),
		}
	}

	var tokenRow = Token{
		Token:      hex.EncodeToString(tokenBytes),
		UserID:     userID,
		Expiration: uint64(time.Now().Add(30 * 24 * time.Hour).Unix()), // 3 months
	}

	// add the newly generated token into the database
	result := addTokenDB(tokenRow)
	if !result.Success {
		return Token{}, result
	}
	printWithID(userID, result.Message)

	// return the new token
	return tokenRow, Result{
		Success: true,
		Message: "Successfully generated and added new token",
	}
}

func checkIfTokenIsValid(r *http.Request) (uint64, Result) {
	log.Println("Checking if received token is valid...")

	cookieToken, cookieResult := findCookie(r.Cookies(), "token")
	if cookieResult.Success { // if user has a token
		// check if token exists in the database
		token, result := getTokenFromDB(cookieToken.Value)
		if result.Success {
			return token.UserID, result
		} else {
			return 0, result
		}
	}
	return 0, cookieResult
}
