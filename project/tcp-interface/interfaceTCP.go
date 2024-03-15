package interfaceTCP

import (
	"fmt"
	"log"
	"net"

	"elevator/structs"
	"encoding/json"
)

// Encodes systemData struct to []byte to be sent by TCP
func EncodeMessage(s *structs.TCPMsg) []byte {
	b, err := json.Marshal(s)
	if err != nil {
		fmt.Print("Error with Marshal \n")
	}
	return b
}

func EncodeSystemData(s *structs.SystemData) []byte {
	b, err := json.Marshal(s)
	if err != nil {
		fmt.Print("Error with Marshal \n")
	}
	return b
}

// Decode []byte sent with TCP into SystemData struct
func DecodeMessage(data []byte) *structs.TCPMsg {
	var received_message *structs.TCPMsg

	err := json.Unmarshal([]byte(data), &received_message)
	if err != nil {
		log.Fatalf("Error with decoding:  %s", err)
	}
	return received_message
}

// Decode []byte sent with TCP into SystemData struct
func DecodeSystemData(data []byte) *structs.SystemData {
	var received_message *structs.SystemData

	err := json.Unmarshal([]byte(data), &received_message)
	if err != nil {
		log.Fatalf("Error with decoding:  %s", err)
	}
	return received_message
}

func DecodeHallOrderMsg(data []byte) *structs.HallorderMsg {
	var received_message *structs.HallorderMsg

	err := json.Unmarshal([]byte(data), &received_message)
	if err != nil {
		log.Fatalf("Error with decoding:  %s", err)
	}
	return received_message
}

// Asks slave_address to connect, then sends a message to slave_address, then reads from slave
func SendData(client_address string, message []byte) {

	// Dial client to establish connection
	conn, err := net.DialTimeout("tcp", client_address, structs.TCP_timeout)
	if err != nil {
		fmt.Printf("Some error 1 %v\n", err)
		return
	}
	//TODO: ADD somwhere in the code
	// defer conn.Close()

	// Send data
	_, err = conn.Write(message)
	if err != nil {
		fmt.Printf("Failed to send message %v\n", err)
		return
		//TODO: fix this err, returns infinitely many "Message: Failed to read message EOF"
	}
}

// Listens and accepts connection on our_port, then sends a message back
func ReceiveSlaveData(listen_address string, message_channel chan structs.TCPMsg) {

	// Listen for connection on specified port
	l, err := net.Listen("tcp", listen_address)
	if err != nil {
		fmt.Printf("Failed to listen message %v\n", err)
	}

	// Runs for loop to wait for message
	for {
		fmt.Printf("Receiving data\n")
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
			read_message := buffer[:n]

			decoded_message := DecodeMessage(read_message)
			message_type := decoded_message.MessageType

			// fmt.Printf("Recieved message in reader. Type: %d\n", decoded_message.MessageType)
			switch message_type {
			case structs.MASTERMSG:

			default:
				fmt.Printf("Gotten slave data message\n")
				message_channel <- *decoded_message
			}
		}
	}
}

// Listens and accepts connection on our_port, then sends a message back
func ReceiveMasterData(listen_address string, message_channel chan structs.TCPMsg) {

	// Listen for connection on specified port
	l, err := net.Listen("tcp", listen_address)
	if err != nil {
		fmt.Printf("Failed to listen message %v\n", err)
	}

	// Runs for loop to wait for message
	for {
		fmt.Printf("Receiving data\n")
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
			read_message := buffer[:n]

			decoded_message := DecodeMessage(read_message)
			message_type := decoded_message.MessageType

			// fmt.Printf("Recieved message in reader. Type: %d\n", decoded_message.MessageType)
			switch message_type {
			case structs.MASTERMSG:
				fmt.Printf("GOTTEN master data message\n")
				message_channel <- *decoded_message
			default:
			}

		}
	}
}
