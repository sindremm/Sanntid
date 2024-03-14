package master_slave

// import  (
// 	// "fmt"
// 	// "net"
// 	// "os/exec"
// 	//"strconv"
// 	// "time"
	
// 	"Driver-go/elevio"

// 	"elevator/structs"
// 	tcp_interface "elevator/tcp-interface"
// 	scheduler "elevator/elevator-scheduler"
// 	single "elevator/single-elevator"
// )


// type MasterSlave struct {
// 	CURRENT_DATA *structs.SystemData
// 	IP_ADDRESS string
// 	UNIT_ID int
// 	ELEVATOR_UNIT single.Elevator
// 	LISTEN_PORT string
// }

// // Create a MasterSlave
// func MakeMasterSlave(UnitID int, port string, elevator single.Elevator) *MasterSlave {
// 	MS := new(MasterSlave)
	
// 	// Initialize current data
// 	SD := structs.SystemData{
//         MASTER_ID: 0,
//         UP_BUTTON_ARRAY: &([structs.N_FLOORS]bool{}),
//         DOWN_BUTTON_ARRAY: &([structs.N_FLOORS]bool{}),
//         ELEVATOR_DATA: &([structs.N_ELEVATORS]structs.ElevatorData{}),
//         COUNTER: 0,
//     }

// 	// Set data
// 	MS.CURRENT_DATA = &SD
	
// 	// Set identifying ID of unit
// 	MS.UNIT_ID = UnitID
	
// 	// Set corresponding elevator
// 	MS.ELEVATOR_UNIT = elevator

// 	// Set the port where tcp messages are received
// 	MS.LISTEN_PORT = port

// 	// Start threads
// 	go elevator.Main()

// 	return MS
// }



// func (ms *MasterSlave) MainLoop() {
// 	// Check if this elevator is Master
// 	is_master := ms.CURRENT_DATA.MASTER_ID == ms.UNIT_ID


// 	// Main loop of Master-slave
// 	for {
// 		if is_master {
// 			// Run if current elevator is master

// 			// TODO: Update SystemData:
// 			// Update calls, buttons
// 			// Update the states of each elevator
// 			// UpdateElevatorTargets() (Only run when new calls, or update in state of elevator)
// 			// Increase counter


// 			// Send updated SystemData
// 			ms.BroadcastSystemData()
			
		
		
// 		} else {
// 			// Run if current elevator is slave

// 			// Receive data from master
// 			own_address := ms.IP_ADDRESS + ms.LISTEN_PORT 
// 			received_data := new(structs.SystemData)
// 			tcp_interface.ReceiveSystemData(own_address, received_data)

// 			// Check if the received data is newer then current data, and update current data if so 
// 			if received_data.COUNTER > ms.CURRENT_DATA.COUNTER {
// 				ms.CURRENT_DATA = received_data
// 			}

// 			calls := ms.CURRENT_DATA.ELEVATOR_DATA[ms.UNIT_ID].ELEVATOR_TARGETS
// 			ms.ELEVATOR_UNIT.PickTarget(calls)
			
// 		}
// 	}

// }

// func (ms *MasterSlave) BroadcastSystemData() {
// 	// Send system data to each elevator
// 	for i := 0; i < structs.N_ELEVATORS; i++ {
// 		// Find corresponding address of elevator client
// 		client_address := ms.CURRENT_DATA.ELEVATOR_DATA[i].ADDRESS
// 		// Send system data to client
// 		tcp_interface.SendSystemData(client_address, ms.CURRENT_DATA)
// 	}
	
// }

// // Read from the channels and put data into variables
// func (ms *MasterSlave) ReadButtons(button_order chan elevio.ButtonEvent) {

// 	for {
// 		select {
// 		case bo := <-button_order:
// 			// Transform order to readable format
// 			floor, btn := ms.InterpretOrder(bo)
// 			// Add order to internal array and set lights
// 			ms.AddOrderToSystemDAta(floor, btn)
// 			elevio.SetButtonLamp(btn, floor, true)
// 		}
// 	}
// }



// // Convert order to readable format
// func (ms *MasterSlave) InterpretOrder(button_order elevio.ButtonEvent) (floor int, button elevio.ButtonType) {
// 	order_floor := button_order.Floor
// 	order_button := button_order.Button

// 	return order_floor, order_button
// }

// // Add order to system data
// func (ms *MasterSlave) AddOrderToSystemDAta(floor int, button elevio.ButtonType) {
// 	switch button {
// 	case 0:
// 		ms.CURRENT_DATA.UP_BUTTON_ARRAY[floor] = true
// 	case 1:
// 		ms.CURRENT_DATA.DOWN_BUTTON_ARRAY[floor] = true
// 	case 2:
// 		ms.CURRENT_DATA.ELEVATOR_DATA[ms.UNIT_ID].INTERNAL_BUTTON_ARRAY[floor] = true
// 	}
// }


// func (ms *MasterSlave) UpdateElevatorTargets() {
// 	// Get new elevator targets
// 	movement_map := *scheduler.CalculateElevatorMovement(*(ms.CURRENT_DATA))

// 	// Map to convert from map of elevators to array of elevators
// 	key_to_int_map := map[string]int{
// 		"one": 1,
// 	 	"two": 2, 
// 		"three": 3,
// 	}
	
// 	// Update values in ELEVATOR_TARGETS of SystemData
// 	for k := range movement_map {
// 		(*ms.CURRENT_DATA.ELEVATOR_DATA)[key_to_int_map[k]].ELEVATOR_TARGETS = movement_map[k];
// 	}
// }


// // // HandleOrderFromMaster is a method on the MasterSlave struct that processes an order from the master.
// // func (ms *MasterSlave) HandleOrderFromMaster(order *structs.ElevatorState) error {
// // 	// Check if the target floor in the order is valid (between 0 and 3)
// // 	if order.TARGET_FLOOR < 0 || order.TARGET_FLOOR > structs.N_FLOORS {
// // 		return fmt.Errorf("Invalid order: floor must be between 0 and 3")
// // 	}
// // 	// Check if the direction in the order is valid (0 for stop, 1 for up, 2 for down)
// // 	if order.DIRECTION < 0 || order.DIRECTION > 2 {
// // 		return fmt.Errorf("Invalid order: direction must be 0, 1 or 2")
// // 	}

// // 	// Update the SystemData based on the order
// // 	// If the direction is 1 (up), set the corresponding floor in the up button array to true
// // 	if order.DIRECTION == 1 {
// // 		ms.CURRENT_DATA.UP_BUTTON_ARRAY[order.TARGET_FLOOR] = true
// // 	// If the direction is 2 (down), set the corresponding floor in the down button array to true
// // 	} else if order.DIRECTION == 2 {
// // 		ms.CURRENT_DATA.DOWN_BUTTON_ARRAY[order.TARGET_FLOOR] = true
// // 	// If the direction is 0 (stop), do nothing
// // 	} else {
// // 		// TODO: Set internal orders for given elevator
// // 		// ms.current_data.INTERNAL_BUTTON_ARRAY[order.TARGET_FLOOR] = true
// // 	}
// // 	// Print a message indicating that the order has been processed
// // 	fmt.Printf("Order for floor %d with direction %d has been processed.\n", order.TARGET_FLOOR, order.DIRECTION)
// // 	return nil
// // }

// // func (ms *structs.SystemData) SwitchToBackup() {
// // 	ms.SENDER = 0
// // 	fmt.Println("Master is dead, switching to backup")
// // }


// // fullAddress = structs.SERVER_IP_ADDRESS + ":" + structs.PORT

// // func StartMasterSlave(leader *MasterSlave) {
// // 	//Set the leader as the Master
// // 	leader.Master = true

// // 	// start_time := time.Now()
// // 	print_counter := time.Now()
// // 	counter := 0

// // 	ms := &MasterSlave{}


// // 	filename := "/home/student/Documents/AjananMiaSindre/Sanntid/exercise_4/main.go"

// // 	listener, err := net.Listen("tcp", fullAddress)
// // 	if err != nil {
// // 		fmt.Printf("Error creating TCP listener: %v\n", err)
// // 		return
// // 	}
// // 	defer listener.Close()

// // 	exec.Command("gnome-terminal", "--", "go", "run", filename).Run()

// // 	for {
// // 		conn, err := listener.Accept()
// // 		if err != nil {
// // 			fmt.Printf("Error accepting TCP connection: %v\n", err)
// // 			continue
// // 		}

// // 		// Handle incoming TCP connection
// // 		go func(conn net.Conn) {
// // 			defer conn.Close()

// // 			// Receive SystemData from the master
// // 			data := &structs.SystemData{}
// // 			if err := network.receiveSystemData(conn, data); err != nil {
// // 				fmt.Printf("Error receiving SystemData: %v\n", err)
// // 				return
// // 			}

// // 			// Process received SystemData
// // 			// (Add your logic here based on the received data)

// // 		}(conn)

// // 		// Send SystemData to the master periodically
// // 		if time.Since(print_counter).Seconds() > 1 {
// // 			counter++
// // 			ms.CURRENT_DATA.COUNTER = counter
// // 			if err := network.sendSystemData(conn, ms.CURRENT_DATA); err != nil {
// // 				fmt.Printf("Error sending SystemData: %v\n", err)
// // 			}

// // 			fmt.Printf("%d\n", counter)
// // 			print_counter = time.Now()
// // 		}
// // 	}
// // }


// // // sendSystemData is a function that sends SystemData over a TCP connection.
// // // It takes a net.Conn object representing the connection and a pointer to the SystemData object to be sent.
// // // It returns an error if any occurs during the process.
// // func sendSystemData(conn net.Conn, data *structs.SystemData) error {
// // 	// Create a new encoder that will write to conn
// // 	encoder := gob.NewEncoder(conn)
// // 	// Encode the SystemData object and send it over the connection
// // 	// If an error occurs during encoding, wrap it in a new error indicating that encoding failed
// // 	if err := encoder.Encode(data); err != nil {
// // 		return fmt.Errorf("failed to encode SystemData: %v", err)
// // 	}
// // 	// If no error occurred, return nil
// // 	return nil
// // }

// // // receiveSystemData is a function that receives SystemData over a TCP connection.
// // // It takes a net.Conn object representing the connection and a pointer to the SystemData object where the received data will be stored.
// // // It returns an error if any occurs during the process.
// // func receiveSystemData(conn net.Conn, data *structs.SystemData) error {
// // 	// Create a new decoder that will read from conn
// // 	decoder := gob.NewDecoder(conn)
// // 	// Decode the received data and store it in the SystemData object
// // 	// If an error occurs during decoding, wrap it in a new error indicating that decoding failed
// // 	if err := decoder.Decode(data); err != nil {
// // 		return fmt.Errorf("failed to decode SystemData: %v", err)
// // 	}
// // 	// If no error occurred, return nil
// // 	return nil
// // }


// // //TODO: Implement function to check if the new targets differ from the current ones
// // func (ms *MasterSlave) CheckIfReceivedNewTargets() {
// // 	ms.CURRENT_DATA.COUNTER
// // }