package main

import (
	"fmt"
	"net"
	"strings"
)

// Added comment
var our_port = "20005"

var localIP string

var master_address = "172.23.70.94"

var slave_IP = "10.100.23.15"

var slave_port = "33546"

var slave_address = slave_IP + ":" + slave_port

func main() {
	listen()
}

func listen() {
	l, err := net.Listen("tcp", ":" + slave_port)
	if err != nil{
		fmt.Printf("Failed to listen message %v\n", err)
	}
	defer l.Close()


	conn, err := l.Accept()
	fmt.Printf("\n accept: %t", conn)
	if err != nil{
		fmt.Printf("Failed to accept message %v\n", err)
	}

	buffer := make([]byte, 2048)

	//Reading from server
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("Failed to read message %v\n", err)
	}
	fmt.Printf("Message: %v", string(buffer[:n]))
	return conn
}