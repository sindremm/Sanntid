package masterselect

import (
	"fmt"
	"sort"
	"strconv"
	elev_structs "elevator/structs"
	//"elevator/network/peers"
)

// DetermineMaster is a function that determines the master node in a distributed system.
// It takes the current node's id, the current master's id, a list of connected peers, and a channel to signal if the current node is the master.
func DetermineMaster(id string, currentMasterId string, connectedPeers []elev_structs.SystemData, isMaster chan<- bool) string {
	// Initialize an empty slice to hold the ids of all peers
	var peers []elev_structs.SystemData
	fmt.Printf("Connected peers: %v\n", connectedPeers)
	// Convert the current node's id to an integer
	idInt, err := strconv.Atoi(id)
	if err != nil {
		// If the id is not an integer, print an error message and exit
		fmt.Println("Error: This elevator id is not a int, reboot with proper integer id")
	}

	// Iterate over the connected peers
    for _, p := range connectedPeers {
        // Add each peer's SystemData to the peers slice
        peers = append(peers, p)
    }

	// Sort the peers slice first by COUNTER in descending order, then by ID in ascending order
    sort.Slice(peers, func(i, j int) bool {
        if peers[i].COUNTER == peers[j].COUNTER {
            return peers[i].ID < peers[j].ID
        }
        return peers[i].COUNTER > peers[j].COUNTER
    })

	// Print the sorted list of peers
	//fmt.Println("Sorted peers: ", peers)

	// Print the id of the master node (the one with the lowest id)
	//fmt.Printf("Elevator %s: Master is elevator %d\n", id, peers[0].ID)
	fmt.Printf("Peers: %v\n", peers)
	// If the current node's id is the lowest, signal that it is the master
	if peers[0].ID == idInt {
		isMaster <- true
	} else {
		// Otherwise, signal that it is not the master
		isMaster <- false
	}

	// Update the current master's id
	currentMasterId = strconv.Itoa(peers[0].ID)

	// Return the current master's id
	return currentMasterId
}

// //TESTKODE
// // Insert the code under in main, if you want to test the function
// // -------------------------------------------------------------- //
// // Create three elevators with different IDs and COUNTERs
// elevator1 := elev_structs.SystemData{ID: 1, COUNTER: 3}
// elevator2 := elev_structs.SystemData{ID: 2, COUNTER: 2}
// elevator3 := elev_structs.SystemData{ID: 3, COUNTER: 1}

// // Create a slice of connected peers
// connectedPeers := []elev_structs.SystemData{elevator1, elevator2, elevator3}

// // Create a channel to signal if the current node is the master
// isMaster := make(chan bool, 1)

// // Call DetermineMaster for each elevator
// for _, elevator := range connectedPeers {
// 	id := strconv.Itoa(elevator.ID)
// 	currentMasterId := masterselect.DetermineMaster(id, "", connectedPeers, isMaster)
// 	fmt.Printf("Elevator %s: New master ID is %s\n", id, currentMasterId)
// 	fmt.Printf("Elevator %s: Is this elevator the master? %v\n", id, <-isMaster)
// }
// fmt.Printf("--------------------------------------------------------------\n")

// time.Sleep(5 * time.Second)
// // Simulate an elevator dying by removing it from the connectedPeers slice
// connectedPeers = connectedPeers[1:]

// fmt.Println("After one elevator dies:")

// // Re-run DetermineMaster for each remaining elevator
// for _, elevator := range connectedPeers {
// 	id := strconv.Itoa(elevator.ID)
// 	currentMasterId := masterselect.DetermineMaster(id, "", connectedPeers, isMaster)
// 	fmt.Printf("Elevator %s: New master ID is %s\n", id, currentMasterId)
// 	fmt.Printf("Elevator %s: Is this elevator the master? %v\n", id, <-isMaster)
// }
// fmt.Printf("--------------------------------------------------------------\n")

// time.Sleep(5 * time.Second)
// // Simulate an elevator returning by adding it back to the connectedPeers slice
// connectedPeers = append(connectedPeers, elevator1)

// fmt.Println("After one elevator returns:")

// // Re-run DetermineMaster for each elevator
// for _, elevator := range connectedPeers {
// 	id := strconv.Itoa(elevator.ID)
// 	currentMasterId := masterselect.DetermineMaster(id, "", connectedPeers, isMaster)
// 	fmt.Printf("Elevator %s: New master ID is %s\n", id, currentMasterId)
// 	fmt.Printf("Elevator %s: Is this elevator the master? %v\n", id, <-isMaster)
// }
// fmt.Printf("--------------------------------------------------------------\n")