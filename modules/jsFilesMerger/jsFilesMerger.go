package jsfilesmerger

import (
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	log "proto-chat/modules/logging"
	"strings"
)

const DynamicMergedJsGeneration = true

var jsFilePaths []string = []string{
	"main.js",
	"notification.js",
	"localStorage.js",
	"comp/httpRequests.js",
	"comp/contextMenu.js",
	"comp/bubble.js",
	"comp/serverList.js",
	"comp/memberList.js",
	"comp/channelList.js",
	"comp/chatMessageList.js",
	"comp/window.js",
	"comp/chatInput.js",
	"dynamicContent.js",
	"websocket.js",
}

var jsFiles = make(map[string][20]byte)

// type JsFile struct {
// 	FilePath string
// 	SHA1 [20]byte
// }

const jsHashesFilename string = "jsHashes.bin"

func Init() {
	if !DynamicMergedJsGeneration {
		return
	}

	_, err := os.Stat("./public/js/script.js")
	if os.IsNotExist(err) {
		log.WarnError(err.Error(), "script.js doesn't exist, creating it...")
		merge()
	}

	loadHashesFromFile()
	CheckForChanges()
}

func CheckForChanges() {
	if !DynamicMergedJsGeneration {
		return
	}

	var changed bool = false
	for i := 0; i < len(jsFilePaths); i++ {
		// open the javascript file
		jsFile, err := os.Open("./public/jsSrc/" + jsFilePaths[i])
		if err != nil {
			log.FatalError(err.Error(), "Error opening javascript file [%s]", jsFilePaths[i])
		}
		defer jsFile.Close()

		// read its content into byte array
		data, err := io.ReadAll(jsFile)
		if err != nil {
			log.FatalError(err.Error(), "Error reading javascript file [%s]", jsFilePaths[i])
		}
		sha1 := sha1.Sum(data)

		if jsFiles[jsFilePaths[i]] == sha1 {
			// if checksum matches with the one in hashmap
			log.Trace("SHA1 of [%s] matches with the one in hashmap", jsFilePaths[i])
		} else {
			// if doesn't match, overwrite with new one
			log.Warn("SHA1 of [%s] doesn't match with the one in hashmap, overwriting...", jsFilePaths[i])
			jsFiles[jsFilePaths[i]] = sha1
			changed = true
		}
	}
	// regenerate merged script.js
	if changed {
		saveHashesToFile()
		merge()
	} else {
		log.Info("SHA1 hashes of all javascript files match with the ones in %s", jsHashesFilename)
	}
}

func saveHashesToFile() {
	file, err := os.Create(jsHashesFilename)
	if err != nil {
		log.FatalError(err.Error(), "Error creating %s", jsHashesFilename)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(jsFiles)
	if err != nil {
		log.FatalError(err.Error(), "Error writing into %s", jsHashesFilename)
	}
	log.Debug("SHA1 hashes of javascript files have been written into %s", jsHashesFilename)
}

func loadHashesFromFile() {
	file, err := os.Open(jsHashesFilename)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			log.Warn("%s doesn't exist yet", jsHashesFilename)
			return
		}
		log.FatalError(err.Error(), "Error opening %s", jsHashesFilename)
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&jsFiles)
	if err != nil {
		log.FatalError(err.Error(), "Error reading from %s", jsHashesFilename)
	}

	log.Debug("Loading SHA1 hashes of javascript files from %s", jsHashesFilename)
}

func merge() {
	var err error
	// Create a new file to store the merged JavaScript code
	log.Trace("Create script.js that will store all js content")
	var outFile *os.File
	outFile, err = os.OpenFile("./public/js/script.js", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.FatalError(err.Error(), "Error creating script.js")
	}
	defer outFile.Close()

	// loop through javascript files and
	log.Trace("Loop through javascript files...")
	log.Trace("Amount of javascript files: [%d]", len(jsFilePaths))
	for _, file := range jsFilePaths {
		log.Trace("Opening javascript file [%s]", file)
		inFile, err := os.Open("./public/jsSrc/" + file)
		if err != nil {
			log.FatalError(err.Error(), "Error opening javascript file [%s]", file)
		}
		defer inFile.Close()

		// writes the filename before copying
		_, err = outFile.WriteString(fmt.Sprintf("// %s\n\n", file))
		if err != nil {
			log.FatalError(err.Error(), "Error writing filename between merged contents in script.js")
		}

		// reading javascript file
		// log.Trace("Reading javascript file [%s]", file)
		// var data []byte
		// data, err = io.ReadAll(inFile)
		// if err != nil {
		// log.FatalError(err.Error(), "Error reading [%s]", file)
		// }

		// removing notes
		// log.Trace("Removing notes from [%s]", file)
		// var trimmedJs string
		// trimmedJs = regexp.MustCompile(`// .*`).ReplaceAllString(string(data), "")

		// remove multiple new lines
		// trimmedJs = regexp.MustCompile(` +`).ReplaceAllString(trimmedJs, "")

		// remove multiple spaces
		// trimmedJs = regexp.MustCompile(`[\n]{2,}|[ ]{2,}`).ReplaceAllString(trimmedJs, "")

		// add into script.js
		// log.Trace("Writing content of javascript file [%s] into script.js", file)
		// _, err = outFile.WriteString(trimmedJs)
		// if err != nil {
		// 	log.Fatal(err.Error(), "Error writing [%s] into [%s]", file, jsHashesFilename)
		// }

		// copy the contents of the javascript file into the merged one
		log.Trace("Writing content of javascript file [%s] into script.js", file)
		_, err = io.Copy(outFile, inFile)
		if err != nil {
			log.FatalError(err.Error(), "Error merging javascript files")
		}

		// add a newline to separate the contents of different files
		_, err = outFile.WriteString("\n\n")
		if err != nil {
			log.FatalError(err.Error(), "Error adding newlines after copying javascript content into script.js")
		}
	}

	log.Info("JavaScript files merged successfully into script.js")
}
