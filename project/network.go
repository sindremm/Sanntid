package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"

	//"strings"
	"elevator/network/bcast"
	"elevator/network/localip"
	"elevator/network/peers"
	"elevator/structs"
	"time"
)



var slave_IP = "10.100.23.15"

var slave_port = "10005"

var slave_address = slave_IP + ":" + slave_port

// Comment out sections: ctrl, k c
func main() {

	//Gets local IP
	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Printf("Local IP error: %v \n", err)
	}
	fmt.Printf("\nIP:", localIP)

	//The id that gets broadcasted to peers
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	//Ports for checking for life
	broadcast_port := 33344
	peers_port := 33224

	fmt.Printf("\n %v %v \n", peers_port, broadcast_port)

	tempMessage := structs.testTCPMsg{"Hello", 544}

	//updateLife(id, peers_port, broadcast_port)
	//checkForLife(id, peers_port, broadcast_port)
	communicate(localIP, tempMessage)
}

// Gets information on life status on the peers of the network
func checkForLife(id string, peers_port int, broadcast_port int) {

	peers_update_channel := make(chan peers.PeerUpdate)

	go peers.Receiver(peers_port, peers_update_channel)

	aliveCheck := make(chan structs.AliveMsg)

	go bcast.Receiver(broadcast_port, aliveCheck)

	//Prints peer update
	for {
		select {
		case p := <-peers_update_channel:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
		case a := <-aliveCheck:
			fmt.Printf("Received %#v \n", a)
		}
	}
}

// Sends message on life status
func updateLife(id string, peers_port int, broadcast_port int) {
	peer_bool := make(chan bool)
	go peers.Transmitter(peers_port, id, peer_bool)

	aliveUpdateMsg := make(chan structs.AliveMsg)

	go bcast.Transmitter(broadcast_port, aliveUpdateMsg)

	//Uncomment if we want updates every second
	/*
		go func() {
			helloMsg := AliveMsg{"Alive from ", id, 0}
			for {
				helloMsg.Iter++
				aliveUpdateMsg <- helloMsg
				time.Sleep(1 * time.Second)
			}
		}()
	*/
}

// Asks slave_address to connect, then sends a message to slave_address, then reads from slave
func communicate(localIP string, tempMessage structs.testTCPMsg) {

	conn, err := net.DialTimeout("tcp", slave_address, structs.TCP_timeout)
	if err != nil {
		fmt.Printf("Some error 1 %v\n", err)
		return
	}
	//defer conn.Close()
	//Message we send to other client (REMEMBER \000)
	msg := tempMessage.SomeMessage + strconv.Itoa(tempMessage.TempOrder) + "\000"
	_, err = conn.Write([]byte(msg))
	if err != nil {
		fmt.Printf("Failed to send message %v\n", err)
		return
		//TODO: fix this err, returns infinitely many "Message: Failed to read message EOF"
	}

	for {
		buffer := make([]byte, 2048)
		//Reading from slave
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Failed to read message %v\n", err)
		}
		fmt.Printf("Message: %v\n", string(buffer[:n]))
	}
}
