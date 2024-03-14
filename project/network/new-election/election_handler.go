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

	// Convert the current node's id to an integer
	idInt, err := strconv.Atoi(id)
	if err != nil {
		// If the id is not an integer, print an error message and exit
		fmt.Println("Error: This elevator id is not a int, reboot with proper integer id")
	}

	// // Check if there are no connected peers
	// noConnPeers := len(connectedPeers) == 0
	// if noConnPeers {
	// 	// If there are no connected peers, add the current node's id to the peers slice
	// 	peers = append(peers, currentSystemData)
	// }

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
	fmt.Println("Sorted peers: ", peers)

	// Print the id of the master node (the one with the lowest id)
	fmt.Printf("Elevator %s: Master is elevator %d\n", id, peers[0].ID)

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