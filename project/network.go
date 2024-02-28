package main

import (
	"fmt"
	"net"
	"strings"
	"time"
	//"Network-go/network/localip"
)


var our_port = "33546"

var localIP string

var master_IP = "172.23.70.94"


var slave_IP = "10.100.23.15"

var slave_port = "33546"

var slave_address = slave_IP + ":" + slave_port
//Comment out sections: ctrl, k c
func main() {
	//localIP, err := localip.LocalIP()
	//if err != nil {
	//	fmt.Printf("Local IP error: %v \n", err)
	//}
	//fmt.Printf("\nIP:", localIP)
	//connection := accept()
	//go handleConnection(connection)
	//doEvery(2000*time.Millisecond, connectToSlave)
	connectToSlave()
	accept()
	//handleConnection()

}

//sudo iptables -A INPUT -p tcp --dport 20005 -m statistic --mode random --probability 0.4 -j DROP
func handleConnection(conn net.Conn) {
	fmt.Printf("\nRemote adress: %s", conn.RemoteAddr().String())
}

func connectToSlave() {
	conn, err := net.Dial("tcp", slave_address)
		if err != nil {
			fmt.Printf("Some error 1 %v\n", err)
		 	return
		}
		defer conn.Close()
	//Message we send to other client (REMEMBER \000)
	msg := master_IP + ":" + our_port +"\000"
	_, err = conn.Write([]byte(msg))
	if err != nil {
		fmt.Printf("Failed to send message %v\n", err)
		return
	}

}

func accept(){
	l, err := net.Listen("tcp", ":"+our_port)
	if err != nil{
		fmt.Printf("Failed to listen message %v\n", err)
	}
	defer l.Close()

	for{
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

	}
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

func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}