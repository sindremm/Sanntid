// package elevatorscheduler
package main

import (
	// "Driver-go/elevio"
	// "sync"
	"elevator/master_slave"
	"encoding/json"
	"fmt"
	"os/exec"
	// "io/ioutil"
)

func assembleArgument(systemData master_slave.SystemData) MessageStruct {

	// Create empty struct to store data
	new_argument := MessageStruct{}

	// Add button calls to 2x4 array HallRequests
	up := systemData.UP_BUTTON_ARRAY
	down := systemData.DOWN_BUTTON_ARRAY
	requests := [4][2]bool{}

	for i := 0; i < len(up); i++ {
		requests[i][0] = up[i]
		requests[i][1] = down[i]
	}

	new_argument.HallRequests = requests

	// Assemble states
	direction_string := [3]string{"stop", "up", "down"}
	new_states := states{}
	for i := 0; i < len(systemData.ELEVATOR_STATES); i++ {

		new_state := singleState{}

		state := systemData.ELEVATOR_STATES[i]

		new_state.Behaviour = state_to_behaviour(state)
		new_state.Floor = state.CURRENT_FLOOR
		new_state.Direction = direction_string[state.Direction]
		new_state.CabRequests = systemData.INTERNAL_BUTTON_ARRAY[i]

		// Set the values for the corresponding elevator
		if i == 0 {
			new_states.One = new_state
		} else if i == 1 {
			new_states.Two = new_state
		} else if i == 2 {
			new_states.Three = new_state
		}
	}

	// Set the new states
	new_argument.States = new_states

	return new_argument
}

// Translate the elevators state to the corresponding string value
func state_to_behaviour(state master_slave.ElevatorState) string {
	// TODO: Find correct corresponding states
	if state.INTERNAL_STATE == 0 {
		return "idle"
	}
	if state.INTERNAL_STATE == 1 {
		return "moving"
	}
	if state.INTERNAL_STATE == 2 {
		return "doorOpen"
	}

	fmt.Errorf("Unknown internal state reached")
	return ""
}

// Structure containing the data for each elevator
type singleState struct {
	Behaviour   string  `json:"behaviour"`
	Floor       int     `json:"floor"`
	Direction   string  `json:"direction"`
	CabRequests [4]bool `json:"cabRequests"`
}

// Struct containing the elevators
type states struct {
	One   singleState `json:"one"`
	Two   singleState `json:"two"`
	Three singleState `json:"Three"`
}

// Structure for the full message
type MessageStruct struct {
	HallRequests [4][2]bool `json:"hallRequests"`
	States       states     `json:"states"`
}

func CalculateElevatorMovement(systemData master_slave.SystemData) *(map[string][][2]bool) {
	command := "/home/sindre/coding/Sanntid/Project-resources/cost_fns/hall_request_assigner/hall_request_assigner"

	// Create json string
	new_struct := assembleArgument(systemData)
	new_json, err := json.MarshalIndent(new_struct, "", "\t")

	if err != nil {
		fmt.Printf("%s", err)
		//TODO: create error output
		return new(map[string][][2]bool)
	}

	// Execute command
	cmd := exec.Command(command, "-i", string(new_json))
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		//TODO: create error output
		return new(map[string][][2]bool)
	}
	

	// Decode to struct
	output := new(map[string][][2]bool)
	json.Unmarshal([]byte(stdout), &output)

	return output
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
			ACTIVE:         true,
			CURRENT_FLOOR:  2,
			TARGET_FLOOR:   2,
			Direction:      1,
			INTERNAL_STATE: 1,
		},
		{
			ACTIVE:         true,
			CURRENT_FLOOR:  0,
			TARGET_FLOOR:   2,
			Direction:      0,
			INTERNAL_STATE: 0,
		},
		{
			ACTIVE:         true,
			CURRENT_FLOOR:  0,
			TARGET_FLOOR:   2,
			Direction:      0,
			INTERNAL_STATE: 1,
		},
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

	return_msg := CalculateElevatorMovement(data)
	
	fmt.Printf("%v", (*return_msg))

}
