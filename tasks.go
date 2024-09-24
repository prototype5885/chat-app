package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	log "proto-chat/modules/logging"
	"proto-chat/modules/snowflake"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func errorDeserializing(errStr string, jsonType string, userID uint64) []byte {
	log.Error(errStr)
	log.Warn("Error deserializing json type [%s] of user ID [%d]", jsonType, userID)
	return respondFailureReason(fmt.Sprintf("Couldn't deserialize json of [%s] request", jsonType))
}

func errorSerializing(errStr string, jsonType string, userID uint64) {
	log.Error(errStr)
	log.Warn("Fatal error serializing response json type [%s] for user ID [%d]", jsonType, userID)
}

func readConfigFile() ConfigFile {
	configFile := "config.json"
	file, err := os.Open(configFile)
	if err != nil {
		log.Error(err.Error())
		log.Fatal("Error opening config file")

	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error(err.Error())
			log.Fatal("Error closing config file")
		}
	}(file)

	var config ConfigFile
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Error(err.Error())
		log.Fatal("Error decoding config file")
	}
	return config
}

func getTimestamp() int64 {
	return time.Now().UnixMilli()
}

func measureTime(start int64, msg string) {
	log.Time("%s took [%d ms]", msg, getTimestamp()-start)
}

func findCookie(cookies []*http.Cookie, cookieName string) (http.Cookie, bool) {
	log.Debug("Searching for cookie called: %s...", cookieName)

	for _, cookie := range cookies {
		log.Debug("Cookie called %s found", cookieName)
		log.Trace("%s=%s", cookie.Name, cookie.Value)
		if cookie.Name == cookieName {
			return *cookie, true
		}
	}
	log.Debug("No cookie with the following name was found: [%s]", cookieName)
	return http.Cookie{}, false
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
		userID = loginUser(loginData.Username, passwordBytes)
		if userID == 0 {
			logRegResult = Result{
				Success: false,
				Message: "Wrong username or password",
			}
		} else {
			logRegResult = Result{
				Success: true,
				Message: "Successful login",
			}
		}
	} else {
		// this is not supposed to happen ever
		log.Fatal("Invalid path URL for user [%s], %s", loginData.Username, pathURL)
	}

	// generate token if login or registration was success,
	// otherwise it will remain empty as it won't be needed
	var cookie http.Cookie
	if logRegResult.Success {
		var token Token = newToken(userID)
		cookie = http.Cookie{
			Name:     "token",
			Value:    hex.EncodeToString(token.Token),
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   true,
			Expires:  time.Unix(int64(token.Expiration), 0),
		}

	}
	log.Info("User ID [%d]: %s", userID, logRegResult.Message)
	return cookie, logRegResult
}

// Register user by adding it into the database
func registerUser(username string, passwordBytes []byte) (uint64, Result) {
	log.Info("Starting registration for new user with name [%s]...", username)

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
	log.Debug("Password hashing for user [%s] took %d ms,", username, time.Now().UnixMilli()-start)

	// generate userID
	var userID uint64 = snowflake.Generate()

	// generate TOTP secret key
	//totpKey, totpResult := generateTOTP(userID)
	//if !totpResult.Success {
	//	return 0, totpResult
	//}
	//printWithName(username, totpResult.Message)

	// add the new user to database
	newUserResult := database.RegisterNewUser(userID, username, "placeholder name", passwordHash, "")
	if !newUserResult {
		return 0, Result{
			Success: false,
			Message: "Registration failed",
		}
	}

	// return the Success
	return userID, Result{
		Success: true,
		Message: "Successful registration",
	}
}

// Login user, first checking if username exists in the database, then getting the password
// hash and checking if user entered the correct password, returns the user's ID.
func loginUser(username string, passwordBytes []byte) uint64 {
	log.Info("Starting login of user [%s]...", username)

	// get the password hash from the database
	passwordHash, userID := database.GetPasswordAndID(username)
	if passwordHash == nil {
		log.Info("No user was found with username [%s]", username)
		return 0
	}

	// compare given password with the retrieved hash
	log.Debug("Comparing password hash and string for user [%s]...", username)
	var start = time.Now().UnixMilli()
	if err := bcrypt.CompareHashAndPassword(passwordHash, passwordBytes); err != nil {
		log.Info("User entered wrong password for username [%s]", username)
		return 0
	}

	log.Debug("%s: password matches with hash, comparison took: %d ms", username, time.Now().UnixMilli()-start)

	// return the Success
	return userID
}

// func generateTOTP(userID uint64) string {
// 	log.Printf("Generating TOTP secret key for user ID [%d]...", userID)

// 	totpKey, err := totp.Generate(totp.GenerateOpts{
// 		AccountName: strconv.FormatUint(userID, 10),
// 		Issuer:      "ProToType",
// 	})
// 	if err != nil {
// 		log.Panic("Error generating TOTP:", err)
// 	}
// 	return totpKey.Secret()
// }

func newToken(userID uint64) Token {
	log.Debug("Generating new token for user ID [%d]...", userID)

	// generate new token
	var tokenBytes []byte = make([]byte, 128)
	_, err := io.ReadFull(rand.Reader, tokenBytes)
	if err != nil {
		log.Error(err.Error())
		log.Fatal("Error generating token for user ID [%d]", userID)
	}

	var tokenRow = Token{
		Token:      tokenBytes,
		UserID:     userID,
		Expiration: uint64(time.Now().Add(30 * 24 * time.Hour).Unix()), // 3 months
	}

	// add the newly generated token into the database
	database.AddToken(tokenRow)

	// return the new token
	return tokenRow
}

func checkIfTokenIsValid(r *http.Request) uint64 {
	cookieToken, found := findCookie(r.Cookies(), "token")
	if found { // if user has a token
		// decode to bytes
		tokenBytes, err := hex.DecodeString(cookieToken.Value)
		if err != nil {
			log.Error(err.Error())
			log.Warn("Error decoding token from cookie to byte array")
			return 0
		}

		// check if token exists in the database
		return database.ConfirmToken(tokenBytes)
	}
	return 0
}

func preparePacket(typeByte byte, jsonBytes []byte) []byte {
	// convert the end index uint32 value into 4 bytes
	var endIndex uint32 = uint32(5 + len(jsonBytes))
	var endIndexBytes []byte = make([]byte, 4)
	binary.LittleEndian.PutUint32(endIndexBytes, endIndex)

	// merge them into a single packet
	var packet []byte = make([]byte, 5+len(jsonBytes))
	copy(packet, endIndexBytes) // first 4 bytes will be the length
	packet[4] = typeByte        // 5th byte will be the packet type
	copy(packet[5:], jsonBytes) // rest will be the json byte array

	log.Trace("Prepared packet: endIndex [%d], type [%d], json [%s]", endIndex, packet[4], string(jsonBytes))

	return packet
}

func respondFailureReason(reason string) []byte {
	type Failure struct {
		Reason string
	}
	var failure = Failure{
		Reason: reason,
	}

	json, err := json.Marshal(failure)
	if err != nil {
		log.Error(err.Error())
		log.Fatal("Could not serialize issue in respondFailureReason")
	}

	return preparePacket(0, json)
}
