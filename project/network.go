package main

import (
	"fmt"
	"net"
	//"strings"
	"time"
	"elevator/network/localip"
	"elevator/network/peers"
)

type OrderMessage struct {
	OrderFloor int
	ButtonType int
}

var TCP_timeout = 500 * time.Millisecond

var our_port = "33546"

//var master_IP = "172.23.70.94"

var localIP string

var slave_IP = "10.100.23.15"

var slave_port = "10005"

var slave_address = slave_IP + ":" + slave_port
//Comment out sections: ctrl, k c
func main() {
	//Prints local IP
	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Printf("Local IP error: %v \n", err)
	}
	fmt.Printf("\nIP:", localIP)

	
	// l, err := net.Listen("tcp", ":"+"33567")
	// if err != nil{
	// 	fmt.Printf("Failed to listen message %v\n", err)
	// }
	

	//connection := accept()
	//go handleConnection(connection)
	//doEvery(2000*time.Millisecond, connectToSlave)

	peerBool := make(chan bool)
	peers.Transmitter(33546, "Hello", peerBool)

	//connectToSlave(localIP)
	//accept(l)
	
}



func connectToSlave(localIP string) {
	conn, err := net.DialTimeout("tcp", slave_address, TCP_timeout)
		if err != nil {
			fmt.Printf("Some error 1 %v\n", err)
		 	return
		}
		defer conn.Close()
	//Message we send to other client (REMEMBER \000)
	msg := localIP + "\000"
	_, err = conn.Write([]byte(msg))
	if err != nil {
		fmt.Printf("Failed to send message %v\n", err)
		return
	}

}

func accept(l net.Listener){
	fmt.Printf("\n Got to accept \n")

	for{
	conn, err := l.Accept()
	fmt.Printf("\n accept: %t", conn)
	if err != nil{
		fmt.Printf("Failed to accept message %v\n", err)
	}
	fmt.Printf("\n Got past accept \n")
	buffer := make([]byte, 2048)

	//Reading from slave
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("Failed to read message %v\n", err)
	}
	fmt.Printf("Message: %v", string(buffer[:n]))

	}
}

//For testing, delete later
func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}