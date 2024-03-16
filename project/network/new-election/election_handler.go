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
