package singleelev

import (
	// "flag"
	"fmt"
	// "os"
	"encoding/json"
	"sync"
	"time"

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
	floor_sensor  *int
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
	ms_unit       *master_slave.MasterSlave
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

	for {
		// Check for stop-button press

		if *e.is_stopped {
			// fmt.Print("Stop\n")
			*e.internal_state = structs.STOPPED
			e.AddElevatorDataToMaster()
		}

		switch state := *e.internal_state; state {
		case structs.IDLE:
			// fmt.Printf("Idle\n")

			// Either move to existing target or choose new target
			if *e.target_floor != -1 {
				// Move towards target if there is one
				e.MoveToTarget()
			} else if *e.at_floor != -1 {
				// Pick new target if no target, and the floor of the elevator is known
				e.PickTarget()
			}

		case structs.MOVING:

			// Run when arriving at new floor or when starting from target floor
			if *e.at_floor != -1 {
				e.PickTarget()
				e.MoveToTarget()
			}
			// fmt.Printf("State: >%v", *e.internal_state)

			if (*e.at_floor != *e.floor_sensor || *e.floor_sensor == *e.target_floor) && *e.floor_sensor != -1 {

				// Set correct floor if not in between floors
				*e.at_floor = *e.floor_sensor

				// Update value of master
				e.AddElevatorDataToMaster()

				elevio.SetFloorIndicator(*e.at_floor)

				// Run visit floor routine

				e.Visit_floor()
				continue
			}

		case structs.DOOR_OPEN:
			e.OpenDoor()

		case structs.STOPPED:
			e.Stop()
		}
		// e.AddElevatorDataToMaster()
		time.Sleep(100 * time.Millisecond)
	}
}

// Read from the channels and put data into variables
func (e Elevator) ReadChannels(button_order chan elevio.ButtonEvent, current_floor chan int, is_obstructed chan bool, is_stopped chan bool) {

	for {
		select {
		case bo := <-button_order:
			// Transform order to readable format
			floor, btn := e.InterpretOrder(bo)
			// Add order to internal array and set lights
			e.AddOrderToSystemDAta(floor, btn)

			// elevio.SetButtonLamp(btn, floor, true)

		case cf := <-current_floor:
			*e.floor_sensor = cf
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

// Convert order to readable format
func (e *Elevator) InterpretOrder(button_order elevio.ButtonEvent) (floor int, button elevio.ButtonType) {
	order_floor := button_order.Floor
	order_button := button_order.Button

	return order_floor, order_button
}

// Add order to system data
func (e *Elevator) AddOrderToSystemDAta(floor int, button elevio.ButtonType) {

	switch button {
	case 0:
		// fmt.Printf("Adding up order to floor %d\n", floor)
		e.AddHallOrderToMaster(floor, button)
		// e.ms_unit.CURRENT_DATA.UP_BUTTON_ARRAY[floor] = true
	case 1:
		// fmt.Printf("Adding down order to floor %d\n", floor)
		e.AddHallOrderToMaster(floor, button)
		// e.ms_unit.CURRENT_DATA.DOWN_BUTTON_ARRAY[floor] = true
	case 2:
		// fmt.Printf("Adding cab order to floor %d\n", floor)
		e.AddCabOrderToMaster(floor)
		// e.ms_unit.CURRENT_DATA.ELEVATOR_DATA[e.ms_unit.UNIT_ID].INTERNAL_BUTTON_ARRAY[floor] = true
	}
}

// // Clears orders when they appear at the same floor as the elevator
// func (e Elevator) ClearOrdersAtFloor() {
// 	// // Check if any of the orders are for the current floor
// 	// if e.internal_button_array[*e.at_floor] || e.up_button_array[*e.at_floor] || e.down_button_array[*e.at_floor] {
// 	// 	// fmt.Printf("ClearOrdersAtFloor\n")

// 	//TODO: Handle Clear orders at floor

// 	// Open door
// 	e.TransitionToOpenDoor()

// 	if *e.target_floor == *e.at_floor {
// 		*e.target_floor = -1
// 	}

// 	// // Remove all orders on floor
// 	// e.internal_button_array[*e.at_floor] = false
// 	// e.up_button_array[*e.at_floor] = false
// 	// e.down_button_array[*e.at_floor] = false

// 	// Reset all lights
// 	// elevio.SetButtonLamp(0, *e.at_floor, false)
// 	// elevio.SetButtonLamp(1, *e.at_floor, false)
// 	// elevio.SetButtonLamp(2, *e.at_floor, false)
// 	// }

// }

func (e *Elevator) PickTarget() {
	self := e.ms_unit.CURRENT_DATA.ELEVATOR_DATA[e.ms_unit.UNIT_ID]

	cab_calls := self.INTERNAL_BUTTON_ARRAY
	targets := self.ELEVATOR_TARGETS
	up_calls := [structs.N_FLOORS]bool{false, false, false, false}
	down_calls := [structs.N_FLOORS]bool{false, false, false, false}

	for i := 0; i < structs.N_FLOORS; i++ {
		up_calls[i] = targets[i][0]
		down_calls[i] = targets[i][1]
	}

	// fmt.Printf("Targets: %v \n", targets)
	// fmt.Printf("Targets (UP): %v \n ", up_calls)
	// fmt.Printf("Targets (DOWN): %v \n ", down_calls)
	// Sets new target to closest floor, prioritizing floors above

	// TODO: Add check to see if there are new orders instead of running this loop every time

	// This code can be reworked to better adhere to the DRY-principle
	// Check floors above
	new_target := *e.target_floor
	updated := false

	for i := 0; i <= 3; i++ {
		if *e.at_floor+i < structs.N_FLOORS {

			// Check floors above
			check_floor := *e.at_floor + i

			// Return if order is out of bound, or if elevator is moving in oposite direction
			if check_floor < 0 || check_floor > 4 || *e.moving_direction == structs.DOWN {
				continue
			}

			// Set target if an order exists on floor
			// Check if elevator is still or going upwards

			// Only move to down-calls when staying still
			down_when_still := down_calls[check_floor] && (*e.moving_direction == structs.STILL)
			// Always serve up calls and cab_calls
			if up_calls[check_floor] || cab_calls[check_floor] || down_when_still {
				new_target = check_floor
				break
			}

		}
		if *e.at_floor-i >= 0 {
			// Check floors below
			check_floor := *e.at_floor - i

			// Return if order is out of bound, or if elevator is moving in oposite direction
			if check_floor < 0 || check_floor > 4 || *e.moving_direction == structs.UP {
				continue
			}

			// Only move to down-calls when staying still
			up_when_still := up_calls[check_floor] && (*e.moving_direction == structs.STILL)
			if down_calls[check_floor] || cab_calls[check_floor] || up_when_still {
				new_target = check_floor
				break
			}
		}
	}

	// fmt.Printf("Picking target new target: %d \n", new_target)
	if updated {
		*e.target_floor = new_target

		// Update value of master
		e.AddElevatorDataToMaster()
	}

}

func (e Elevator) Visit_floor() {

	// The only time the code reaches this state is during initialization
	if *e.target_floor == -1 {
		elevio.SetMotorDirection(elevio.MD_Stop)
		*e.internal_state = structs.IDLE
		e.AddElevatorDataToMaster()
		return
	}

	if *e.at_floor == *e.target_floor {

		// Reset target
		*e.target_floor = -1
		*e.moving_direction = structs.STILL
		fmt.Printf("At DOOR_OPEN\n")
		*e.internal_state = structs.DOOR_OPEN

		// // Make sure the correct orders are removed
		// e.RemoveOrdersAtFloor(*e.at_floor, *e.moving_direction)
		// Find id
		// fmt.Printf("at_floor: %d\n", *e.at_floor)
		// fmt.Printf("current_floor: %d\n", *e.current_floor)
		// fmt.Printf("target_floor: %d\n", *e.target_floor)
		id := e.ms_unit.UNIT_ID
		// Find corresponding unit
		unit := e.ms_unit.CURRENT_DATA.ELEVATOR_DATA[id]
		// Get allocated orders at floor
		target := unit.ELEVATOR_TARGETS[*e.at_floor]

		e.ClearOrderFromMaster(*e.at_floor, target)
		e.AddElevatorDataToMaster()

		// Transition to OpenDoor state
		e.TransitionToOpenDoor()

		// TODO: Figure out logic when several buttons are pressed at target
		elevio.SetButtonLamp(0, *e.at_floor, false)
		elevio.SetButtonLamp(1, *e.at_floor, false)
		elevio.SetButtonLamp(2, *e.at_floor, false)

	}
	// fmt.Printf("id: %d\n", e.ms_unit.UNIT_ID)
	//e._debug_print_internal_states()
	//e._debug_print_master_data()

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

		e.AddElevatorDataToMaster()
	}

}


func (e Elevator) MoveToTarget() {
	// Set state to MOVING and set motor direction
	//fmt.Printf("Moving to target\n")
	*e.internal_state = structs.MOVING

	if *e.target_floor > *e.at_floor {
		*e.moving_direction = structs.UP
		elevio.SetMotorDirection(elevio.MD_Up)
	} else if *e.target_floor < *e.at_floor {
		*e.moving_direction = structs.DOWN
		elevio.SetMotorDirection(elevio.MD_Down)
	}

	e.AddElevatorDataToMaster()
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

//TODO: Find somewhere to place in the code
// Send new cab orders to master
func (e Elevator) AddCabOrderToMaster(floor int) {
	// Encode data
	data := structs.HallorderMsg{
		Order_floor:     floor,
		Order_direction: [2]bool{false, false},
	}
	encoded_data, _ := json.Marshal(&data)

	// Send data to master if master is alive
	if e.ms_unit.CURRENT_DATA.ELEVATOR_DATA[e.ms_unit.CURRENT_DATA.MASTER_ID].ALIVE {
		e._message_data_to_master(encoded_data, structs.NEWCABCALL)
	}
}

func (e Elevator) AddHallOrderToMaster(floor int, button elevio.ButtonType) {
	dir_bool := [2]bool{false, false}
	if button == elevio.BT_HallUp {
		dir_bool[0] = true
	}
	if button == elevio.BT_HallDown {
		dir_bool[1] = true
	}

	// Encode data
	data := structs.HallorderMsg{
		Order_floor:     floor,
		Order_direction: dir_bool,
	}

	encoded_data, _ := json.Marshal(&data)

	e._message_data_to_master(encoded_data, structs.NEWHALLORDER)
}

func (e *Elevator) AddElevatorDataToMaster() {
	// Encode data
	encoded_data := tcp_interface.EncodeSystemData(e.ms_unit.CURRENT_DATA)

	// Send data to master
	e._message_data_to_master(encoded_data, structs.UPDATEELEVATOR)
}

func (e Elevator) ClearOrderFromMaster(floor int, dir [2]bool) {
	// Encode data
	data := structs.HallorderMsg{
		Order_floor:     floor,
		Order_direction: dir,
	}

	encoded_data, _ := json.Marshal(&data)

	e._message_data_to_master(encoded_data, structs.CLEARHALLORDER)
}

// Send a tcp-message with data to the master unit
func (e Elevator) _message_data_to_master(data []byte, msg_type structs.MessageType) {
	master_id := e.ms_unit.CURRENT_DATA.MASTER_ID
	
	// Construct message
	msg := structs.TCPMsg{
		MessageType: msg_type,
		Sender_id:   e.ms_unit.UNIT_ID,
		Data:        data,
	}
	encoded_msg := tcp_interface.EncodeMessage(&msg)
	
	// fmt.Printf("_message_data_to_master send type: %d\n", msg_type)
	// fmt.Printf("_message_data_to_master addres: %s\n", e.ms_unit.CURRENT_DATA.ELEVATOR_DATA[master_id].ADDRESS)
	// Send message to master
	tcp_interface.SendData(e.ms_unit.CURRENT_DATA.ELEVATOR_DATA[master_id].ADDRESS, encoded_msg)
}

//TODO: Remember to remove
func (e Elevator) _debug_print_internal_states() {
	fmt.Printf("Internal_state: %d", e.internal_state)
	fmt.Printf("Stopped: %t \n", *e.is_stopped)
	fmt.Printf("Obstructed: %t \n", *e.is_obstructed)
	fmt.Printf("Floor sensor: %d\n", *e.floor_sensor)
	fmt.Printf("At floor: %d\n", *e.at_floor)
	fmt.Printf("Target floor: %d\n", *e.target_floor)
	fmt.Printf("Moving direction: %d\n", *e.moving_direction)
	fmt.Printf("---\n")
}

func (e Elevator) _debug_print_master_data() {
	fmt.Printf("%s", structs.SystemData_to_string(*e.ms_unit.CURRENT_DATA))
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
