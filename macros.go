package main

import (
	"log"
	"strconv"
)

func printWithName(username string, text string) {
	log.Printf("[%s]: %s\n", username, text)
}

func fatalWithName(username string, text string, err string) {
	log.Fatalf("[%s]: %s: %s\n", username, text, err)
}

func printWithID(userID uint64, text string) {
	log.Printf("[%d]: %s\n", userID, text)
}

func fatalWithID(userID uint64, text string, err string) {
	log.Fatalf("[%d]: %s: %s\n", userID, text, err)
}

func noUserIdFoundText(userID uint64) string {
	return "No user found with given id: " + strconv.FormatUint(userID, 10)
}

func noUsernameFoundText(username string) string {
	return "No user found with given name: " + username
}

func printReceivedRequest(url string, method string) {
	log.Printf("Received %s %s request\n", url, method)
}
