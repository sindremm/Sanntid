package main

import (
	"fmt"
	"net"
	"strings"
)


var our_port = "20005"

var localIP string

var master_address = "172.23.70.94"

var slave_IP = "10.100.23.15"

var slave_port = "33546"

var slave_address = slave_IP + ":" + slave_port

func main() {
	//localIP := localIPfunc()
	//fmt.Printf("\nIP:", localIP)
	//connection := accept()
	//go handleConnection(connection)
	connectToSlave()
	accept()

}


func handleConnection(conn net.Conn) {
	fmt.Printf("\nRemote adress: %s", conn.RemoteAddr().String())
}

func localIPfunc() string{
	if localIP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
		//fmt.Printf("\n remote address: %s", conn.RemoteAddr().String())
	}
	return localIP
}

func connectToSlave() {
	conn, err := net.Dial("tcp", slave_address)
		if err != nil {
			fmt.Printf("Some error 1 %v\n", err)
		 	return
		}
		defer conn.Close()

}

func accept() net.Conn{
	l, err := net.Listen("tcp", ":"+our_port)
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