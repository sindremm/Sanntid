package main

import (
	"fmt"
	"net"
	"time"
	// "bufio"
)

const SERVER_IP_ADDRESS = "10.100.23.129"

// const SERVER_IP_ADDRESS = "255.255.255.255"
var port = "20005"
var full_address = SERVER_IP_ADDRESS + ":" + port

func main() {
	fmt.Printf(full_address + "\n")

	go listen2()
	time.Sleep(3 * time.Second)
	send()
	for i := 0; i < 10; i ++{
		time.Sleep(10 * time.Second)
	}
	fmt.Printf("Finished\n")

}

func listen() {
	pc, err := net.ListenPacket("udp", ":"+port) //evt full_adress
	if err != nil {
		//panic(err)
		fmt.Printf("Some error 6 %v\n", err)
		return
	}

	defer pc.Close()

	buf := make([]byte, 1024)
	for {
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			fmt.Printf(("Some error 5 %v\n"), err)
			return
		}
		fmt.Printf("%s sent this: %s\n", addr, buf[:n])
	}
	/* n,addr,err := pc.ReadFrom(buf)
	if err != nil {
	  panic(err)
	} */

}

func listen2() {
	var p = make([]byte, 2048)

	ServerAddr, err := net.ResolveUDPAddr("udp", full_address)
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, ServerAddr)
	if err != nil {
		fmt.Printf("Some error 4 %v", err)
		return
	}
	
	// _, err = bufio.NewReader(conn).Read(p)

	_, err = conn.Read(p)
	if err == nil {
		fmt.Printf("The message: %s\n", p)
	} else {
		fmt.Printf("Some error 2 %v\n", err)

	}

	conn.Close()

}

func looping() {
	for i := 0; i < 10; i++ {
		time.Sleep(1000 * time.Millisecond)
		fmt.Println("Hi")
	}
}

func send() {
	fmt.Printf("Sending message...\n")
	conn, err := net.Dial("udp", full_address)
	if err != nil {
		fmt.Printf("Some error 1 %v", err)
		return
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "Hi, how are you doing?")
	// var msg = "Hello, how are you"
	// _, err = conn.Write([]byte(msg))
	if err != nil {
		fmt.Printf("error 3: %s", err)
	}
}
