package main

import (
	"chat-app/modules/database"
	log "chat-app/modules/logging"
	"chat-app/modules/snowflake"
	"chat-app/modules/token"
	"chat-app/modules/webRequests"
	"chat-app/modules/websocket"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Starting server...")

	// handle termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Received termination signal...")
		err := database.CloseDatabaseConnection()
		if err != nil {
			log.Error("Error closing db connection")
		}
		fmt.Println("Closed main db connection successfully")
		os.Exit(0)
	}()

	// reading config file

	type ConfigFile struct {
		LocalhostOnly              bool
		Port                       uint32
		TLS                        bool
		LogConsole                 bool
		LogFile                    bool
		Sqlite                     bool
		ImageServerAddressWithPort string
		DatabaseAddress            string
		DatabasePort               uint32
		DatabaseUsername           string
		DatabasePassword           string
		DatabaseName               string
	}

	readConfigFile := func() ConfigFile {
		fmt.Println("Reading config file...")
		configFile := "config.json"
		file, err := os.Open(configFile)
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println("Error opening config file")
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				fmt.Println(err.Error())
				fmt.Println("Error closing config file")
			}
		}(file)

		var config ConfigFile
		err = json.NewDecoder(file).Decode(&config)
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println("Error decoding config file")
		}
		return config
	}

	config := readConfigFile()
	log.SetupLogging("TRACE", config.LogConsole, config.LogFile)

	//jsfilesmerger.Init()

	// database
	if config.Sqlite {
		database.ConnectSqlite()
	} else {
		database.ConnectMariadb(config.DatabaseUsername, config.DatabasePassword, config.DatabaseAddress, strconv.Itoa(int(config.DatabasePort)), config.DatabaseName)
	}
	database.CreateTables()

	// snowflake
	snowflake.SetSnowflakeWorkerID(0)

	// websocket
	websocket.Init()

	websocket.ImageHost = config.ImageServerAddressWithPort

	if websocket.ImageHost != "" {
		var err error
		websocket.ParsedImageHost, err = url.Parse(websocket.ImageHost)
		if err != nil {
			log.FatalError(err.Error(), "Error parsing image host URL")
			return
		}
		websocket.ImageHostAddress = strings.Split(websocket.ParsedImageHost.Host, ":")[0]
	}

	// handle http requests
	http.HandleFunc("/", webRequests.MainHandler)

	// maintenance goroutine
	go maintenance()

	var address string
	if config.LocalhostOnly {
		address = fmt.Sprintf("%s:%d", "127.0.0.1", config.Port)
	} else {
		address = fmt.Sprintf("%s:%d", "0.0.0.0", config.Port)
	}

	if config.TLS {
		//const certFile = "./sslcert/cert.crt"
		//const keyFile = "./sslcert/key.key"
		//
		//log.Info("Listening on https://%s", address)
		//if err := http.ListenAndServeTLS(address, certFile, keyFile, nil); err != nil {
		//	log.FatalError(err.Error(), "Error starting TLS server")
		//}
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("prototype585.ddns.net"),
			Cache:      autocert.DirCache("certs"),
		}

		server := &http.Server{
			Addr: ":https",
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
				MinVersion:     tls.VersionTLS13,
			},
		}

		//if err := http.ListenAndServeTLS(address, certFile, keyFile, nil); err != nil {
		//	log.FatalError(err.Error(), "Error starting TLS server")
		//}

		go func() {
			err := http.ListenAndServe(":http", certManager.HTTPHandler(nil))
			if err != nil {
				log.FatalError(err.Error(), "Error serving HTTPS server")
				return
			}
		}()

		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.FatalError(err.Error(), "Error starting TLS server")
		}

	} else {
		log.Info("Listening on http://%s", address)
		if err := http.ListenAndServe(address, nil); err != nil {
			log.FatalError(err.Error(), "Error starting non-TLS server")
		}
	}
}

func maintenance() {
	time.Sleep(1 * time.Second)
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	task := func() {
		startMaintetance := time.Now().UnixMilli()
		token.DeleteExpiredTokens()
		finished := time.Now().UnixMilli() - startMaintetance
		log.Info("Maintenance finished in %d ms or %d seconds", finished, finished/1000)
	}

	task()

	for {
		select {
		case <-ticker.C:
			task()
		}
	}
}
