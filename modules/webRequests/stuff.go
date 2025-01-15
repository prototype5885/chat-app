package webRequests

import (
	"fmt"
	"image"
	"net/http"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func printReceivedRequest(r *http.Request) {
	log.Trace("Received %s %s request from IP address %s", r.URL.Path, r.Method, r.RemoteAddr)
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
