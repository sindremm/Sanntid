package master_slave

import  (
	"fmt"
	"net"
	"os/exec"
	//"strconv"
	"time"
	"encoding/gob"
)

// Layout of the system data
type SystemData struct{
	// The elevator sending the message (who is also master)
	SENDER int
	
	// ALL RECEIVED ORDERS
	UP_BUTTON_ARRAY 	  *[4]bool
	DOWN_BUTTON_ARRAY     *[4]bool
	INTERNAL_BUTTON_ARRAY *[3][4]bool

	// ALL CURRENTLY WORKING ELEVATORS
	WORKING_ELEVATORS    *[4]bool

	// POSITION AND TARGET OF EACH ELEVATOR
	ELEVATOR_STATES	   *[]ElevatorState

	// COUNTER FOR MESSAGE SYNCHRONIZATION 
	COUNTER int

}


type ElevatorState struct{
	ACTIVE bool
	INTERNAL_STATE int // State machine state of elevator
	CURRENT_FLOOR int
	TARGET_FLOOR int
	//TODO: update usage of direction
	DIRECTION int // 0 for stop, 1 for up, 2 for down
}

const (
	SERVER_IP_ADDRESS = "127.0.0.1"
	PORT = "20005"
	FILENAME = "home/student/Documents/AjananMiaSindre/Sanntid/project/driver-go/master_slave/master_slave.go"
)


type MasterSlave struct {
	CURRENT_DATA *SystemData
	ELEVATOR_NUMBER int

}

func NewMasterSlave() *SystemData {
    return &SystemData{
        SENDER: 0,
        UP_BUTTON_ARRAY: &([4]bool{}),
        DOWN_BUTTON_ARRAY: &([4]bool{}),
        INTERNAL_BUTTON_ARRAY: &([3][4]bool{}),
        WORKING_ELEVATORS: &([4]bool{}),
        ELEVATOR_STATES: &([]ElevatorState{}),
        COUNTER: 0,
    }
}

// HandleOrderFromMaster is a method on the MasterSlave struct that processes an order from the master.
func (ms *MasterSlave) HandleOrderFromMaster(order *ElevatorState) error {
	// Check if the target floor in the order is valid (between 0 and 3)
	if order.TARGET_FLOOR < 0 || order.TARGET_FLOOR > 3 {
		return fmt.Errorf("Invalid order: floor must be between 0 and 3")
	}
	// Check if the direction in the order is valid (-1 for down, 0 for internal, 1 for up)
	if order.DIRECTION < -1 || order.DIRECTION > 1 {
		return fmt.Errorf("Invalid order: direction must be -1, 0 or 1")
	}

	// Update the SystemData based on the order
	// If the direction is 1 (up), set the corresponding floor in the up button array to true
	if order.DIRECTION == 1 {
		ms.CURRENT_DATA.UP_BUTTON_ARRAY[order.TARGET_FLOOR] = true
	// If the direction is -1 (down), set the corresponding floor in the down button array to true
	} else if order.DIRECTION == -1 {
		ms.CURRENT_DATA.DOWN_BUTTON_ARRAY[order.TARGET_FLOOR] = true
	// If the direction is 0 (internal), set the corresponding floor in the internal button array to true
	} else {
		// TODO: Set internal orders for given elevator
		// ms.current_data.INTERNAL_BUTTON_ARRAY[order.TARGET_FLOOR] = true
	}
	// Print a message indicating that the order has been processed
	fmt.Printf("Order for floor %d with direction %d has been processed.\n", order.TARGET_FLOOR, order.DIRECTION)
	return nil
}

func (ms *SystemData) SwitchToBackup() {
	ms.SENDER = 0
	fmt.Println("Master is dead, switching to backup")
}


var fullAddress = SERVER_IP_ADDRESS + ":" + PORT

func StartMasterSlave(leader *Elevator) {
	//Set the leader as the Master
	leader.Master = true

	// start_time := time.Now()
	print_counter := time.Now()
	counter := 0

	ms := &MasterSlave{}


	filename := "/home/student/Documents/AjananMiaSindre/Sanntid/exercise_4/main.go"

	listener, err := net.Listen("tcp", fullAddress)
	if err != nil {
		fmt.Printf("Error creating TCP listener: %v\n", err)
		return
	}
	defer listener.Close()

	exec.Command("gnome-terminal", "--", "go", "run", filename).Run()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting TCP connection: %v\n", err)
			continue
		}

		// Handle incoming TCP connection
		go func(conn net.Conn) {
			defer conn.Close()

			// Receive SystemData from the master
			data := &SystemData{}
			if err := receiveSystemData(conn, data); err != nil {
				fmt.Printf("Error receiving SystemData: %v\n", err)
				return
			}

			// Process received SystemData
			// (Add your logic here based on the received data)

		}(conn)

		// Send SystemData to the master periodically
		if time.Since(print_counter).Seconds() > 1 {
			counter++
			ms.CURRENT_DATA.COUNTER = counter
			if err := sendSystemData(conn, ms.CURRENT_DATA); err != nil {
				fmt.Printf("Error sending SystemData: %v\n", err)
			}

			fmt.Printf("%d\n", counter)
			print_counter = time.Now()
		}
	}
}


// sendSystemData is a function that sends SystemData over a TCP connection.
// It takes a net.Conn object representing the connection and a pointer to the SystemData object to be sent.
// It returns an error if any occurs during the process.
func sendSystemData(conn net.Conn, data *SystemData) error {
	// Create a new encoder that will write to conn
	encoder := gob.NewEncoder(conn)
	// Encode the SystemData object and send it over the connection
	// If an error occurs during encoding, wrap it in a new error indicating that encoding failed
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode SystemData: %v", err)
	}
	// If no error occurred, return nil
	return nil
}

// receiveSystemData is a function that receives SystemData over a TCP connection.
// It takes a net.Conn object representing the connection and a pointer to the SystemData object where the received data will be stored.
// It returns an error if any occurs during the process.
func receiveSystemData(conn net.Conn, data *SystemData) error {
	// Create a new decoder that will read from conn
	decoder := gob.NewDecoder(conn)
	// Decode the received data and store it in the SystemData object
	// If an error occurs during decoding, wrap it in a new error indicating that decoding failed
	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("failed to decode SystemData: %v", err)
	}
	// If no error occurred, return nil
	return nil
}


