package singleelev

import (
	// "flag"
	// "fmt"
	// "os"
	"sync"
	"time"
	"encoding/json"

	"Driver-go/elevio"

	master_slave "elevator/master-slave"
	"elevator/structs"
	tcp_interface "elevator/tcp-interface"
)

//TODO: Remove unused code

// Temporary placement of Mutex
var order_mutex sync.Mutex

type Elevator struct {
	// The buffer values received from the elevio interface
	button_order  *elevio.ButtonEvent
	current_floor *int
	is_obstructed *bool
	is_stopped    *bool

	// Variable containing the current state
	internal_state *structs.ElevatorState

	// Variable showing the last visited floor
	at_floor *int

	// The current target of the elevator (-1 for no target)
	target_floor *int

	// Variable for the direction of the elevator
	moving_direction *structs.Direction

	// Variable for keeping track of when interrupt ends
	interrupt_end *time.Time
	ms_unit   *master_slave.MasterSlave
}

func MakeElevator(elevatorNumber int, master *master_slave.MasterSlave) Elevator {
	// Set state to idle
	var start_state structs.ElevatorState = structs.IDLE

	// Exception value
	starting_floor := -1

	// Target floor
	target_floor := -1

	// Starting direction
	starting_direction := structs.STILL

	// Pointer values
	floor_number := -1
	is_obstructed := false
	is_stopped := false

	start_time := time.Now()

	return Elevator{
		&elevio.ButtonEvent{},
		&floor_number,
		&is_obstructed,
		&is_stopped,
		&start_state,
		&starting_floor,
		&target_floor,
		&starting_direction,
		&start_time,
		master}
}

func (e Elevator) ElevatorLoop() {

	if *e.at_floor == -1 {
		elevio.SetMotorDirection(elevio.MD_Up)
		*e.internal_state = structs.MOVING
	}

	i := 0
	for {
		i += 1
		// Check for stop-button press

		if *e.is_stopped {
			// fmt.Print("Stop\n")
			*e.internal_state = structs.STOPPED
		}

		// fmt.Printf("Internal_state: %v", e.internal_state)
		// fmt.Printf("Stopped: %t \n", *e.is_stopped)
		// fmt.Printf("Obstructed: %t \n", *e.is_obstructed)
		// fmt.Printf("current floor: %d\n", *e.current_floor)
		// fmt.Printf("at floor: %d\n", *e.at_floor)
		// fmt.Printf("target floor: %d\n", *e.target_floor)
		// fmt.Printf("moving direction: %d\n", *e.moving_direction)
		// fmt.Printf("Count %d \n", i)
		// fmt.Printf("---\n")

		switch state := *e.internal_state; state {
		case structs.IDLE:
			// fmt.Printf("Idle\n")

			// Update floor when channel gets new value
			if *e.current_floor != *e.at_floor {
				*e.at_floor = *e.current_floor
				elevio.SetFloorIndicator(*e.at_floor)
			}

			// Either move to existing target or choose new target
			if *e.target_floor != -1 {
				// Move towards target if there is one
				e.MoveToTarget()
			} else {
				// TODO: Update this one
				if *e.at_floor != -1 {
					//TODO: Find out why this causes the loop to run slower, and fix
					// e.ClearOrdersAtFloor()
				}
			}

			// Pick new target if none
			e.PickTarget()

		case structs.MOVING:
			e.PickTarget()

			// Run when arriving at new floor
			if *e.current_floor != *e.at_floor {
				// Update value of master
				e.AddElevatorDataToMaster()

				// Set correct floor
				*e.at_floor = *e.current_floor
				elevio.SetFloorIndicator(*e.at_floor)

				// Run visit floor routine
				e.Visit_floor()
			}

		case structs.DOOR_OPEN:
			e.OpenDoor()

		case structs.STOPPED:
			e.Stop()
		}
		// time.Sleep(500 * time.Millisecond)
	}
}

// Read from the channels and put data into variables
func (e Elevator) ReadChannels(current_floor chan int, is_obstructed chan bool, is_stopped chan bool) {

	for {
		select {
		case cf := <-current_floor:
			*e.current_floor = cf
			// fmt.Printf("\n From channel: %t \n", cf)

		case io := <-is_obstructed:
			*e.is_obstructed = io

		case is := <-is_stopped:
			order_mutex.Lock()
			*e.is_stopped = is
			order_mutex.Unlock()
			// fmt.Printf("Stopping: %t\n", *e.is_stopped)
		}
	}
}

// Clears orders when they appear at the same floor as the elevator
func (e Elevator) ClearOrdersAtFloor() {
	// // Check if any of the orders are for the current floor
	// if e.internal_button_array[*e.at_floor] || e.up_button_array[*e.at_floor] || e.down_button_array[*e.at_floor] {
	// 	// fmt.Printf("ClearOrdersAtFloor\n")

	//TODO: Handle Clear orders at floor	

	// Open door
	e.TransitionToOpenDoor()

	if *e.target_floor == *e.at_floor {
		*e.target_floor = -1
	}

	// // Remove all orders on floor
	// e.internal_button_array[*e.at_floor] = false
	// e.up_button_array[*e.at_floor] = false
	// e.down_button_array[*e.at_floor] = false

	// Reset all lights
	elevio.SetButtonLamp(0, *e.at_floor, false)
	elevio.SetButtonLamp(1, *e.at_floor, false)
	elevio.SetButtonLamp(2, *e.at_floor, false)
	// }

}

func (e Elevator) PickTarget() {
	cab_calls := e.ms_unit.CURRENT_DATA.ELEVATOR_DATA[e.ms_unit.UNIT_ID].INTERNAL_BUTTON_ARRAY
	up_calls := e.ms_unit.CURRENT_DATA.UP_BUTTON_ARRAY
	down_calls := e.ms_unit.CURRENT_DATA.DOWN_BUTTON_ARRAY
	// Sets new target to closest floor, prioritizing floors above

	// TODO: Add check to see if there are new orders instead of running this loop every time

	// This code can be reworked to better adhere to the DRY-principle
	// Check floors above

	for i := 1; i <= 3; i++ {
		if *e.at_floor+i < 4 {

			// Check floors above
			check_floor := *e.at_floor + i

			if check_floor < 0 || check_floor > 4 {
				continue
			}

			// Set target if an order exists on floor
			if up_calls[check_floor] || down_calls[check_floor] || cab_calls[check_floor] {
				*e.target_floor = check_floor
				return
			} 

		}
		if *e.at_floor-i >= 0 {
			// Check floors below
			check_floor := *e.at_floor - i

			if check_floor < 0 || check_floor > 4 {
				continue
			}

			// Set target if an order exists on floor
			if up_calls[check_floor] || down_calls[check_floor] || cab_calls[check_floor] {
				*e.target_floor = check_floor
				return
			} 
		}
	}

	*e.target_floor = -1
}

func (e Elevator) Visit_floor() {

	// Run when no floor at initialization
	if *e.target_floor == -1 {
		elevio.SetMotorDirection(elevio.MD_Stop)
		*e.internal_state = structs.IDLE
	}

	*e.at_floor = *e.current_floor

	if *e.at_floor == *e.target_floor {
		// Reset internal button
		elevio.SetButtonLamp(2, *e.at_floor, false)

		// // Make sure the correct orders are removed
		// e.RemoveOrdersAtFloor(*e.at_floor, *e.moving_direction)
		id := e.ms_unit.UNIT_ID
		e.RemoveOrdersAtFloor(*e.at_floor, e.ms_unit.CURRENT_DATA.ELEVATOR_DATA[id].ELEVATOR_TARGETS[*e.at_floor])

		// Transition to OpenDoor state
		e.TransitionToOpenDoor()

		// TODO: Figure out logic when several buttons are pressed at target
		elevio.SetButtonLamp(0, *e.at_floor, false)
		elevio.SetButtonLamp(1, *e.at_floor, false)
		elevio.SetButtonLamp(2, *e.at_floor, false)

	}
}

func (e Elevator) TransitionToOpenDoor() {
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetDoorOpenLamp(true)
	*e.internal_state = structs.DOOR_OPEN
}

func (e Elevator) OpenDoor() {

	// Keep door open until not obstructed
	obstruction_check := *e.is_obstructed

	if !(obstruction_check) {

		// Close door
		time.Sleep(3 * time.Second)
		elevio.SetDoorOpenLamp(false)

		*e.internal_state = structs.IDLE

		// Remove target if current floor is target floor
		if *e.at_floor == *e.target_floor {
			*e.target_floor = -1
		}
	}

}

func (e Elevator) MoveToTarget() {
	// Set state to MOVING and set motor direction
	*e.internal_state = structs.MOVING

	if *e.target_floor > *e.at_floor {
		*e.moving_direction = structs.UP
		elevio.SetMotorDirection(elevio.MD_Up)
	} else if *e.target_floor < *e.at_floor {
		*e.moving_direction = structs.DOWN
		elevio.SetMotorDirection(elevio.MD_Down)
	}
}

func (e Elevator) Stop() {
	// Handles stopping

	elevator_stop := *e.is_stopped

	elevio.SetStopLamp(true)
	elevio.SetMotorDirection(elevio.MD_Stop)

	if !elevator_stop {
		time.Sleep(3 * time.Second)
		*e.internal_state = structs.IDLE
		elevio.SetStopLamp(false)
		elevio.SetDoorOpenLamp(false)
		// fmt.Print(e.internal_state)
	}
}
func (e Elevator) RemoveOrdersAtFloor(floor int, calls [2]bool) {
	//TODO: Fill in function. May have to be placed in Master-Slave or elsewhere
	if calls[0] {
		e.ClearOrderFromMaster(floor, structs.UP)
	} else if calls[1] {
		e.ClearOrderFromMaster(floor, structs.DOWN)
	}
	
}



// Send new cab orders to master
func (e Elevator) AddCabOrderToMaster(floor int) {
	master_id := e.ms_unit.CURRENT_DATA.MASTER_ID
	unit_id := e.ms_unit.UNIT_ID
	if  master_id !=  unit_id {

		// Encode data
		data := [structs.N_FLOORS]bool{false, false, false, false}
		data[floor] = true
		encoded_data, _ := json.Marshal(&data)

		// Send data to master
		e._message_data_to_master(encoded_data, structs.NEWCABCALL)
	}

	e.ms_unit.CURRENT_DATA.ELEVATOR_DATA[unit_id].INTERNAL_BUTTON_ARRAY[floor] = true

	
}

func (e Elevator) AddHallOrderToMaster(floor int, dir structs.Direction) {
	master_id := e.ms_unit.CURRENT_DATA.MASTER_ID
	if  master_id != e.ms_unit.UNIT_ID {
		// Encode data
		data := structs.ClearHallorderMsg{
			Clear_floor: floor,
			Clear_direction: dir,
		}
		
		encoded_data, _ := json.Marshal(&data)

		e._message_data_to_master(encoded_data, structs.NEWHALLORDER)
	}
	

	// Set cab order
	if dir == structs.UP {
		e.ms_unit.CURRENT_DATA.UP_BUTTON_ARRAY[floor] = true
	} else if dir == structs.DOWN {
		e.ms_unit.CURRENT_DATA.DOWN_BUTTON_ARRAY[floor] = true
	}
}

func (e Elevator) AddElevatorDataToMaster() {
	master_id := e.ms_unit.CURRENT_DATA.MASTER_ID
	unit_id := e.ms_unit.UNIT_ID

	

	if  master_id !=  unit_id {

		// Encode data
		data := tcp_interface.EncodeSystemData(e.ms_unit.CURRENT_DATA)

		// Send data to master
		e._message_data_to_master(data, structs.UPDATEELEVATOR)
	}
}

func (e Elevator) ClearOrderFromMaster(floor int, dir structs.Direction) {
	master_id := e.ms_unit.CURRENT_DATA.MASTER_ID
	unit_id := e.ms_unit.UNIT_ID
	if  master_id != e.ms_unit.UNIT_ID {
		// Encode data
		data := structs.ClearHallorderMsg{
			Clear_floor: floor,
			Clear_direction: dir,
		}
		
		encoded_data, _ := json.Marshal(&data)

		e._message_data_to_master(encoded_data, structs.CLEARHALLORDER)
	}
	

	// Clear internal order
	e.ms_unit.CURRENT_DATA.ELEVATOR_DATA[unit_id - 1].INTERNAL_BUTTON_ARRAY[floor] = false
	
	// Clear order in the given direction
	if dir == structs.UP {
		e.ms_unit.CURRENT_DATA.UP_BUTTON_ARRAY[floor] = false
	} else if dir == structs.DOWN {
		e.ms_unit.CURRENT_DATA.DOWN_BUTTON_ARRAY[floor] = false
	}
}

// Send a tcp-message with data to the master unit
func (e Elevator) _message_data_to_master(data []byte, msg_type structs.MessageType) {
	master_id := e.ms_unit.CURRENT_DATA.MASTER_ID
		// Construct message
		msg := structs.TCPMsg{
			MessageType: msg_type,
			Sender_id: e.ms_unit.UNIT_ID,
			Data: data,
		}
		encoded_msg := tcp_interface.EncodeMessage(&msg)
		// Send message to master
		tcp_interface.SendData(e.ms_unit.CURRENT_DATA.ELEVATOR_DATA[master_id].ADDRESS, encoded_msg)
}


// Reset all elevator elements
func ResetElevator() {
	// Set motor direction to stop
	elevio.SetMotorDirection(elevio.MD_Stop)

	// Turn off stop lamb
	elevio.SetStopLamp(false)

	// Turn of open door lamp
	elevio.SetDoorOpenLamp(false)

	// Reset all order lights
	for f := 0; f < structs.N_FLOORS; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
			elevio.SetButtonLamp(b, f, false)
		}
	}
}
