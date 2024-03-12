package interfaceTCP

import (
	"flag"
	"fmt"
	"net"
	"log"

	"strings"
	"elevator/network/localip"
	"elevator/structs"
	"encoding/json"
	"strconv"
	//"time"
)


// Added comment
// var our_port = "10005"

// var slave_IP = "172.23.70.94"

// var slave_port = "10005"

// var slave_address = slave_IP + ":" + slave_port



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

	//tempMessage := structs.TestTCPMsg{"Hello", 544}

	// updateLife(id, peers_port, broadcast_port)
	// checkForLife(id, peers_port, broadcast_port)
	//communicate(localIP, tempMessage)
}

//TODO: Delete when AJ is done refactoring
/*
// Gets information on life status on the peers of the local network
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


// Sends message on life status on local network
func updateLife(id string, peers_port int, broadcast_port int) {
	peer_bool := make(chan bool)
	go peers.Transmitter(peers_port, id, peer_bool)

	aliveUpdateMsg := make(chan structs.AliveMsg)

	go bcast.Transmitter(broadcast_port, aliveUpdateMsg)

	//Uncomment if we want updates every second
	
	// go func() {
	// 	helloMsg := AliveMsg{"Alive from ", id, 0}
	// 	for {
	// 		helloMsg.Iter++
	// 		aliveUpdateMsg <- helloMsg
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }()

}
*/

// Encodes systemData struct to []byte to be sent by TCP
func EncodeSystemData(s *structs.SystemData) ([]byte){
	b, err := json.Marshal(s)
	if err!= nil {
		fmt.Print("Error with Marshal \n")
	}
	return b
}

// Decode []byte sent with TCP into SystemData struct
func DecodeSystemData(data []byte) *structs.SystemData{
	var systemData *structs.SystemData

	err := json.Unmarshal([]byte(data), &systemData)
	if err != nil {
        log.Fatalf("Error with decoding:  %s", err)
    }
	return systemData
}





// Asks slave_address to connect, then sends a message to slave_address, then reads from slave
func SendSystemData(client_address string, TCPmessage *structs.SystemData) (){

	// Dial client to establish connection
	conn, err := net.DialTimeout("tcp", client_address, structs.TCP_timeout)
	if err != nil {
		//TODO: Use peers to check if alive, and remove if gone
		fmt.Printf("Some error 1 %v\n", err)
		return
	}
	//TODO: ADD somwhere in the code
	// defer conn.Close()

	//Encode systemdata and add zero termination
	msg := append(EncodeSystemData(TCPmessage), "\000"...)

	// Send data
	_, err = conn.Write(msg)
	if err != nil {
		fmt.Printf("Failed to send message %v\n", err)
		return
		//TODO: fix this err, returns infinitely many "Message: Failed to read message EOF"
	}
}




// Listens and accepts connection on our_port, then sends a message back
func ReceiveSystemData(listen_address string, write_data *structs.SystemData) (structs.SystemData) {

	// Listen for connection on specified port 
	l, err := net.Listen("tcp", listen_address)
	if err != nil {
		fmt.Printf("Failed to listen message %v\n", err)
	}

	// Runs for loop to wait for message
	for {
		// Accepts message if received
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
		} else {
			// Return received data
		write_data = DecodeSystemData(buffer[:n])
		}
	}
}


// The slave sends it's info to the master unit
func SlaveSendInfoToMaster(master_address string, slave_message *structs.SystemData) {
	//Encode systemdata and add zero termination

	SendSystemData(master_address, slave_message)
	
}

//Adds new address and id number to the ElevatorMap
func UpdateElevatorMap(SystemData *structs.SystemData, newElevatorID string) {

	splitString := strings.Split(newElevatorID, "-")
	elevatorNum, _ := strconv.Atoi(splitString[0])
	elevatorAddress := splitString[1]
	SystemData.ELEVATOR_DATA[elevatorNum].ADDRESS = elevatorAddress
}