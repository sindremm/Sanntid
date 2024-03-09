package singleelev

import (
	"Driver-go/elevio"
	"elevator/structs"
	// "flag"
	// "fmt"
	// "os"
	"sync"
	"time"
)


// Temporary placement of Mutex
var order_mutex sync.Mutex

type Elevator struct {
	// The buffer values received from the elevio interface
	button_order  *elevio.ButtonEvent
	current_floor *int
	is_obstructed *bool
	is_stopped    *bool

	// Arrays that show awhich buttons have been pressed
	up_button_array       *[4]bool
	down_button_array     *[4]bool
	internal_button_array *[4]bool

	// Variable containing the current state
	internal_state *structs.State

	// Variable showing the last visited floor
	at_floor *int

	// The current target of the elevator (-1 for no target)
	target_floor *int

	// Variable for the direction of the elevator
	moving_direction *structs.Direction

	// Variable for keeping track of when interrupt ends
	interrupt_end *time.Time
}

func MakeElevator() Elevator {
	// Set state to idle
	var start_state structs.State = structs.IDLE

	// Exception value
	starting_floor := -1

	// Target floor
	target_floor := -1

	// Starting direction
	starting_direction := structs.STILL

	// Initialize empty button arrays
	up_array := [structs.N_FLOORS]bool{}
	down_array := [structs.N_FLOORS]bool{}
	internal_array := [structs.N_FLOORS]bool{}

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
		&up_array,
		&down_array,
		&internal_array,
		&start_state,
		&starting_floor,
		&target_floor,
		&starting_direction,
		&start_time}

}

func (e Elevator) Main() {

	if *e.at_floor == -1 {
		elevio.SetMotorDirection(elevio.MD_Up)
		*e.internal_state = structs.MOVING
	}

	// fmt.Printf("%s", e.internal_state)
	for {
		// Check for stop-button press
		//fmt.Printf("Stopped: %t \n", *e.is_stopped)
		//fmt.Printf("Obstructed: %t \n", *e.is_obstructed)
		//fmt.Printf("Obstructed: %t \n", *e.current_floor)
		if *e.is_stopped {
			// fmt.Print("Stop\n")
			*e.internal_state = structs.STOPPED
		}

		// fmt.Printf("current floor: %d\n", *e.current_floor)
		// fmt.Printf("at floor: %d\n", *e.at_floor)
		// fmt.Printf("target floor: %d\n", *e.target_floor)
		// fmt.Printf("moving direction: %d\n", *e.moving_direction)
		// fmt.Printf("d buttons: %t\n", *e.down_button_array)
		// fmt.Printf("u buttons: %t\n", *e.up_button_array)
		// fmt.Printf("i buttons:%t\n", *e.internal_button_array)
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
				e.MoveToTarget()
			} else {
				if *e.at_floor != -1 {
					e.ClearOrdersAtFloor()
				}

				e.PickTarget()
			}
			

		case structs.MOVING:
			// fmt.Printf("Moving\n")

			// Run when arriving at new floor
			if *e.current_floor != *e.at_floor {
				*e.at_floor = *e.current_floor
				elevio.SetFloorIndicator(*e.at_floor)
				e.Visit_floor()
			}

		case structs.DOOR_OPEN:
			// fmt.Printf("open door\n")
			e.OpenDoor()

		case structs.STOPPED:
			e.Stop()
		}
		// time.Sleep(500 * time.Millisecond)
	}
}

// Read from the channels and put data into variables
func (e Elevator) ReadChannels(button_order chan elevio.ButtonEvent, current_floor chan int, is_obstructed chan bool, is_stopped chan bool) {

	for {
		select {
		case bo := <-button_order:
			// Transform order to readable format
			floor, btn := e.ReadOrder(bo)
			// Add order to internal array and set lights
			e.AddOrders(floor, btn)

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
	// Check if any of the orders are for the current floor
	if e.internal_button_array[*e.at_floor] || e.up_button_array[*e.at_floor] || e.down_button_array[*e.at_floor] {
		// fmt.Printf("ClearOrdersAtFloor\n")
		
		// Open door
		e.TransitionToOpenDoor()

		if *e.target_floor == *e.at_floor {
			*e.target_floor = -1;
		}

		// Remove all orders on floor
		e.internal_button_array[*e.at_floor] = false
		e.up_button_array[*e.at_floor] = false
		e.down_button_array[*e.at_floor] = false

		// Reset all lights
		elevio.SetButtonLamp(0, *e.at_floor, false)
		elevio.SetButtonLamp(1, *e.at_floor, false)
		elevio.SetButtonLamp(2, *e.at_floor, false)
	}

}

func (e Elevator) PickTarget() {
	// Sets new target to closest floor, prioritizing floors above
	new_target := -1

	// TODO: Add check to see if there are new orders instead of running this loop every time

	// This code can be reworked to better adhere to the DRY-principle
	// Check floors above

	for i := 1; i <= 3; i++{
		if *e.at_floor+i < 4 {

			// Check floors above
			check_floor := *e.at_floor + i

			if check_floor < 0 || check_floor > 4 {
				continue
			}

			// Set target if an order exists on floor
			if e.up_button_array[check_floor] || e.down_button_array[check_floor] || e.internal_button_array[check_floor] {
				new_target = check_floor
				break
			}

		}
		if *e.at_floor-i >= 0 {
			// Check floors below
			check_floor := *e.at_floor - i

			if check_floor < 0 || check_floor > 4 {
				continue
			}

			// Set target if an order exists on floor
			if e.up_button_array[check_floor] || e.down_button_array[check_floor] || e.internal_button_array[check_floor] {
				new_target = check_floor
				break
			}
		}
	}

	*e.target_floor = new_target
	
}

func (e Elevator) AddOrders(floor int, button elevio.ButtonType) {
	// Set elevator lights
	elevio.SetButtonLamp(button, floor, true)
	switch button {
	case 0:
		e.up_button_array[floor] = true
	case 1:
		e.down_button_array[floor] = true
	case 2:
		e.internal_button_array[floor] = true
	}
}


// Convert order to readable format
func (e Elevator) ReadOrder(button_order elevio.ButtonEvent) (floor int, button elevio.ButtonType) {
	order_floor := button_order.Floor
	order_button := button_order.Button

	return order_floor, order_button
}

func (e Elevator) Visit_floor() {

	// Run when no floor at initialization
	if *e.target_floor == -1 {
		elevio.SetMotorDirection(elevio.MD_Stop)
		*e.internal_state = structs.IDLE
	}

	*e.at_floor = *e.current_floor

	// Remove internal order when opening door at requested floor, and opens door
	if e.internal_button_array[*e.at_floor] {
		// Remove order
		e.internal_button_array[*e.at_floor] = false
		elevio.SetButtonLamp(2, *e.at_floor, false)	

		
		// Open door
		e.TransitionToOpenDoor()
	}

	// Remove orders in same direction, and sets door to open
	switch *e.moving_direction {
	case structs.UP:
		if e.up_button_array[*e.at_floor] {
			e.up_button_array[*e.at_floor] = false
			elevio.SetButtonLamp(0, *e.at_floor, false)
			
			
			// Open door
			e.TransitionToOpenDoor()
		}
	case structs.DOWN:
		if e.down_button_array[*e.at_floor] {
			e.down_button_array[*e.at_floor] = false
			elevio.SetButtonLamp(1, *e.at_floor, false)
			
			
			// Open door
			e.TransitionToOpenDoor()
		}
	}
	// Reset internal button
	elevio.SetButtonLamp(2, *e.at_floor, false)

	if *e.at_floor == *e.target_floor {
		e.TransitionToOpenDoor()

		e.internal_button_array[*e.at_floor] = false
		e.up_button_array[*e.at_floor] = false
		e.down_button_array[*e.at_floor] = false


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