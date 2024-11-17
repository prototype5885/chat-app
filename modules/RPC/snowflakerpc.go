package RPC

import (
	"encoding/binary"
	"fmt"
	"net"
)

var SnowflakeRPC net.Conn

func ConnectSnowflakeRPC() {
	// snowflake rpc
	var err error
	SnowflakeRPC, err = net.Dial("tcp", "localhost:4000")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	// defer SnowflakeRPC.Close()
}

func RPCGenSnowflake() uint64 {
	_, err := SnowflakeRPC.Write([]byte{0})
	if err != nil {
		fmt.Println("Error sending message:", err)
		return 0
	}

	msg := make([]byte, 1024)
	_, err = SnowflakeRPC.Read(msg)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return 0
	}
	fmt.Println("Received snowflake")
	return binary.BigEndian.Uint64(msg)
}
