package structs

import (
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

// ###################### Single Elevator ##########################3
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

type AliveMsg struct {
	Message string
	address string
	Iter    int
}

//DELETE SOON
/*
type TestTCPMsg struct {
	SomeMessage string
	TempOrder   int
}
*/

// Changes timeout time for Dial. 500 milliseconds = 0.5 second
var TCP_timeout = 500 * time.Millisecond
