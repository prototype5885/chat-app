package webRequests

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/snowflake"
	"proto-chat/modules/structs"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func printReceivedRequest(url string, method string) {
	log.Trace("Received %s %s request", url, method)
}

func respondText(w http.ResponseWriter, response string, v ...any) {
	_, err := fmt.Fprintf(w, response+"\n", v...)
	if err != nil {
		log.Error(err.Error())
	}
}

func redirect(w http.ResponseWriter, r *http.Request, target string) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	http.Redirect(w, r, target, http.StatusFound)
}

func getHtmlFilePath(name string) string {
	var htmlFilePath string = fmt.Sprintf("%s%s", publicFolder, name)
	return htmlFilePath
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

func loginOrRegister(bodyBytes []byte, pathURL string) (http.Cookie, structs.Result) {
	// deserialize the body message into LoginData struct
	type LoginData struct {
		Username string
		Password string
	}
	var loginData LoginData
	jsonErr := json.Unmarshal(bodyBytes, &loginData)
	if jsonErr != nil {
		return http.Cookie{}, structs.Result{
			Success: false,
			Message: "Error deserializing received loginData json from POST request",
		}
	}

	// decode password from base64 string to byte array so bcrypt can hash it, password is in SHA512 format
	// so the server can't really know what the original password was
	passwordBytes, err := base64.StdEncoding.DecodeString(loginData.Password)
	if err != nil {
		return http.Cookie{}, structs.Result{
			Success: false,
			Message: "Error decoding base64 password to byte array",
		}
	}

	// the values received next will be stored in this
	var logRegResult structs.Result
	var userID uint64

	// run depending on if its registration or login request
	if pathURL == "/register" {
		userID, logRegResult = registerUser(loginData.Username, passwordBytes)
	} else if pathURL == "/login" {
		userID = loginUser(loginData.Username, passwordBytes)
		if userID == 0 {
			logRegResult = structs.Result{
				Success: false,
				Message: "Wrong username or password",
			}
		} else {
			logRegResult = structs.Result{
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
		var token database.Token = newToken(userID)
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
func registerUser(username string, passwordBytes []byte) (uint64, structs.Result) {
	log.Info("Starting registration for new user with name [%s]...", username)

	// check if received password is in proper format
	if len(passwordBytes) != 64 {
		return 0, structs.Result{
			Success: false,
			Message: "Password byte array length isn't 64 bytes",
		}
	} else if len(username) > 16 {
		return 0, structs.Result{
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
		return 0, structs.Result{
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
	var user = database.User{
		UserID:      userID,
		Username:    username,
		DisplayName: "placeholder name",
		Picture:     "default_profilepic.webp",
		Password:    passwordHash,
		Totp:        "",
	}

	newUserResult := database.Insert(user)
	if !newUserResult {
		return 0, structs.Result{
			Success: false,
			Message: "Registration failed",
		}
	}

	// return the Success
	return userID, structs.Result{
		Success: true,
		Message: "Successful registration",
	}
}

// Login user, first checking if username exists in the database, then getting the password
// hash and checking if user entered the correct password, returns the user's ID.
func loginUser(username string, passwordBytes []byte) uint64 {
	log.Debug("Starting login of user [%s]...", username)

	// get the password hash from the database
	passwordHash, userID := database.GetPasswordAndID(username)
	if passwordHash == nil {
		log.Warn("No user was found with username [%s]", username)
		return 0
	}

	// compare given password with the retrieved hash
	log.Debug("Comparing password hash and string for user [%s]...", username)
	var start = time.Now().UnixMilli()
	if err := bcrypt.CompareHashAndPassword(passwordHash, passwordBytes); err != nil {
		log.Warn("User entered wrong password for username [%s]", username)
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

func newTokenExpiration() uint64 {
	return uint64(time.Now().Add(30 * 24 * time.Hour).Unix()) // 30 days from current time
}

func newToken(userID uint64) database.Token {
	log.Debug("Generating new token for user ID [%d]...", userID)

	// generate new token
	var tokenBytes []byte = make([]byte, 128)
	_, err := io.ReadFull(rand.Reader, tokenBytes)
	if err != nil {
		log.FatalError(err.Error(), "Error generating token for user ID [%d]", userID)
	}

	var token = database.Token{
		Token:      tokenBytes,
		UserID:     userID,
		Expiration: newTokenExpiration(),
	}

	// add the newly generated token into the database
	database.Insert(token)

	// return the new token
	return token
}

func CheckIfTokenIsValid(w http.ResponseWriter, r *http.Request) uint64 {
	cookieToken, found := findCookie(r.Cookies(), "token")
	if found { // if user has a token
		// decode to bytes
		tokenBytes, err := hex.DecodeString(cookieToken.Value)
		if err != nil {
			log.WarnError(err.Error(), "Error decoding token from cookie to byte array")
			return 0
		}

		userID, _ := database.ConfirmToken(tokenBytes)

		// check if expired already
		// log.Debug("%d", expiration)

		// renew the token
		if userID != 0 {
			log.Trace("Renewing cookie for user ID [%d]", userID)
			var newExpiration uint64 = newTokenExpiration()
			database.RenewTokenExpiration(newExpiration, tokenBytes)
			var cookie = http.Cookie{
				Name:     "token",
				Value:    cookieToken.Value,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				Secure:   true,
				Expires:  time.Unix(int64(newExpiration), 0),
			}
			http.SetCookie(w, &cookie)
		}
		// check if token exists in the database
		return userID
	}
	return 0
}

// func MergeJsFilesMem() string {
// 	var javascript string
// 	// loop through javascript files and
// 	log.Trace("Loop through javascript files...")
// 	log.Trace("Amount of javascript files: [%d]", len(jsFilePaths))
// 	for _, file := range jsFilePaths {
// 		log.Trace("Opening javascript file [%s]", file)
// 		inFile, err := os.Open("./public/jsSrc/" + file)
// 		if err != nil {
// 			log.FatalError(err.Error(), "Error opening javascript file [%s]", file)
// 		}
// 		defer inFile.Close()

// 		// writes the filename before copying
// 		// _, err = outFile.WriteString(fmt.Sprintf("// %s\n\n", file))
// 		// if err != nil {
// 		// 	log.FatalError(err.Error(), "Error writing filename between merged contents in script.js")
// 		// }

// 		javascript += fmt.Sprintf("// %s\n\n", file)

// 		// Copy the contents of the input file to the output file
// 		// log.Trace("Writing content of javascript file [%s] into script.js", file)
// 		// _, err = io.Copy(outFile, inFile)
// 		// if err != nil {
// 		// 	log.FatalError(err.Error(), "Error merging javascript files")
// 		// }

// 		content, err := os.ReadFile(inFile.Name())
// 		if err != nil {
// 			log.Fatal(err.Error())
// 		}
// 		javascript += string(content)

// 		// Add a newline to separate the contents of different files
// 		// _, err = outFile.WriteString("\n\n")
// 		// if err != nil {
// 		// 	log.FatalError(err.Error(), "Error adding newlines after copying javascript content into script.js")
// 		// }
// 		javascript += "\n\n"
// 	}

// 	log.Info("JavaScript files merged successfully into javascript string")
// 	return javascript
// }
