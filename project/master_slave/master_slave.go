package master_slave

import  (
	"fmt"
	"net"
	"os/exec"
	//"strconv"
	"time"
	"elevator/structs"
	"elevator/network"
	scheduler "elevator/elevator-scheduler"
)

type MasterSlave struct {
	CURRENT_DATA *structs.SystemData
	IP_ADDRESS string
	ELEVATOR_NUMBER int
}

// Create a MasterSlave
func CreateMasterSlave(ElevatorNumber int) *MasterSlave {
	MS := new(MasterSlave)
	
	// Initialize current data
	MS.CURRENT_DATA = structs.SystemData{
        SENDER: 0,
        UP_BUTTON_ARRAY: &([structs.N_FLOORS]bool{}),
        DOWN_BUTTON_ARRAY: &([structs.N_FLOORS]bool{}),
        INTERNAL_BUTTON_ARRAY: &([structs.N_ELEVATORS][structs.N_FLOORS]bool{}),
        WORKING_ELEVATORS: &([structs.N_FLOORS]bool{}),
        ELEVATOR_STATES: &([]structs.ElevatorState{}),
        COUNTER: 0,
    }
	// Set elevator number
	MS.ELEVATOR_NUMBER = ElevatorNumber

	return MS
}



func (ms *MasterSlave) MainLoop() {
	// Check if this elevator is Master
	is_master := ms.CURRENT_DATA.SENDER == ms.ELEVATOR_NUMBER

	// Main loop of Master-slave
	for {
		if is_master {
			// Run if current elevator is master

			// TODO: Update SystemData:
			// Update calls, buttons
			// Update the states of each elevator
			// UpdateElevatorTargets()
			// Increase counter


			// Send updated SystemData
			network.SendSystemData(ms.CURRENT_DATA)
		
		
		} else {
			// Run if current elevator is slave

			// Receive data from master
			received_data := network.ReceiveSystemData()

			// Check if the received data is newer then current data, and update current data if so 
			if received_data.COUNTER > ms.CURRENT_DATA.COUNTER {
				ms.CURRENT_DATA = received_data
			}


			
		}
	}

}

func (ms *MasterSlave) UpdateElevatorTargets() {
	// Get new elevator targets
	movement_map := scheduler.CalculateElevatorMovement(ms.CURRENT_DATA)

	// Map to convert from map of elevators to array of elevators
	key_to_int_map := map[string]int{
		"one": 1,
	 	"two": 2, 
		"three": 3,
	}
	
	// Update values in ELEVATOR_TARGETS of SystemData
	for k, v := range movement_map {
		*ms.CURRENT_DATA.ELEVATOR_TARGETS[key_to_int_map[k]] = v;
	}
}


// HandleOrderFromMaster is a method on the MasterSlave struct that processes an order from the master.
func (ms *MasterSlave) HandleOrderFromMaster(order *structs.ElevatorState) error {
	// Check if the target floor in the order is valid (between 0 and 3)
	if order.TARGET_FLOOR < 0 || order.TARGET_FLOOR > structs.N_FLOORS {
		return fmt.Errorf("Invalid order: floor must be between 0 and 3")
	}
	// Check if the direction in the order is valid (0 for stop, 1 for up, 2 for down)
	if order.DIRECTION < 0 || order.DIRECTION > 2 {
		return fmt.Errorf("Invalid order: direction must be 0, 1 or 2")
	}

	// Update the SystemData based on the order
	// If the direction is 1 (up), set the corresponding floor in the up button array to true
	if order.DIRECTION == 1 {
		ms.CURRENT_DATA.UP_BUTTON_ARRAY[order.TARGET_FLOOR] = true
	// If the direction is 2 (down), set the corresponding floor in the down button array to true
	} else if order.DIRECTION == 2 {
		ms.CURRENT_DATA.DOWN_BUTTON_ARRAY[order.TARGET_FLOOR] = true
	// If the direction is 0 (stop), do nothing
	} else {
		// TODO: Set internal orders for given elevator
		// ms.current_data.INTERNAL_BUTTON_ARRAY[order.TARGET_FLOOR] = true
	}
	// Print a message indicating that the order has been processed
	fmt.Printf("Order for floor %d with direction %d has been processed.\n", order.TARGET_FLOOR, order.DIRECTION)
	return nil
}

// func (ms *structs.SystemData) SwitchToBackup() {
// 	ms.SENDER = 0
// 	fmt.Println("Master is dead, switching to backup")
// }


var fullAddress = structs.SERVER_IP_ADDRESS + ":" + structs.PORT

func StartMasterSlave(leader *MasterSlave) {
	//Set the leader as the Master
	leader.Master = true

	// start_time := time.Now()
	print_counter := time.Now()
	counter := 0

	ms := &MasterSlave{}


	filename := "/home/student/Documents/AjananMiaSindre/Sanntid/exercise_4/main.go"

	listener, err := net.Listen("tcp", fullAddress)
	if err != nil {
		fmt.Printf("Error creating TCP listener: %v\n", err)
		return
	}
	defer listener.Close()

	exec.Command("gnome-terminal", "--", "go", "run", filename).Run()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting TCP connection: %v\n", err)
			continue
		}

		// Handle incoming TCP connection
		go func(conn net.Conn) {
			defer conn.Close()

			// Receive SystemData from the master
			data := &structs.SystemData{}
			if err := network.receiveSystemData(conn, data); err != nil {
				fmt.Printf("Error receiving SystemData: %v\n", err)
				return
			}

			// Process received SystemData
			// (Add your logic here based on the received data)

		}(conn)

		// Send SystemData to the master periodically
		if time.Since(print_counter).Seconds() > 1 {
			counter++
			ms.CURRENT_DATA.COUNTER = counter
			if err := network.sendSystemData(conn, ms.CURRENT_DATA); err != nil {
				fmt.Printf("Error sending SystemData: %v\n", err)
			}

			fmt.Printf("%d\n", counter)
			print_counter = time.Now()
		}
	}
}


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


//TODO: Implement function to check if the new targets differ from the current ones
func (ms *MasterSlave) CheckIfReceivedNewTargets() {
	ms.CURRENT_DATA.COUNTER
}



