package master_slave

import  (
	// "fmt"
	// "net"
	// "os/exec"
	//"strconv"
	// "time"
	"Driver-go/elevio"
	"elevator/structs"
	tcp_interface "elevator/tcp-interface"
	scheduler "elevator/elevator-scheduler"
	single "elevator/single-elevator"
)

type MasterSlave struct {
	CURRENT_DATA *structs.SystemData
	IP_ADDRESS string
	UNIT_ID int
	ELEVATOR_UNIT single.Elevator
	LISTEN_PORT string
}

// Create a MasterSlave
func MakeMasterSlave(UnitID int, port string, elevator single.Elevator) *MasterSlave {
	MS := new(MasterSlave)
	
	// Initialize current data
	SD := structs.SystemData{
        MASTER_ID: 0,
        UP_BUTTON_ARRAY: &([structs.N_FLOORS]bool{}),
        DOWN_BUTTON_ARRAY: &([structs.N_FLOORS]bool{}),
        ELEVATOR_DATA: &([structs.N_ELEVATORS]structs.ElevatorData{}),
        COUNTER: 0,
    }

	// Set data
	MS.CURRENT_DATA = &SD
	
	// Set identifying ID of unit
	MS.UNIT_ID = UnitID
	
	// Set corresponding elevator
	MS.ELEVATOR_UNIT = elevator

	// Set the port where tcp messages are received
	MS.LISTEN_PORT = port

	// Start threads
	go elevator.Main()

	return MS
}


func (ms *MasterSlave) MainLoop() {
	// Check if this elevator is Master
	//If the MASTER_ID is not set or if the current UNIT_ID is lower than the MASTER_ID, set the current UNIT_ID as the MASTER_ID
	if ms.CURRENT_DATA.MASTER_ID == 0 || ms.UNIT_ID < ms.CURRENT_DATA.MASTER_ID {
		ms.CURRENT_DATA.MASTER_ID = ms.UNIT_ID
	}
	is_master := ms.CURRENT_DATA.MASTER_ID == ms.UNIT_ID

	// Main loop of Master-slave
	for {
		if is_master {
			// Run if current elevator is master

			// TODO: Update SystemData:
			// Update calls, buttons
			// Update the states of each elevator
			// UpdateElevatorTargets() (Only run when new calls, or update in state of elevator)
			// Increase counter


			// Send updated SystemData
			ms.BroadcastSystemData()	
		
		} else {
			// Run if current elevator is slave

			// Receive data from master
			own_address := ms.IP_ADDRESS + ms.LISTEN_PORT 
			received_data := new(structs.SystemData)
			tcp_interface.ReceiveSystemData(own_address, received_data)

			// Check if the received data is newer then current data, and update current data if so 
			if received_data.COUNTER > ms.CURRENT_DATA.COUNTER {
				ms.CURRENT_DATA = received_data
			}

			calls := ms.CURRENT_DATA.ELEVATOR_DATA[ms.UNIT_ID].ELEVATOR_TARGETS
			ms.ELEVATOR_UNIT.PickTarget(calls)
			
		}
	}
}

func (ms *MasterSlave) BroadcastSystemData() {
	// Send system data to each elevator
	for i := 0; i < structs.N_ELEVATORS; i++ {
		// Find corresponding address of elevator client
		client_address := ms.CURRENT_DATA.ELEVATOR_DATA[i].ADDRESS
		// Send system data to client
		tcp_interface.SendSystemData(client_address, ms.CURRENT_DATA)
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
	case 1:
		ms.CURRENT_DATA.DOWN_BUTTON_ARRAY[floor] = true
	case 2:
		ms.CURRENT_DATA.ELEVATOR_DATA[ms.UNIT_ID].INTERNAL_BUTTON_ARRAY[floor] = true
	}
}


func (ms *MasterSlave) UpdateElevatorTargets() {
	// Get new elevator targets
	movement_map := *scheduler.CalculateElevatorMovement(*(ms.CURRENT_DATA))

	// Map to convert from map of elevators to array of elevators
	key_to_int_map := map[string]int{
		"one": 1,
	 	"two": 2, 
		"three": 3,
	}
	
	// Update values in ELEVATOR_TARGETS of SystemData
	for k := range movement_map {
		(*ms.CURRENT_DATA.ELEVATOR_DATA)[key_to_int_map[k]].ELEVATOR_TARGETS = movement_map[k];
	}
}
