package master_slave

import (
	"fmt"
	//"time"
	// "net"
	// "os/exec"
	"strconv"
	"strings"

	// "time"

	"Driver-go/elevio"

	scheduler "elevator/elevator-scheduler"
	"elevator/structs"
	tcp_interface "elevator/tcp-interface"

	"elevator/network/bcast"
	"elevator/network/localip"
	"elevator/network/peers"
	//election "elevator/network/new-election"
)

type MasterSlave struct {
	CURRENT_DATA *structs.SystemData
	UNIT_ID      int
	IP_ADDRESS   string
	LISTEN_PORT  string
}

// Create a MasterSlave
func MakeMasterSlave(UnitID int, port string) *MasterSlave {
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
	// MS.ELEVATOR_UNIT = &elevator

	//IP

	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Printf("Error with localIP \n")
	}
	MS.IP_ADDRESS = localIP

	// Set the port where tcp messages are received
	MS.LISTEN_PORT = port

	return MS
}

func (ms *MasterSlave) StartMasterSlave() {
	// Heartbeat
	newMasterChoice(ms)
	fmt.Printf("Start of loop \n")
	peers_port := 33224
	broadcast_port := 32244
	input_id := strconv.Itoa(ms.UNIT_ID) + "-" + ms.IP_ADDRESS + ms.LISTEN_PORT
	Heartbeat(input_id, peers_port, broadcast_port)

	go CheckHeartbeat(ms, peers_port, broadcast_port)

	// Create slave and master message channels
	slave_messages_channel := make(chan structs.TCPMsg)
	master_messages_channel := make(chan structs.TCPMsg)

	input_address := ms.IP_ADDRESS + ms.LISTEN_PORT

	// Put data into slave and master channels
	go tcp_interface.ReceiveData(input_address, slave_messages_channel, master_messages_channel)

	go ms.MasterLoop(slave_messages_channel)
	go ms.SlaveLoop(master_messages_channel)

	fmt.Printf("\n%s\n", structs.SystemData_to_string(*ms.CURRENT_DATA))
	// Main loop of Master-slave

}

func (ms *MasterSlave) MasterLoop(slave_messages_channel chan structs.TCPMsg) {
	for {
		// Check if this elevator is Master

		is_master := ms.CURRENT_DATA.MASTER_ID == ms.UNIT_ID
		has_updated := false
		if is_master {
			fmt.Printf("Master: %v", ms.CURRENT_DATA.MASTER_ID)
			// Run if current elevator is master
			// TODO: Update SystemData:
		master_loop:
			for {
				select {
				case slave_data := <-slave_messages_channel:

					// Notify that an update has occured
					has_updated = true
					// fmt.Printf("Received Slave Data\n")

					//Decodes the data recieved from slave
					id := slave_data.Sender_id

					msg_type := slave_data.MessageType
					switch msg_type {
					case structs.NEWCABCALL:
						// Decode cab call
						decoded_cabCall := tcp_interface.DecodeHallOrderMsg(slave_data.Data)

						// Set corresponding cab order to true
						ms.CURRENT_DATA.ELEVATOR_DATA[id].INTERNAL_BUTTON_ARRAY[decoded_cabCall.Order_floor] = true

					case structs.NEWHALLORDER:
						// Decode hall
						decoded_hallOrderMessage := tcp_interface.DecodeHallOrderMsg(slave_data.Data)

						// Find corresponding floor of order
						clear_floor := decoded_hallOrderMessage.Order_floor

						// Sets the order in the right direction to true
						if decoded_hallOrderMessage.Order_direction[0] {
							ms.CURRENT_DATA.UP_BUTTON_ARRAY[clear_floor] = true
						}
						if decoded_hallOrderMessage.Order_direction[1] {
							ms.CURRENT_DATA.DOWN_BUTTON_ARRAY[clear_floor] = true
						}

					case structs.UPDATEELEVATOR:
						// Decode system data
						decoded_systemData := tcp_interface.DecodeSystemData(slave_data.Data)
						// fmt.Printf("Received Decoded data:\n")
						// fmt.Printf("%s", structs.SystemData_to_string(*decoded_systemData))


						//Insert data into SystemData
						// ms.CURRENT_DATA.ELEVATOR_DATA[id] = decoded_systemData.ELEVATOR_DATA[id]
						ms.CURRENT_DATA.ELEVATOR_DATA[id].CURRENT_FLOOR = decoded_systemData.ELEVATOR_DATA[id].CURRENT_FLOOR
						ms.CURRENT_DATA.ELEVATOR_DATA[id].DIRECTION = decoded_systemData.ELEVATOR_DATA[id].DIRECTION
						ms.CURRENT_DATA.ELEVATOR_DATA[id].INTERNAL_STATE = decoded_systemData.ELEVATOR_DATA[id].INTERNAL_STATE

					case structs.CLEARHALLORDER:
						//Clears The direction button and the internal button of the cleared floor
						hallOrderMsg := tcp_interface.DecodeHallOrderMsg(slave_data.Data)
						clear_floor := hallOrderMsg.Order_floor
						clear_direction := hallOrderMsg.Order_direction

						// Clear cab order
						ms.CURRENT_DATA.ELEVATOR_DATA[id].INTERNAL_BUTTON_ARRAY[clear_floor] = false

						// Check and clear up and down order
						if clear_direction[0] {
							//fmt.Printf("Clear in up in master\n")
							ms.CURRENT_DATA.UP_BUTTON_ARRAY[clear_floor] = false
						}
						if clear_direction[1] {
							//fmt.Printf("Clear in down in master\n")
							ms.CURRENT_DATA.DOWN_BUTTON_ARRAY[clear_floor] = false
						}
					default:
						fmt.Printf("Unrecognized type: %v\n", msg_type)
					}

				default:
					break master_loop
				}
			}

			// Check if there has been an update
			if has_updated {
				// fmt.Printf("\n%s\n", structs.SystemData_to_string(*ms.CURRENT_DATA))

				// Increase counter of data
				ms.CURRENT_DATA.COUNTER += 1

				// Only run when new calls, or update in state of elevator
				ms.UpdateElevatorTargets()

				// Send updated SystemData
				ms.BroadcastSystemData()
			}

		}
		//fmt.Printf("mjau mjau\n")

		// calls := ms.CURRENT_DATA.ELEVATOR_DATA[ms.UNIT_ID].ELEVATOR_TARGETS
		// ms.ELEVATOR_UNIT.PickTarget(calls)
	}
}

func (ms *MasterSlave) SlaveLoop(master_messages_channel chan structs.TCPMsg) {
	for {
	slave_loop:
		for {
			select {
			case master_data := <-master_messages_channel:
				// fmt.Printf("Received Master Data\n")

				// // Receive data from master
				// decoded_data := tcp_interface.DecodeMessage(master_data)
				decoded_systemData := tcp_interface.DecodeSystemData(master_data.Data)
				if decoded_systemData.MASTER_ID != ms.CURRENT_DATA.MASTER_ID {
					ms.CURRENT_DATA.MASTER_ID = decoded_systemData.MASTER_ID
				}

				// Find type of message
				master_data_type := master_data.MessageType

				// Check if message is from master
				switch master_data_type {
				case structs.MASTERMSG:

					// Check if the received data is newer then current data, and update current data if so
					if decoded_systemData.COUNTER > ms.CURRENT_DATA.COUNTER {
						ms.CURRENT_DATA = decoded_systemData
					}

					UpdateElevatorLights(ms)
				default:
					fmt.Printf("Unrecognized master message: %d\n", master_data_type)
				}

				// fmt.Printf("\n%s\n", structs.SystemData_to_string(*ms.CURRENT_DATA))

			default:
				break slave_loop
			}

		}
	}
}

func (ms *MasterSlave) BroadcastSystemData() {

	// Encode system data
	encoded_current_data := tcp_interface.EncodeSystemData(ms.CURRENT_DATA)

	// Construct broadcast message
	send_message := structs.TCPMsg{
		MessageType: structs.MASTERMSG,
		Sender_id:   ms.UNIT_ID,
		Data:        encoded_current_data,
	}

	// Encode into message
	encoded_system_data := tcp_interface.EncodeMessage(&send_message)

	// Send message into each elevator
	for i := 0; i < structs.N_ELEVATORS; i++ {
		// Find corresponding address of elevator client
		client_address := ms.CURRENT_DATA.ELEVATOR_DATA[i].ADDRESS
		if client_address == "" {
			continue
		}

		//TODO: Find out if the if statement is needed:
		//Send only data if the slave is alive
		if ms.CURRENT_DATA.ELEVATOR_DATA[i].ALIVE {
			tcp_interface.SendData(client_address, encoded_system_data)
		}
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
		//fmt.Printf("movement_map: %v \n", movement_map)
	}
}

func UpdateElevatorLights(ms *MasterSlave) {
	//fmt.Printf("Lamp set\n")
	for i := 0; i < structs.N_FLOORS; i++ {
		if !ms.CURRENT_DATA.UP_BUTTON_ARRAY[i] {
			elevio.SetButtonLamp(0, i, false)
		}
		if !ms.CURRENT_DATA.DOWN_BUTTON_ARRAY[i] {
			elevio.SetButtonLamp(1, i, false)
		}
		if !ms.CURRENT_DATA.ELEVATOR_DATA[ms.UNIT_ID].INTERNAL_BUTTON_ARRAY[i] {
			elevio.SetButtonLamp(2, i, false)
		}
		if ms.CURRENT_DATA.UP_BUTTON_ARRAY[i] {
			elevio.SetButtonLamp(0, i, true)
		}
		if ms.CURRENT_DATA.DOWN_BUTTON_ARRAY[i] {
			elevio.SetButtonLamp(1, i, true)
		}
		if ms.CURRENT_DATA.ELEVATOR_DATA[ms.UNIT_ID].INTERNAL_BUTTON_ARRAY[i] {
			elevio.SetButtonLamp(2, i, true)
		}
	}
	// time.Sleep(500 * time.Millisecond)
}

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
	//Receives peer update
	go peers.Receiver(peers_port, peers_update_channel)

	aliveCheck := make(chan structs.AliveMsg)

	go bcast.Receiver(broadcast_port, aliveCheck)

	//Prints peer update and adds peer info to current data
	for {
		select {
		case p := <-peers_update_channel:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			if p.New != "" {
				UpdateNewConnection(ms, p.New)
			}
			if p.Lost != nil {
				UpdateLostConnection(ms, p.Lost)
			}
		case a := <-aliveCheck:
			fmt.Printf("Received %#v \n", a)
		}
	}
}

// Changes alive status and adds address when a peer connects
func UpdateNewConnection(ms *MasterSlave, newElevatorID string) {

	elevatorNum, elevatorAddress := splitPeerString(newElevatorID)
	ms.CURRENT_DATA.ELEVATOR_DATA[elevatorNum].ADDRESS = elevatorAddress
	ms.CURRENT_DATA.ELEVATOR_DATA[elevatorNum].ALIVE = true

	newMasterChoice(ms)
}

// Changes alive status when a peer disconnects
func UpdateLostConnection(ms *MasterSlave, lostElevatorID []string) {
	for i := range lostElevatorID {
		elevatorNum, _ := splitPeerString(lostElevatorID[i])
		ms.CURRENT_DATA.ELEVATOR_DATA[elevatorNum].ALIVE = false
	}
	newMasterChoice(ms)
}

//var peerDataMap = make(map[string]elev_structs.SystemData)
// connectedPeers := []elev_structs.SystemData{}
// p := <-peers_update_channel
// fmt.Printf("Peers channel1: %v \n", p)

// for i := 0; i < structs.N_ELEVATORS; i++ {
// 	if ms.CURRENT_DATA.ELEVATOR_DATA[i].ALIVE {
// 		connectedPeers = append(connectedPeers, ms.CURRENT_DATA)
// 	}
// }
// for data, _:= range p.Peers {
// 	elevatorNum, _ := splitPeerString(data)
// 		connectedPeers = append(connectedPeers, ms.CURRENT_DATA)
// }
// fmt.Printf("Peers channel2: %v \n", connectedPeers)
// currentMasterId := election.DetermineMaster(strconv.Itoa(ms.UNIT_ID), strconv.Itoa(ms.CURRENT_DATA.MASTER_ID), connectedPeers, isMaster)
// ms.CURRENT_DATA.MASTER_ID, _ = strconv.Atoi(currentMasterId)

func newMasterChoice(ms *MasterSlave) {
	if !ms.CURRENT_DATA.ELEVATOR_DATA[ms.CURRENT_DATA.MASTER_ID].ALIVE {
		for i := 0; i < structs.N_ELEVATORS; i++ {
			if ms.CURRENT_DATA.ELEVATOR_DATA[i].ALIVE {
				ms.CURRENT_DATA.MASTER_ID = i
				fmt.Printf("New master: %v\n", ms.CURRENT_DATA.MASTER_ID)

				//ms.MainLoop()
				break
				//is_master := ms.CURRENT_DATA.MASTER_ID == ms.UNIT_ID
			}
		}
	}
}

// Splits peer string to the unit ID and address
func splitPeerString(peerString string) (elevatorNum int, elevatoraddress string) {
	splitString := strings.Split(peerString, "-")
	elevatorNum, err := strconv.Atoi(splitString[0])
	if err != nil {
		fmt.Printf("Error with string splitting: %v \n", err)
	}
	elevatorAddress := splitString[1]
	return elevatorNum, elevatorAddress
}
