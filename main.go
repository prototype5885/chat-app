package main

import (
	"log"
	"net/http"
	"proto-chat/modules/snowflake"
	"strconv"
)

func main() {
	log.Println("Starting server...")

	config := readConfigFile()

	setupLogging(config.LogInFile)

	snowflake.SetSnowflakeServerID(0)

	// database
	if config.Sqlite {
		ConnectSqlite()
	} else {
		ConnectMariadb(config.Username, config.Password, config.Address, strconv.Itoa(int(config.Port)), config.DatabaseName)
	}

	//test()

	// websocket
	var hub *Hub
	hub = newHub()
	go hub.run()
	http.HandleFunc("/wss", func(w http.ResponseWriter, r *http.Request) {
		wssHandler(hub, w, r)
	})

	// http.HandleFunc("GET /wss", wssHandler)
	http.HandleFunc("GET /login-register.html", loginRegisterHandler)
	http.HandleFunc("GET /chat.html", chatHandler)

	http.HandleFunc("POST /login", postRequestHandler)
	http.HandleFunc("POST /register", postRequestHandler)

	http.HandleFunc("/", mainHandler)

	const certFile = "./sslcert/selfsigned.crt"
	const keyFile = "./sslcert/selfsigned.key"

	// var certFile string = "/etc/letsencrypt/live/prototype585.asuscomm.com/fullchain.pem"
	// var keyFile string = "/etc/letsencrypt/live/prototype585.asuscomm.com/privkey.pem"

	log.Println("Listening on port 3000")
	if err := http.ListenAndServeTLS(":3000", certFile, keyFile, nil); err != nil {
		log.Fatal("Error starting server:", err)
	}
}

// func test() {
// 	start := time.Now().UnixMilli()
// 	for i := 0; i < 16; i++ {
// 		passwordSHA512, _ := base64.StdEncoding.DecodeString("XD")
// 		registerUser("testuser1", passwordSHA512)
// 		loginUser("testuser1", passwordSHA512)
// 	}
// 	log.Println("Time elapsed: ", time.Now().UnixMilli()-start)
// 	// registerUser("testuser1", "XD")
// 	// loginUser("testuser1", "XD")

// 	// var sha256bytes []byte = []byte("7a77dec6ccea995eddff86fad0769b88ab10898b2893887ceb47e7d8f1db8d38")
// 	// log.Println(len(sha256bytes))

// 	//var hash hash.Hash = sha512.New()
// 	//hash.Write([]byte("XD"))
// 	//var bytes []byte = hash.Sum(nil)
// 	//log.Println(len(bytes))
// 	//log.Println(base64.StdEncoding.EncodeToString(bytes))
// 	//
// 	//bytes2, _ := base64.StdEncoding.DecodeString("enfexszqmV7d/4b60HabiKsQiYsok4h860fn2PHbjTg=")
// 	//log.Println(len(bytes2))
// 	//
// 	//var endTime int64
// 	//
// 	//startTime := time.Now().UnixMilli()
// 	//
// 	//result := updateUserFieldDB("testuser1", "userid", uint64(9999999))
// 	//printWithName("testuser1", result.Message)
// 	//if !result.Success {
// 	//
// 	//}
// 	//endTime = time.Now().UnixMilli() - startTime
// 	//log.Println("updateUserField took:", endTime, "ms")

// 	// var num int = 2147483647
// 	// println(num)

// 	// var num2 uint = 18446744073709551615
// 	// var num = time.Now().Add(1844674407370955165).UnixMilli()
// 	// var snwflk = snowflake.Generate(0)
// 	// snowflake.Print(snwflk)

// 	// var ts int64 = 9223372036854775807
// 	// log.Println(ts + 10)
// 	// log.Println(uint64(ts + 10))
// 	// log.Println(uint64(0))

// 	// for i := 0; i < 8192*4; i++ {
// 	// 	getUserFieldFromDB("testuser1", "password_hash")
// 	// }
// 	// endTime = time.Now().UnixMilli() - startTime
// 	// log.Println("getUserField took:", endTime, "ms")

// 	// Success, user := GetUserRowStruct("testuser1")
// 	// if !Success {

// 	// }
// 	// log.Println(user.username)

// 	// var snflk = snowflake.Generate(0)
// 	// snowflake.Print(snflk)

// 	//type ClientChatMsg struct {
// 	//	ChannelId uint64
// 	//	ChatMsg   string
// 	//}
// 	//msg := ClientChatMsg{
// 	//	ChannelId: 0,
// 	//	ChatMsg:   "Hello World!",
// 	//}

// 	//msgJson, _ := json.Marshal(msg)

// 	//var start = time.Now().UnixMilli()
// 	//
// 	//var wg sync.WaitGroup
// 	//for i := 0; i < 128; i++ {
// 	//	wg.Add(1)
// 	//	go func() {
// 	//		defer wg.Done()
// 	//		passwordSHA512, _ := base64.StdEncoding.DecodeString("XD")
// 	//		//registerUser(strconv.Itoa(i), passwordSHA512)
// 	//		loginUser(strconv.Itoa(i), passwordSHA512)
// 	//		//addChatMessage([]byte(msgJson))
// 	//	}()
// 	//}
// 	//wg.Wait()

// 	//fmt.Println("All goroutines have finished in", time.Now().UnixMilli()-start)
// 	os.Exit(1)
// }
