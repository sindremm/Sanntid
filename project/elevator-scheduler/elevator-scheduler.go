// package elevatorscheduler
package main

import (
	// "Driver-go/elevio"
	// "sync"
	"elevator/master_slave"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"strconv"
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

func assembleArgument(systemData master_slave.SystemData) string {

	arguments := "'{"

	// Add hall requests:
	up := systemData.UP_BUTTON_ARRAY
	down := systemData.DOWN_BUTTON_ARRAY
	requests := make([]string, len(up), len(down))

	for i := 0; i < len(up); i++ {
		element := []string{fmt.Sprintf("%t", up[i]), fmt.Sprintf("%t", down[i])}
		requests[i] = "[" + strings.Join(element, ",") + "]"
	}
	request_string := "[" + strings.Join(requests, ",") + "]"
	arguments += "\n\t\"hallRequests\" :\n\t\t" + request_string + ",\n"

	// Assemble states
	arguments += "\t\"states\" : { \n"
	elev_number := [3]string{"one", "two", "three"}

	for i := 0; i < len(systemData.ELEVATOR_STATES); i++ {
		state := systemData.ELEVATOR_STATES[i]
		if !state.ACTIVE {
			continue
		}
		arguments += "\t\t" + "\"" + elev_number[i] + "\"" + " : {\n\t\t\t"

		arguments += state_to_behaviour(state)
		new_array := []string{}
		for _, el := range systemData.INTERNAL_BUTTON_ARRAY[i] {
			new_array = append(new_array, strconv.FormatBool(el))
		}
		arguments += "\"cabRequests\":" + "[" + strings.Join(new_array, ",") + "]" + "\n\t\t}"
		if i + 2 < len(systemData.ELEVATOR_STATES) {
			arguments += ","
		}
		arguments += "\n"
	}
	arguments += "\n\t}\n}\n'"
	return arguments
}

func state_to_behaviour(state master_slave.ElevatorState) string {
	output := "\"behaviour\":"

	// TODO add the rest

	// Add behaviour to output
	if state.INTERNAL_STATE == 0 {
		output += "\"idle\",\n\t\t\t"
	} else if state.INTERNAL_STATE == 1 {
		output += "\"moving\",\n\t\t\t"
	}

	// Add floor to output
	output += "\"floor\":" + fmt.Sprintf("%v", state.CURRENT_FLOOR) + ",\n\t\t\t"

	// Add direction to output
	output += "\"direction\":"
	if state.Direction == 1 {
		output += "\"up\",\n"
	} else if state.Direction == -1 {
		output += "\"down\",\n"
	} else if state.Direction == 0 {
		output += "\"stop\",\n"
	}
	output += "\t\t\t"

	return output
}

func runCommandLine(systemData master_slave.SystemData) {
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
	// fmt.Printf(arg)
	arg = assembleArgument(systemData)
	// fmt.Printf(arg)

	cmd := exec.Command(command, "-i", arg)
	stdout, err := cmd.Output()
	// fmt.Printf(arg + "\n")

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(stdout), &jsonMap)

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
	states := [3]master_slave.ElevatorState{
		{
			ACTIVE:        true,
			CURRENT_FLOOR: 2,
			TARGET_FLOOR:  2,
			Direction:     1,
			INTERNAL_STATE: 1},
		{
			ACTIVE:        true,
			CURRENT_FLOOR: 0,
			TARGET_FLOOR:  2,
			Direction:     0,
			INTERNAL_STATE: 0},
	}

	data := master_slave.SystemData{
		SENDER:            0,
		UP_BUTTON_ARRAY:   &([4]bool{false, true, false, false}),
		DOWN_BUTTON_ARRAY: &([4]bool{false, false, false, true}),
		INTERNAL_BUTTON_ARRAY: &([3][4]bool{
			{false, false, true, true},
			{false, false, false, false},
			{false, false, true, true},
		}),
		WORKING_ELEVATORS: &([4]bool{false, false, true, true}),
		ELEVATOR_STATES:   &(states),
		COUNTER:           0,
	}

	runCommandLine(data)

}
