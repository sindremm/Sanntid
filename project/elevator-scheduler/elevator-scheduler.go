// package elevatorscheduler
package main

import (
	// "Driver-go/elevio"
	// "sync"
	"elevator/structs"
	"encoding/json"
	"fmt"
	"os/exec"
	// "io/ioutil"
)

// Create the argument in the correct format for the cost function
func assembleArgument(systemData structs.SystemData) MessageStruct {

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
	new_states := make(map[string]singleState)
	for i := 0; i < len(*systemData.ELEVATOR_STATES); i++ {

		new_state := singleState{}

		state := (*systemData.ELEVATOR_STATES)[i]

		new_state.Behaviour = state_to_behaviour(state)
		new_state.Floor = state.CURRENT_FLOOR
		new_state.Direction = direction_string[state.DIRECTION]
		new_state.CabRequests = systemData.INTERNAL_BUTTON_ARRAY[i]

		// Set the values for the corresponding elevator
		if i == 0 {
			new_states["one"] = new_state
		} else if i == 1 {
			new_states["two"] = new_state
		} else if i == 2 {
			new_states["three"] = new_state
		}
	}

	// Set the new states
	new_argument.States = new_states

	return new_argument
}

// Translate the elevators state to the corresponding string value
func state_to_behaviour(state structs.ElevatorState) string {
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
// type states struct {
// 	One   singleState `json:"one"`
// 	Two   singleState `json:"two"`
// 	Three singleState `json:"Three"`
// }

// Structure for the full message
type MessageStruct struct {
	HallRequests [4][2]bool             `json:"hallRequests"`
	States       map[string]singleState `json:"states"`
}

// Return the movements of the elevator
func CalculateElevatorMovement(systemData structs.SystemData) *(map[string][][2]bool) {
	command := "./hall_request_assigner"

	// Create json string from the system data
	new_struct := assembleArgument(systemData)
	new_json, err := json.MarshalIndent(new_struct, "", "\t")

	if err != nil {
		fmt.Printf("%s", err)
		//TODO: create error output
		return new(map[string][][2]bool)
	}

	// Run the cost function to get new orders
	cmd := exec.Command(command, "-i", string(new_json))
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		//TODO: create error output
		return new(map[string][][2]bool)
	}

	// Decode the new orders
	output := new(map[string][][2]bool)
	json.Unmarshal([]byte(stdout), &output)

	return output
}

func main() {

	states := []structs.ElevatorState{
		{
			ACTIVE:         true,
			CURRENT_FLOOR:  2,
			TARGET_FLOOR:   2,
			DIRECTION:      1,
			INTERNAL_STATE: 1,
		},
		{
			ACTIVE:         true,
			CURRENT_FLOOR:  0,
			TARGET_FLOOR:   2,
			DIRECTION:      0,
			INTERNAL_STATE: 0,
		},
		{
			ACTIVE:         true,
			CURRENT_FLOOR:  0,
			TARGET_FLOOR:   2,
			DIRECTION:      0,
			INTERNAL_STATE: 1,
		},
	}

	data := structs.SystemData{
		SENDER:            0,
		UP_BUTTON_ARRAY:   &([structs.N_FLOORS]bool{false, true, false, false}),
		DOWN_BUTTON_ARRAY: &([structs.N_FLOORS]bool{false, false, false, true}),
		INTERNAL_BUTTON_ARRAY: &([structs.N_ELEVATORS][structs.N_FLOORS]bool{
			{false, false, true, true},
			{false, false, false, false},
			{false, false, true, true},
		}),
		WORKING_ELEVATORS: &([structs.N_FLOORS]bool{false, false, true, true}),
		ELEVATOR_STATES:   &(states),
		COUNTER:           0,
	}

	return_msg := CalculateElevatorMovement(data)

	fmt.Printf("%v", (*return_msg))

}
