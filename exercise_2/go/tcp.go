package main

import (
	"fmt"
	"net"
	// "io"
	//"time"
	// "bufio"
)

const TCP_SERVER_IP_ADDRESS = "10.100.23.129"

// const SERVER_IP_ADDRESS = "255.255.255.255"
var tcp_port = "33546"
var tcp_full_address = TCP_SERVER_IP_ADDRESS + ":" + tcp_port
var our_port = "20005"
var our_IP = "10.100.23.15"

func main() {
	connectToTCP()
	accept()
}

func readFromTCP() {
	buffer := make([]byte, 2048)
	conn, err := net.Dial("tcp", tcp_full_address)
	if err != nil {
		fmt.Printf("Some error 1 %v\n", err)
		 return
	}
	defer conn.Close()

	//Reading from server
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("Failed to read message %v\n", err)
		return
	}
	
	fmt.Printf("Message: %v", string(buffer[:n]))
	
}

func connectToTCP() {
		//Connecting to server
		conn, err := net.Dial("tcp", tcp_full_address)
		if err != nil {
			fmt.Printf("Some error 1 %v\n", err)
		 	return
		}
		defer conn.Close()

		//Messaging to server
		msg := "Connect to: " + our_IP + ":" + our_port +"\000"
		_, err = conn.Write([]byte(msg))
		if err != nil {
			fmt.Printf("Failed to send message %v\n", err)
			return
		}

}

func accept() {
	l, err := net.Listen("tcp", ":"+our_port)
	if err != nil{
		fmt.Printf("Failed to listen message %v\n", err)
		return
	}
	defer l.Close()


	conn, err := l.Accept()
	if err != nil{
		fmt.Printf("Failed to accept message %v\n", err)
		return
	}

	buffer := make([]byte, 2048)

	//Reading from server
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("Failed to read message %v\n", err)
		return
	}
	
	fmt.Printf("Message: %v", string(buffer[:n]))
	// go func(c net.Conn) {
	// 	io.Copy(c, c)

	// 	c.Close()
	// }(conn)

}