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

	ID int
}

// Data specific to each elevator
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
	INTERNAL_STATE ElevatorState
	// The last floor the elevator visited
	CURRENT_FLOOR int
	// The direction the elevator moves in
	DIRECTION Direction 
}

// ###################### Single Elevator ##########################

// Contains the states used in the elevator state machine
type ElevatorState int
const (
	IDLE ElevatorState = iota
	MOVING
	STOPPED
	DOOR_OPEN
	OBSTRUCTED
)

// Used for indicating direction of elevator and orders
type Direction int
const (
	UP Direction = iota
	DOWN
	STILL
)

// ########################### Network ############################

// Struct used for sending data over TCP
type TCPMsg struct {
	MessageType MessageType
	Sender_id   int    `json:"sender_id"`
	Data        []byte `json:"data"`
}


// Message used to find out which units are alive
type AliveMsg struct {
	Message string
	address string
	Iter    int
}

// Types of messages passed between master and slave
type MessageType int
const (
	NEWCABCALL MessageType = iota
	NEWHALLORDER
	UPDATEELEVATOR
	CLEARHALLORDER
	MASTERMSG
)

type HallorderMsg struct {
	Order_floor     int
	Order_direction [2]bool
}

//Used for specifying the timeout for communication via TCP
var TCP_timeout = 500 * time.Millisecond
