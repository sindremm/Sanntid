package main

import (
	// "Network-go/network/bcast"
	// "Network-go/network/localip"
	// "Network-go/network/peers"
	//"Driver-go/elevio"
	//"time"
	// "strings"
	//singleelev "elevator/single-elevator"
	//master "elevator/master-slave"
	masterselect "elevator/network/new-election"
	elev_structs "elevator/structs"
	"fmt"
	"strconv"
	"time"
)

var addresses [4]string;

// SystemData represents the data of a system in a distributed network.
type SystemData struct {
    ID      int
    COUNTER int
}

func main() {
	// Create three elevators with different IDs and COUNTERs
    elevator1 := elev_structs.SystemData{ID: 1, COUNTER: 3}
    elevator2 := elev_structs.SystemData{ID: 2, COUNTER: 2}
    elevator3 := elev_structs.SystemData{ID: 3, COUNTER: 1}

    // Create a slice of connected peers
    connectedPeers := []elev_structs.SystemData{elevator1, elevator2, elevator3}

    // Create a channel to signal if the current node is the master
    isMaster := make(chan bool, 1)

    // Call DetermineMaster for each elevator
    for _, elevator := range connectedPeers {
        id := strconv.Itoa(elevator.ID)
        currentMasterId := masterselect.DetermineMaster(id, "", connectedPeers, isMaster)
        fmt.Printf("Elevator %s: New master ID is %s\n", id, currentMasterId)
        fmt.Printf("Elevator %s: Is this elevator the master? %v\n", id, <-isMaster)
    }
	fmt.Printf("--------------------------------------------------------------\n")

	time.Sleep(5 * time.Second)
	// Simulate an elevator dying by removing it from the connectedPeers slice
    connectedPeers = connectedPeers[1:]

    fmt.Println("After one elevator dies:")

    // Re-run DetermineMaster for each remaining elevator
    for _, elevator := range connectedPeers {
        id := strconv.Itoa(elevator.ID)
        currentMasterId := masterselect.DetermineMaster(id, "", connectedPeers, isMaster)
        fmt.Printf("Elevator %s: New master ID is %s\n", id, currentMasterId)
        fmt.Printf("Elevator %s: Is this elevator the master? %v\n", id, <-isMaster)
    }
	fmt.Printf("--------------------------------------------------------------\n")

	time.Sleep(5 * time.Second)
    // Simulate an elevator returning by adding it back to the connectedPeers slice
    connectedPeers = append(connectedPeers, elevator1)

    fmt.Println("After one elevator returns:")

    // Re-run DetermineMaster for each elevator
    for _, elevator := range connectedPeers {
        id := strconv.Itoa(elevator.ID)
        currentMasterId := masterselect.DetermineMaster(id, "", connectedPeers, isMaster)
        fmt.Printf("Elevator %s: New master ID is %s\n", id, currentMasterId)
        fmt.Printf("Elevator %s: Is this elevator the master? %v\n", id, <-isMaster)
    }
	fmt.Printf("--------------------------------------------------------------\n")

	// addresses[0] = "ad"
	// addresses[1] = "dsa"
	// addresses[2] = "dsa"
	// addresses[3] = "fdsa"

	// elevio.Init("localhost:15657", structs.N_FLOORS)

	// singleelev.ResetElevator()

	// // Initialize the channels for receiving data from the elevio interface
	// drv_buttons := make(chan elevio.ButtonEvent)
	// drv_floors := make(chan int)
	// drv_obstr := make(chan bool)
	// drv_stop := make(chan bool)

	// go elevio.PollButtons(drv_buttons)
	// go elevio.PollFloorSensor(drv_floors)
	// go elevio.PollObstructionSwitch(drv_obstr)
	// go elevio.PollStopButton(drv_stop)

	// unit_number := 0

	// // Create elevator and start main loop
	// elevator := singleelev.MakeElevator(unit_number)

	// // Specify elevator port
	// port := ":8080"
	// // Create master slave
	// master_slave := master.MakeMasterSlave(unit_number, port, elevator)

	// // Start reading elevator channels
	// go elevator.ReadChannels(drv_floors, drv_obstr, drv_stop)

	// // Start master main loop
	// go master_slave.MainLoop()

	// // Prevent the program from terminating
	// for { 
	// 	time.Sleep(time.Minute)
	// }
	
}
