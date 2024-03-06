package main

import (
	"fmt"
	"net"

	//"strings"
	"elevator/network/localip"
)

// Added comment
var our_port = "10005"

func main() {
	communicateClient()
}

// Listens and accepts connection on our_port, then sends a message back
func communicateClient() {
	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Printf("\nIP: %v", localIP)
	}
	fmt.Printf("\nIP: %v", localIP)

	l, err := net.Listen("tcp", ":"+our_port)
	if err != nil {
		fmt.Printf("Failed to listen message %v\n", err)
	}

	for {
		conn, err := l.Accept()
		//fmt.Printf("\n accept: %t", conn)
		if err != nil {
			fmt.Printf("Failed to accept message %v\n", err)
		}
		defer conn.Close()
		buffer := make([]byte, 2048)

		//buffer[:n] is message from client
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Failed to read message %v\n", err)
		}
		received_address := string(buffer[:n])
		fmt.Printf("\n%v", received_address)
		//fmt.Printf("\n Remote address: %v", conn.RemoteAddr().String())

		msg_back := "Hello back\000"
		_, err = conn.Write([]byte(msg_back))
		if err != nil {
			fmt.Printf("Failed \n")
		}
	}

}
