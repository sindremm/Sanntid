package master_slave

import (
	"fmt"
	// "net"
	// "os/exec"
	"strconv"
	"strings"
	"time"

	"Driver-go/elevio"

	scheduler "elevator/elevator-scheduler"
	single "elevator/single-elevator"
	"elevator/structs"
	tcp_interface "elevator/tcp-interface"

	"elevator/network/bcast"
	"elevator/network/localip"
	"elevator/network/peers"
)

type MasterSlave struct {
	CURRENT_DATA  *structs.SystemData
	UNIT_ID       int
	ELEVATOR_UNIT *single.Elevator
	IP_ADDRESS    string
	LISTEN_PORT   string
}

// Create a MasterSlave
func MakeMasterSlave(UnitID int, port string, elevator single.Elevator) *MasterSlave {
	MS := new(MasterSlave)

	// Initialize current data
	SD := structs.SystemData{
		MASTER_ID:         0,
		UP_BUTTON_ARRAY:   &([structs.N_FLOORS]bool{}),
		DOWN_BUTTON_ARRAY: &([structs.N_FLOORS]bool{}),
		ELEVATOR_DATA:     &([structs.N_ELEVATORS]structs.ElevatorData{}),
		COUNTER:           0,
	}

	// Set data
	MS.CURRENT_DATA = &SD

	// Set identifying ID of unit
	MS.UNIT_ID = UnitID

	// Set corresponding elevator
	MS.ELEVATOR_UNIT = &elevator

	//IP

	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Printf("Error with localIP \n")
	}
	MS.IP_ADDRESS = localIP

	// Set the port where tcp messages are received
	MS.LISTEN_PORT = port

	// Start threads
	go elevator.Main()

	return MS
}

func (ms *MasterSlave) MainLoop() {
	// Heartbeat
	peers_port := 33224
	broadcast_port := 32244
	input_id := strconv.Itoa(ms.UNIT_ID) + "-" + ms.IP_ADDRESS + ms.LISTEN_PORT
	Heartbeat(input_id, peers_port, broadcast_port)

	go CheckHeartbeat(ms, peers_port, broadcast_port)

	// Check if this elevator is Master
	is_master := ms.CURRENT_DATA.MASTER_ID == ms.UNIT_ID

	// Start listening to received data
	received_data_channel := make(chan []byte)
	own_address := ms.IP_ADDRESS + ms.LISTEN_PORT

	// Start reading received data
	go tcp_interface.ReceiveData(own_address, received_data_channel)

	// Main loop of Master-slave
	for {
		// fmt.Printf("%s", structs.SystemData_to_string(*ms.CURRENT_DATA))
		if is_master {

			// Run if current elevator is master

			// TODO: Update SystemData:

			// Get all data from channel and insert into SystemData

			loop:
			for {
				select {
				case data := <-received_data_channel:
					decoded_data := tcp_interface.DecodeMessage(data)
					id := decoded_data.Sender_id
					ms.CURRENT_DATA.ELEVATOR_DATA[id] = decoded_data.Data.ELEVATOR_DATA[id]
					// fmt.Printf("data: %s", structs.SystemData_to_string(decoded_data.Data))
				default:
					break loop
				}
			}
			
			// Update calls, buttons
			// Update the states of each elevator

			// Increase counter
			ms.CURRENT_DATA.COUNTER += 1

			// UpdateElevatorTargets() (Only run when new calls, or update in state of elevator)
			ms.UpdateElevatorTargets()

			// Send updated SystemData
			ms.BroadcastSystemData()
			// fmt.Printf("%s", structs.SystemData_to_string(*ms.CURRENT_DATA))
			
		} else {
			// Run if current elevator is slave

			// Receive data from master
			received_data := <-received_data_channel
			decoded_data := tcp_interface.DecodeMessage(received_data)

			// Check if the received data is newer then current data, and update current data if so
			if decoded_data.Data.COUNTER > ms.CURRENT_DATA.COUNTER {
				ms.CURRENT_DATA = &decoded_data.Data
			}
		}

		calls := ms.CURRENT_DATA.ELEVATOR_DATA[ms.UNIT_ID].ELEVATOR_TARGETS
		ms.ELEVATOR_UNIT.PickTarget(calls)

		time.Sleep(5*time.Second)
	}
}

func (ms *MasterSlave) BroadcastSystemData() {
	// Send system data to each elevator
	var encoded_system_data []byte
	for i := 0; i < structs.N_ELEVATORS; i++ {
		// Find corresponding address of elevator client
		client_address := ms.CURRENT_DATA.ELEVATOR_DATA[i].ADDRESS
		if client_address == "" {
			continue
		}
		// Send system data to client
		send_message := structs.TCPMsg{
			Sender_id: ms.UNIT_ID,
			Data:      *ms.CURRENT_DATA,
		}
		encoded_system_data = tcp_interface.EncodeMessage(&send_message)
		tcp_interface.SendData(client_address, encoded_system_data)
	}

}

// Read from the channels and put data into variables
func (ms *MasterSlave) ReadButtons(button_order chan elevio.ButtonEvent) {
	for {
		select {
		case bo := <-button_order:
			// Transform order to readable format
			floor, btn := ms.InterpretOrder(bo)
			// Add order to internal array and set lights
			ms.AddOrderToSystemDAta(floor, btn)
			elevio.SetButtonLamp(btn, floor, true)
		}
	}
}

// Convert order to readable format
func (ms *MasterSlave) InterpretOrder(button_order elevio.ButtonEvent) (floor int, button elevio.ButtonType) {
	order_floor := button_order.Floor
	order_button := button_order.Button

	return order_floor, order_button
}

// Add order to system data
func (ms *MasterSlave) AddOrderToSystemDAta(floor int, button elevio.ButtonType) {
	switch button {
	case 0:
		ms.CURRENT_DATA.UP_BUTTON_ARRAY[floor] = true
		elevio.SetButtonLamp(elevio.BT_HallUp, floor, true)
	case 1:
		ms.CURRENT_DATA.DOWN_BUTTON_ARRAY[floor] = true
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, true)
	case 2:
		ms.CURRENT_DATA.ELEVATOR_DATA[ms.UNIT_ID].INTERNAL_BUTTON_ARRAY[floor] = true
		elevio.SetButtonLamp(elevio.BT_Cab, floor, true)
	}
}

func (ms *MasterSlave) UpdateElevatorTargets() {
	// Get new elevator targets
	movement_map := *scheduler.CalculateElevatorMovement(*(ms.CURRENT_DATA))

	// Map to convert from map of elevators to array of elevators
	key_to_int_map := map[string]int{
		"one":   0,
		"two":   1,
		"three": 2,
	}

	// Update values in ELEVATOR_TARGETS of SystemData
	for k := range movement_map {
		(*ms.CURRENT_DATA.ELEVATOR_DATA)[key_to_int_map[k]].ELEVATOR_TARGETS = movement_map[k]
	}
}

func (ms *MasterSlave) RequestSlaveData(data_string string) {

}

// // HandleOrderFromMaster is a method on the MasterSlave struct that processes an order from the master.
// func (ms *MasterSlave) HandleOrderFromMaster(order *structs.ElevatorState) error {
// 	// Check if the target floor in the order is valid (between 0 and 3)
// 	if order.TARGET_FLOOR < 0 || order.TARGET_FLOOR > structs.N_FLOORS {
// 		return fmt.Errorf("Invalid order: floor must be between 0 and 3")
// 	}
// 	// Check if the direction in the order is valid (0 for stop, 1 for up, 2 for down)
// 	if order.DIRECTION < 0 || order.DIRECTION > 2 {
// 		return fmt.Errorf("Invalid order: direction must be 0, 1 or 2")
// 	}

// 	// Update the SystemData based on the order
// 	// If the direction is 1 (up), set the corresponding floor in the up button array to true
// 	if order.DIRECTION == 1 {
// 		ms.CURRENT_DATA.UP_BUTTON_ARRAY[order.TARGET_FLOOR] = true
// 	// If the direction is 2 (down), set the corresponding floor in the down button array to true
// 	} else if order.DIRECTION == 2 {
// 		ms.CURRENT_DATA.DOWN_BUTTON_ARRAY[order.TARGET_FLOOR] = true
// 	// If the direction is 0 (stop), do nothing
// 	} else {
// 		// TODO: Set internal orders for given elevator
// 		// ms.current_data.INTERNAL_BUTTON_ARRAY[order.TARGET_FLOOR] = true
// 	}
// 	// Print a message indicating that the order has been processed
// 	fmt.Printf("Order for floor %d with direction %d has been processed.\n", order.TARGET_FLOOR, order.DIRECTION)
// 	return nil
// }

// func (ms *structs.SystemData) SwitchToBackup() {
// 	ms.SENDER = 0
// 	fmt.Println("Master is dead, switching to backup")
// }

var fullAddress = structs.SERVER_IP_ADDRESS + ":" + structs.PORT

// func StartMasterSlave(leader *MasterSlave) {
// 	//Set the leader as the Master
// 	leader.Master = true

// 	// start_time := time.Now()
// 	print_counter := time.Now()
// 	counter := 0

// 	ms := &MasterSlave{}

// 	filename := "/home/student/Documents/AjananMiaSindre/Sanntid/exercise_4/main.go"

// 	listener, err := net.Listen("tcp", fullAddress)
// 	if err != nil {
// 		fmt.Printf("Error creating TCP listener: %v\n", err)
// 		return
// 	}
// 	defer listener.Close()

// 	exec.Command("gnome-terminal", "--", "go", "run", filename).Run()

// 	for {
// 		conn, err := listener.Accept()
// 		if err != nil {
// 			fmt.Printf("Error accepting TCP connection: %v\n", err)
// 			continue
// 		}

// 		// Handle incoming TCP connection
// 		go func(conn net.Conn) {
// 			defer conn.Close()

// 			// Receive SystemData from the master
// 			data := &structs.SystemData{}
// 			if err := network.receiveSystemData(conn, data); err != nil {
// 				fmt.Printf("Error receiving SystemData: %v\n", err)
// 				return
// 			}

// 			// Process received SystemData
// 			// (Add your logic here based on the received data)

// 		}(conn)

// 		// Send SystemData to the master periodically
// 		if time.Since(print_counter).Seconds() > 1 {
// 			counter++
// 			ms.CURRENT_DATA.COUNTER = counter
// 			if err := network.sendSystemData(conn, ms.CURRENT_DATA); err != nil {
// 				fmt.Printf("Error sending SystemData: %v\n", err)
// 			}

// 			fmt.Printf("%d\n", counter)
// 			print_counter = time.Now()
// 		}
// 	}
// }

// // sendSystemData is a function that sends SystemData over a TCP connection.
// // It takes a net.Conn object representing the connection and a pointer to the SystemData object to be sent.
// // It returns an error if any occurs during the process.
// func sendSystemData(conn net.Conn, data *structs.SystemData) error {
// 	// Create a new encoder that will write to conn
// 	encoder := gob.NewEncoder(conn)
// 	// Encode the SystemData object and send it over the connection
// 	// If an error occurs during encoding, wrap it in a new error indicating that encoding failed
// 	if err := encoder.Encode(data); err != nil {
// 		return fmt.Errorf("failed to encode SystemData: %v", err)
// 	}
// 	// If no error occurred, return nil
// 	return nil
// }

// // receiveSystemData is a function that receives SystemData over a TCP connection.
// // It takes a net.Conn object representing the connection and a pointer to the SystemData object where the received data will be stored.
// // It returns an error if any occurs during the process.
// func receiveSystemData(conn net.Conn, data *structs.SystemData) error {
// 	// Create a new decoder that will read from conn
// 	decoder := gob.NewDecoder(conn)
// 	// Decode the received data and store it in the SystemData object
// 	// If an error occurs during decoding, wrap it in a new error indicating that decoding failed
// 	if err := decoder.Decode(data); err != nil {
// 		return fmt.Errorf("failed to decode SystemData: %v", err)
// 	}
// 	// If no error occurred, return nil
// 	return nil
// }

// //TODO: Implement function to check if the new targets differ from the current ones
// func (ms *MasterSlave) CheckIfReceivedNewTargets() {
// 	ms.CURRENT_DATA.COUNTER
// }

// Heartbeat sends a heartbeat message to all other elevators.
func Heartbeat(id string, peers_port int, broadcast_port int) {

	peer_bool := make(chan bool)
	go peers.Transmitter(peers_port, id, peer_bool)

	aliveUpdateMsg := make(chan structs.AliveMsg)

	go bcast.Transmitter(broadcast_port, aliveUpdateMsg)
}

// CheckHeartbeat checks if a heartbeat has been received from the leader.
func CheckHeartbeat(ms *MasterSlave, peers_port int, broadcast_port int) {
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
			if p.New != "" {
				UpdateElevatorMap(ms, p.New)
			}
			if p.Lost != nil {
				UpdateLostConnection(ms, p.Lost)
			}
		case a := <-aliveCheck:
			fmt.Printf("Received %#v \n", a)
		}
	}
}

// Chenges alive status and adds address when a peer connects
func UpdateElevatorMap(ms *MasterSlave, newElevatorID string) {

	elevatorNum, elevatorAddress := splitPeerString(newElevatorID)
	ms.CURRENT_DATA.ELEVATOR_DATA[elevatorNum].ADDRESS = elevatorAddress
	ms.CURRENT_DATA.ELEVATOR_DATA[elevatorNum].ALIVE = true
}

//Chenges alive status when a peer disconnects
func UpdateLostConnection(ms *MasterSlave, lostElevatorID []string) {
	for i := range lostElevatorID {
		elevatorNum, _ := splitPeerString(lostElevatorID[i])
		ms.CURRENT_DATA.ELEVATOR_DATA[elevatorNum].ALIVE = false
	}

}

//Splits pper string to the unit ID and address
func splitPeerString(peerString string) (elevatorNum int, elevatoraddress string){
	splitString := strings.Split(peerString, "-")
	elevatorNum, err := strconv.Atoi(splitString[0])
	if err!=nil {
		fmt.Printf("Error with string splitting: %v \n", err)
	}
	elevatorAddress := splitString[1]
	return elevatorNum, elevatorAddress
}
