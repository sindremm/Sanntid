package main

import (
	"flag"
	"fmt"
	"net"

	//"strings"
	"elevator/network/bcast"
	"elevator/network/localip"
	"elevator/network/peers"
	"time"
)

type OrderMessage struct {
	OrderFloor int
	ButtonType int
}

type AliveMsg struct {
	Message string
	address string
	Iter    int
}

var TCP_timeout = 500 * time.Millisecond

var our_port = "33546"

var broadcast_port = 33344

var peers_port = 33224

var localIP string

var slave_IP = "10.100.23.15"

var slave_port = "10005"

var slave_address = slave_IP + ":" + slave_port

// Comment out sections: ctrl, k c
func main() {

	//Prints local IP
	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Printf("Local IP error: %v \n", err)
	}
	fmt.Printf("\nIP:", localIP)

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	updateLife(id)
	checkForLife(id)
	//communicate(localIP)
}

func checkForLife(id string) {

	peers_update_channel := make(chan peers.PeerUpdate)

	go peers.Receiver(peers_port, peers_update_channel)

	aliveCheck := make(chan AliveMsg)

	go bcast.Receiver(broadcast_port, aliveCheck)

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

func updateLife(id string) {
	peer_bool := make(chan bool)
	go peers.Transmitter(peers_port, id, peer_bool)

	aliveUpdateMsg := make(chan AliveMsg)

	go bcast.Transmitter(broadcast_port, aliveUpdateMsg)
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

func communicate(localIP string) {

	conn, err := net.DialTimeout("tcp", slave_address, TCP_timeout)
	if err != nil {
		fmt.Printf("Some error 1 %v\n", err)
		return
	}
	//defer conn.Close()
	//Message we send to other client (REMEMBER \000)
	msg := localIP + "\000"
	_, err = conn.Write([]byte(msg))
	if err != nil {
		fmt.Printf("Failed to send message %v\n", err)
		return
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
	//connectToSlave(localIP)
	//accept(l)

}

// Not currently used, delete later
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

// Not currently use, delete later
func accept(l net.Listener) {

	l, err := net.Listen("tcp", ":"+"33567")
	fmt.Printf("\n %s", l)
	if err != nil {
		fmt.Printf("Failed to listen message %v\n", err)
	}

	fmt.Printf("\n Got to accept \n")
	defer l.Close()
	for {
		conn, err := l.Accept()
		fmt.Printf("\n accept: %t", conn)
		if err != nil {
			fmt.Printf("Failed to accept message %v\n", err)
		}
		defer conn.Close()
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

// For testing, delete later
func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}
