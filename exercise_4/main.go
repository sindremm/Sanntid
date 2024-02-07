package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"time"
)

const SERVER_IP_ADDRESS = "127.0.0.1"

var port = "20005"
var full_address = SERVER_IP_ADDRESS + ":" + port

func main() {
	start_time := time.Now()
	delay_milli := 100
	print_counter := time.Now()
	counter := 0

	filename := "/home/student/Documents/AjananMiaSindre/Sanntid/exercise_4/main.go"

	for {
		if start_time.Add(time.Duration(delay_milli) * time.Millisecond).Before(time.Now()) {
			break
		}
		counter = listen(&start_time, counter)
	}

	exec.Command("gnome-terminal", "--", "go", "run", filename).Run()

	for {
		send(counter)
		if print_counter.Add(time.Duration(1) * time.Second).Before(time.Now()) {
			counter += 1
			fmt.Printf("%d\n", counter)
			print_counter = time.Now()
		}

	}

}

func send(counter int) {

	// fmt.Printf("Sending message...\n")
	conn, err := net.Dial("udp", full_address)
	if err != nil {
		fmt.Printf("Some error 1 %v \n", err)
		return
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "%d", counter)
	// var msg = "Hello, how are you"
	// _, err = conn.Write([]byte(msg))
	if err != nil {
		fmt.Printf("error 3: %s \n", err)
	}

}

func listen(start_time *time.Time, counter int) (int){
	p := make([]byte, 2048)

	ServerAddr, err := net.ResolveUDPAddr("udp", full_address)
	if err != nil {
		fmt.Printf("Some error %v", err)
		return -1
	}

	conn, err := net.ListenUDP("udp", ServerAddr)
	if err != nil {
		fmt.Printf("Some error 1a %v", err)
		return -1
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	n, _, err := conn.ReadFromUDP(p)

	if err == nil {
		// err = binary.Read(bytes.NewReader(p[:n]), binary.BigEndian, &counter)
		counter, err = strconv.Atoi(string(p[:n]))
		*start_time = time.Now()
	} 
	// else {
	// 	fmt.Printf("Some error 2 %v\n", err)
	// }
	return counter
}
