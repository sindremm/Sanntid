package main

import (
	"fmt"
	"net"
	//"strings"
	"elevator/network/localip"
)

// Added comment
var our_port = "20005"

var localIP string

var master_address = "172.23.70.94"

var slave_IP = "10.100.23.15"

var slave_port = "33546"

var slave_address = slave_IP + ":" + slave_port

func main() {
	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Printf("\nIP: %v", localIP)	
	}
	fmt.Printf("\nIP: %v", localIP)
	listen()
}

func listen() {

	l, err := net.Listen("tcp", ":"+slave_port)
	if err != nil {
		fmt.Printf("Failed to listen message %v\n", err)
	}
	defer l.Close()
	//loop to contiouusly receive messages
	for {
		conn, err := l.Accept()
		//fmt.Printf("\n accept: %t", conn)
		if err != nil {
			fmt.Printf("Failed to accept message %v\n", err)
		}

		buffer := make([]byte, 2048)

		//buffer[:n] is message from client
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Failed to read message %v\n", err)
		}
		received_address := string(buffer[:n])
		fmt.Printf("\n%v", received_address)
		//fmt.Printf("\n Remote address: %v", conn.RemoteAddr().String())
		connectToMaster(received_address)
	}
}
func connectToMaster(received_address string) {
	conn, err := net.Dial("tcp", received_address)
	if err != nil {
		fmt.Printf("Error with sending %v \n", err)
	}
	defer conn.Close()
	msg_back := "Hello back\000"
	_, err = conn.Write([]byte(msg_back))
	if err != nil {
		fmt.Printf("Failed \n")
	}

}
