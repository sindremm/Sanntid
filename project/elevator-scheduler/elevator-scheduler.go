// package elevatorscheduler
package main

import (
	// "Driver-go/elevio"
	// "sync"
	"encoding/json"
	"fmt"
	"os/exec"
)

// const N_ELEVATORS int = 4
// const N_FLOORS int = 4

// type ElevatorScheduler struct {
// 	// The buffer values received from the elevio interface

// 	current_floor *[N_ELEVATORS]int
// 	is_obstructed *[N_ELEVATORS]bool
// 	is_stopped    *[N_ELEVATORS]bool

// 	up_button_array       *[N_FLOORS]bool
// 	down_button_array     *[N_FLOORS]bool
// 	internal_button_array *[N_FLOORS]bool
// }

// func MakeElevatorScheduler() ElevatorScheduler {
// 	// Set state to idle
// 	var start_state State = IDLE

// 	// Exception value
// 	starting_floor := -1

// 	// Target floor
// 	target_floor := -1

// 	// Starting direction
// 	starting_direction := STILL

// 	// Initialize empty button arrays
// 	up_array := [N_FLOORS]bool{}
// 	down_array := [N_FLOORS]bool{}
// 	internal_array := [N_FLOORS]bool{}

// 	// Pointer values
// 	floor_number := -1
// 	is_obstructed := false
// 	is_stopped := false

// 	start_time := time.Now()

// 	return ElevatorScheduler{
// 		&elevio.ButtonEvent{},
// 		&floor_number,
// 		&is_obstructed,
// 		&is_stopped,
// 		&up_array,
// 		&down_array,
// 		&internal_array,
// 		&start_state,
// 		&starting_floor,
// 		&target_floor,
// 		&starting_direction,
// 		&start_time}

// }

// // Temporary placement of Mutex
// var order_mutex sync.Mutex

// // Read from the channels and put data into variables
// func (e ElevatorScheduler) ReadChannels(button_order chan elevio.ButtonEvent, floor chan int, obstruction chan bool, stopped chan bool) {

// 	for {
// 		select {
// 		case bo := <-button_order:
// 			// Gett floor and button data
// 			floor, btn := e.ReadOrder(bo)

// 			// Ready order for handling
// 			e.AddOrders(floor, btn)

// 		case cf := <-floor:
// 			order_mutex.Lock()
// 			*e.current_floor = cf
// 			order_mutex.Unlock()
// 			// fmt.Printf("\n From channel: %t \n", cf)

// 		case io := <-obstruction:
// 			order_mutex.Lock()
// 			*e.is_obstructed = io
// 			order_mutex.Unlock()

// 		case is := <-stopped:
// 			order_mutex.Lock()
// 			*e.is_stopped = is
// 			order_mutex.Unlock()
// 			// fmt.Printf("Stopping: %t\n", *e.is_stopped)
// 		}
// 	}
// }

// // Adds orders to the button arrays
// func (e ElevatorScheduler) AddOrders(floor int, button elevio.ButtonType) {
// 	// Set elevator lights
// 	elevio.SetButtonLamp(button, floor, true)

// 	order_mutex.Lock()
// 	switch button {
// 	case 0:
// 		e.up_button_array[floor] = true
// 	case 1:
// 		e.down_button_array[floor] = true
// 	case 2:
// 		e.internal_button_array[floor] = true
// 	}
// 	order_mutex.Unlock()
// }

// // Convert order to readable format
// func (e ElevatorScheduler) ReadOrder(button_order elevio.ButtonEvent) (floor int, button elevio.ButtonType) {
// 	order_floor := button_order.Floor
// 	order_button := button_order.Button

// 	return order_floor, order_button
// }

func assembleArgument(systemData SystemData) {
	
}

func runCommandLine() {
	command := "/home/sindre/coding/Sanntid/Project-resources/cost_fns/hall_request_assigner/hall_request_assigner"
	arg := `{
	"hallRequests" : 
		[[false,false],[true,false],[false,false],[false,true]],
	"states" : {
		"one" : {
			"behaviour":"moving",
			"floor":2,
			"direction":"up",
			"cabRequests":[false,false,true,true]
		},
		"two" : {
			"behaviour":"idle",
			"floor":0,
			"direction":"stop",
			"cabRequests":[false,false,false,false]
		}
	}
}`
	cmd := exec.Command(command, "-i", arg)
	stdout, err := cmd.Output()
	fmt.Printf(arg + "\n")

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(stdout ), &jsonMap)

	fmt.Println(jsonMap["one"])    
}

func main() {

	// elevio.Init("localhost:15657", N_FLOORS)

	// // Initialize the channels for receiving data from the elevio interface
	// drv_buttons := make(chan elevio.ButtonEvent)
	// drv_floors := make(chan int)
	// drv_obstr := make(chan bool)
	// drv_stop := make(chan bool)

	// go elevio.PollButtons(drv_buttons)
	// go elevio.PollFloorSensor(drv_floors)
	// go elevio.PollObstructionSwitch(drv_obstr)
	// go elevio.PollStopButton(drv_stop)
	runCommandLine()
}
