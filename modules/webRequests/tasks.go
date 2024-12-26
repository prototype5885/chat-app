package webRequests

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"io"
	"net/http"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func printReceivedRequest(url string, method string) {
	log.Trace("Received %s %s request", url, method)
}

func respondText(w http.ResponseWriter, response string, v ...any) {
	_, err := fmt.Fprintf(w, response+"\n", v...)
	if err != nil {
		log.Error("%s", err.Error())
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

func GenerateTOTP(userID uint64) (*otp.Key, image.Image) {
	log.Trace("Generating TOTP secret key for user ID [%d]...", userID)

	generateOpts := totp.GenerateOpts{
		AccountName: database.GetUsername(userID),
		Issuer:      "ProToChat",
	}

	totpKey, err := totp.Generate(generateOpts)
	if err != nil {
		log.FatalError(err.Error(), "Error generating TOTP for user ID [%d]", userID)
	}

	image, err := totpKey.Image(256, 256)
	if err != nil {
		log.FatalError(err.Error(), "Error generating TOTP QR code for user ID [%d]", userID)
	}

	return totpKey, image
}

func newTokenExpiration() int64 {
	return time.Now().Add(30 * 24 * time.Hour).Unix() // 30 days from current time
}

func newTokenCookie(userID uint64) http.Cookie {
	log.Debug("Generating new token for user ID [%d]...", userID)

	// generate new token
	tokenBytes := make([]byte, 128)
	_, err := io.ReadFull(rand.Reader, tokenBytes)
	if err != nil {
		log.FatalError(err.Error(), "Error generating token for user ID [%d]", userID)
	}

	token := database.Token{
		Token:      tokenBytes,
		UserID:     userID,
		Expiration: newTokenExpiration(),
	}

	// add the newly generated token into the database
	err = database.Insert(token)
	if err != nil {
		log.Fatal("Failed inserting new generated token for user ID [%d] into database", userID)
	}

	cookie := http.Cookie{
		Name:     "token",
		Value:    hex.EncodeToString(token.Token),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
		Expires:  time.Unix(int64(token.Expiration), 0),
	}
	return cookie
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

		token := database.ConfirmToken(tokenBytes)

		// check if token expired
		currentTime := time.Now().Unix()
		if token.Expiration != 0 && currentTime > token.Expiration {
			difference := currentTime - token.Expiration
			daysPassed := difference / 60 / 60 / 24
			log.Trace("Token [%s] of user ID [%d] expired [%d] days ago", macros.ShortenToken(tokenBytes), token.UserID, daysPassed)
			success := database.Delete(token)
			if !success {
				log.Impossible("How did deleting a confirmed token fail?")
			}
			return 0
		}

		// renew the token if token was found
		if token.UserID != 0 {
			log.Trace("Renewing cookie for user ID [%d]", token.UserID)
			var newExpiration int64 = newTokenExpiration()
			success := database.RenewTokenExpiration(newExpiration, tokenBytes)
			if !success {
				log.Error("Failed renewing token for user ID [%d]", token.UserID)
				return 0
			}
			var cookie = http.Cookie{
				Name:     "token",
				Value:    cookieToken.Value,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				Secure:   true,
				Expires:  time.Unix(newExpiration, 0),
			}
			http.SetCookie(w, &cookie)
		}
		// check if token exists in the database
		return token.UserID
	}
	return 0
}
