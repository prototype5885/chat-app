package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"proto-chat/modules/snowflake"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func setupLogging(logInFile bool) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	if logInFile {
		err := os.MkdirAll("logs", fs.FileMode(os.ModePerm))
		if err != nil {
			log.Panic("Error creating log folder:", err)
		}
		file, err := os.OpenFile("./logs/protochat.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Panic("Error opening log file:", err)
		}
		log.SetOutput(file)
	}
}

func readConfigFile() ConfigFile {
	configFile := "config.json"
	file, err := os.Open(configFile)
	if err != nil {
		log.Panicln("Error opening config file:", err)
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
	}
	return config
}

func findCookie(cookies []*http.Cookie, cookieName string) (http.Cookie, bool) {
	log.Printf("Searching for cookie called: %s...\n", cookieName)

	for _, cookie := range cookies {
		log.Printf("Cookie: %s=%s\n", cookie.Name, cookie.Value)
		if cookie.Name == cookieName {
			return *cookie, true
		}
	}
	log.Printf("No cookie with the following name was found: [%s]\n", cookieName)
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
		log.Panicf("Invalid path URL for user [%s], %s\n", loginData.Username, pathURL)
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
	log.Printf("User ID [%d]: %s\n", userID, logRegResult.Message)
	return cookie, logRegResult
}

// Register user by adding it into the database
func registerUser(username string, passwordBytes []byte) (uint64, Result) {
	log.Printf("Starting registration for new user with name [%s]...\n", username)

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
	log.Printf("Password hashing for user [%s] took %d ms,", username, time.Now().UnixMilli()-start)

	// generate userID
	var userID uint64 = snowflake.Generate()

	// generate TOTP secret key
	//totpKey, totpResult := generateTOTP(userID)
	//if !totpResult.Success {
	//	return 0, totpResult
	//}
	//printWithName(username, totpResult.Message)

	// add the new user to database
	newUserResult := database.RegisterNewUser(userID, username, passwordHash, "")
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
	log.Printf("Starting login of user [%s]...\n", username)

	// get the password hash from the database
	passwordHash, userID := database.GetPasswordAndID(username)
	if passwordHash == nil {
		log.Printf("No user was found with username [%s]\n", username)
		return 0
	}

	// compare given password with the retrieved hash
	log.Printf("Comparing password hash and string for user [%s]...\n", username)
	var start = time.Now().UnixMilli()
	if err := bcrypt.CompareHashAndPassword(passwordHash, passwordBytes); err != nil {
		log.Printf("User entered wrong password for username [%s]\n", username)
		return 0
	}

	log.Printf("%s: password matches with hash, comparison took: %d ms\n", username, time.Now().UnixMilli()-start)

	// return the Success
	return userID
}

// func generateTOTP(userID uint64) string {
// 	log.Printf("Generating TOTP secret key for user ID [%d]...\n", userID)

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
	log.Printf("Generating new token for user ID [%d]...\n", userID)

	// generate new token
	var tokenBytes []byte = make([]byte, 128)
	_, err := io.ReadFull(rand.Reader, tokenBytes)
	if err != nil {
		log.Println(err.Error())
		log.Panicf("Error generating token for user ID [%d]\n", userID)
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
	log.Println("Checking if received token is valid...")

	cookieToken, found := findCookie(r.Cookies(), "token")
	if found { // if user has a token
		// decode to bytes
		tokenBytes, err := hex.DecodeString(cookieToken.Value)
		if err != nil {
			log.Println(err.Error())
			log.Println("Error decoding token from cookie to byte array")
			return 0
		}

		// check if token exists in the database
		return database.ConfirmToken(tokenBytes)
		// var userID uint64 = database.ConfirmToken(tokenBytes)
		// if userID == 0 {
		// 	return 0
		// } else {
		// 	return userID
		// }
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

	log.Println("Prepared packet:", endIndex, packet[4], string(jsonBytes))

	return packet
}

func setProblem(problem string) []byte {
	type Issue struct {
		Issue string
	}
	var issue = Issue{
		Issue: problem,
	}

	json, err := json.Marshal(issue)
	if err != nil {
		log.Printf("Could not serialize issue in formatProblem\n")
		log.Panicln(err.Error())
	}

	return preparePacket(0, json)
}
