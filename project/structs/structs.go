package structs

import (
	"fmt"
	"time"
)

// ##################### System Variables #####################

const N_FLOORS int = 4
const N_ELEVATORS int = 3

// ##################### Master Slave ##################

// Datastruct containing all the information of the system
type SystemData struct {
	// The elevator sending the message (who is also master)
	MASTER_ID int

	// ALL RECEIVED ORDERS
	UP_BUTTON_ARRAY   *[N_FLOORS]bool
	DOWN_BUTTON_ARRAY *[N_FLOORS]bool

	// POSITION AND TARGET OF EACH ELEVATOR
	ELEVATOR_DATA *[N_ELEVATORS]ElevatorData

	// COUNTER FOR MESSAGE SYNCHRONIZATION
	COUNTER int
}

type ElevatorData struct {
	// Specifies wether the elevator is in working condition
	ALIVE bool
	// The address of the elevator for TCP communication
	ADDRESS string

	// All active cab buttons
	INTERNAL_BUTTON_ARRAY [N_FLOORS]bool
	// TARGETS OF EACH ELEVATOR
	ELEVATOR_TARGETS [N_FLOORS][2]bool
	// State machine state of elevator
	INTERNAL_STATE int
	// The last floor the elevator visited
	CURRENT_FLOOR int
	//TODO: update usage of direction
	DIRECTION Direction // 0 for stop, 1 for up, 2 for down
}

const (
	SERVER_IP_ADDRESS = "127.0.0.1"
	PORT              = "20005"
	FILENAME          = "home/student/Documents/AjananMiaSindre/Sanntid/project/driver-go/master-slave/master-slave.go"
)

// ###################### Single Elevator ##########################
type ElevatorState int

const (
	IDLE ElevatorState = iota
	MOVING
	STOPPED
	DOOR_OPEN
)

type Direction int

const (
	UP Direction = iota
	DOWN
	STILL
)

// ########################### Network ############################

type TCPMsg struct {
	MessageType MessageType
	Sender_id   int    `json:"sender_id"`
	Data        []byte `json:"data"`
}

type AliveMsg struct {
	Message string
	address string
	Iter    int
}
type MessageType int

const (
	NEWCABCALL MessageType = iota
	NEWHALLORDER
	UPDATEELEVATOR
	CLEARHALLORDER
	MASTERMSG
)

type ClearHallorderMsg struct {
	clear_floor int
	clear_direction Direction
}


// Changes timeout time for Dial. 500 milliseconds = 0.5 second
var TCP_timeout = 500 * time.Millisecond

//TODO: remove
// ############################ For testing purposes ##################################

func SystemData_to_string(data SystemData) string {
	end_string := ""

	end_string += fmt.Sprintf("MASTER_ID: %d \n", data.MASTER_ID)

	// ALL RECEIVED ORDERS
	end_string += fmt.Sprintf("UP_BUTTON_ARRAY: %v \n", data.UP_BUTTON_ARRAY)
	end_string += fmt.Sprintf("DOWN_BUTTON_ARRAY: %v \n", data.DOWN_BUTTON_ARRAY)

	for i := 0; i < 3; i++ {
		// POSITION AND TARGET OF EACH ELEVATOR
		end_string += fmt.Sprintf("ELEVATOR_DATA[%d] : { \n", i)
		// Specifies wether the elevator is in working condition
		end_string += fmt.Sprintf("\tALIVE: %t \n", data.ELEVATOR_DATA[i].ALIVE)
		// The address of the elevator for TCP communication
		end_string += fmt.Sprintf("\tADDRESS: %s \n", data.ELEVATOR_DATA[i].ADDRESS)

		// All active cab buttons
		end_string += fmt.Sprintf("\tINTERNAL_BUTTON_ARRAY: %v \n", data.ELEVATOR_DATA[i].INTERNAL_BUTTON_ARRAY)

		// TARGETS OF EACH ELEVATOR
		end_string += fmt.Sprintf("\tELEVATOR_TARGETS: %v \n", data.ELEVATOR_DATA[i].ELEVATOR_TARGETS)
		// State machine state of elevator
		end_string += fmt.Sprintf("\tINTERNAL_STATE: %v \n", data.ELEVATOR_DATA[i].INTERNAL_STATE)
		// The last floor the elevator visited
		end_string += fmt.Sprintf("\tCURRENT_FLOOR: %d \n", data.ELEVATOR_DATA[i].CURRENT_FLOOR)

		//TODO: update usage of direction
		end_string += fmt.Sprintf("\tDIRECTION: %v \n", data.ELEVATOR_DATA[i].DIRECTION)

		end_string += "\n}\n\n"

	}

	// COUNTER FOR MESSAGE SYNCHRONIZATION
	end_string += fmt.Sprintf("COUNTER: %d \n", data.COUNTER)

	return end_string
}
