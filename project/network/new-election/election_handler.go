package masterselect

import (
	"fmt"
	"sort"
	"strconv"

	"elevator/network/peers"
)

// DetermineMaster is a function that determines the master node in a distributed system.
// It takes the current node's id, the current master's id, a list of connected peers, and a channel to signal if the current node is the master.
func DetermineMaster(id string, currentMasterId string, connectedPeers []peers.Peer, isMaster chan<- bool) string {
	// Initialize an empty slice to hold the ids of all peers
	var peers []int

	// Convert the current node's id to an integer
	idInt, err := strconv.Atoi(id)
	if err != nil {
		// If the id is not an integer, print an error message and exit
		fmt.Println("Error: This elevator id is not a int, reboot with proper integer id")
	}

	// Check if there are no connected peers
	noConnPeers := len(connectedPeers) == 0
	if noConnPeers {
		// If there are no connected peers, add the current node's id to the peers slice
		peers = append(peers, idInt)
	}

	// Iterate over the connected peers
	for _, p := range connectedPeers {
		// Convert each peer's id to an integer and add it to the peers slice
		pInt, _ := strconv.Atoi(p.Id)
		peers = append(peers, pInt)
	}

	// Sort the peers slice in ascending order
	sort.Ints(peers)

	// Print the sorted list of peers
	fmt.Println("Sorted peers: ", peers)

	// Print the id of the master node (the one with the lowest id)
	fmt.Printf("Elevator %s: Master is elevator %d\n", id, peers[0])

	// If the current node's id is the lowest, signal that it is the master
	if peers[0] == idInt {
		isMaster <- true
	} else {
		// Otherwise, signal that it is not the master
		isMaster <- false
	}

	// Update the current master's id
	currentMasterId = strconv.Itoa(peers[0])

	// Return the current master's id
	return currentMasterId
}

//TESTKODE
// type Peer struct {
//     Id string
// }

// func main() {
//     // Mock data
//     id := "1"
//     currentMasterId := "2"
//     connectedPeers := []peers.Peer{
//         {Id: "3"},
//         {Id: "4"},
//         {Id: "5"},
//     }
//     isMaster := make(chan bool, 1)

//     // Call the function with mock data
//     newMasterId := DetermineMaster(id, currentMasterId, connectedPeers, isMaster)

//     // Print the new master ID
//     fmt.Println("New master ID:", newMasterId)

//     // Print whether this elevator is the master
//     fmt.Println("Is this elevator the master?", <-isMaster)
// }