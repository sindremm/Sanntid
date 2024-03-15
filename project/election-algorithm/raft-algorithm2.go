package election_algorithm

import (
	//"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"
	bcast "elevator/network/bcast"
	"elevator/network/peers"
	elev_structs "elevator/structs"
	tcp_interface "elevator/tcp-interface"
)

//TODO: endre navnet fra raft-agortihm2 til election_algorithm

//TODO: Find a more appropiate name for struct
type ElectionHandler struct {
	Id              int
	Leader          bool
	LeaderID        int
	LastHeartbeat   time.Time
	Elevators       []*ElectionHandler
	Acknowledgement chan bool
	VotesReceived   int
	Term int
	VotedInTerm int
	IsElectionStarter bool
	PORT int
	Address string

	//TCP connection
	Conn   net.Conn
	ID     string
	Timestamp int64
	Master bool
}

// VoteMsg represents a vote message
type VoteMsg struct {
    ID string
    Term int
    Vote int
}

type State struct {
	//Last ID of the leader
	LeaderID int `json:"leader_id"`
	//Timestamp of the last heartbeat
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

func NewElevator(id int, conn net.Conn) *ElectionHandler {
	return &ElectionHandler{
		Id:              id,
		Leader:          false,
		Elevators:       make([]*ElectionHandler, 0),
		Acknowledgement: make(chan bool),
		Conn:            conn,
	}
}


func (e *ElectionHandler) BroadcastID() {
	// Create a WaitGroup to wait for all goroutines to finish.
	var wg sync.WaitGroup
	// Create a channel to signal when all acknowledgments are received.
	allAckReceived := make(chan struct{})
	message := &elev_structs.SystemData{ID: e.Id}
	
	// Iterate over all elevators.
	for _, elevator := range e.Elevators {
		// Increment the WaitGroup counter.
		wg.Add(1)
		// Start a new goroutine for each elevator.
		go func(elevator *ElectionHandler) {
			// Decrement the WaitGroup counter when the goroutine completes.
			defer wg.Done()

			// Simulate sending the message over TCP to the current elevator.
			// In a real implementation, replace this with actual message sending logic.
			fmt.Printf("Sending message to elevator %d\n", elevator.Id)

			// Create a new message with the elevator's ID.
			tcp_interface.SendSystemData(elevator.Address, message)

			// Simulate receiving acknowledgment from the current elevator.
			// In a real implementation, replace this with acknowledgment logic.
			// Here, we assume that the acknowledgment is received immediately after sending the message.
			fmt.Printf("Received acknowledgment from elevator %d\n", elevator.Id)

			// Send true on the Acknowledgement channel.
			elevator.Acknowledgement <- true
		}(elevator)
	}

	// Start a goroutine to wait for all acknowledgments.
	go func() {
		// Wait for all goroutines to finish.
		wg.Wait()
		// Signal that all acknowledgments are received.
		close(allAckReceived)
	}()

	// Wait for all acknowledgments to be received or for a timeout.
	select {
	case <-allAckReceived:
		fmt.Println("All acknowledgments received")
	case <-time.After(10 * time.Second): // Adjust timeout duration as needed
		fmt.Println("Timeout: Not all acknowledgments received")
		// If acknowledgments have not been received from all elevators, handle the situation accordingly.
		// For example, you could start a new election here.
	}
}


func (e *ElectionHandler) ReceiveBroadcast(id int) {
	if id > e.Id {
		e.Acknowledgement <- true
	}
}


// SendVote sends a vote to all other elevators.
func (e *ElectionHandler) SendVote(vote int) {
	// Only send a vote if the elevator has not already voted in the current term.
	if e.VotedInTerm != e.Term {
		// Create a new vote message with the elevator's ID, the current term, and the vote.
		voteMsg := VoteMsg{ID: e.ID, Term: e.Term, Vote: vote}

		// Loop over all elevators in the system.
		for _, elevator := range e.Elevators {
			// If the current elevator is not the one sending the vote...
			if elevator.ID != e.ID {
				// Send the vote message to the current elevator.
				go func(elevator *ElectionHandler) {
					elevator.ReceiveVote(voteMsg)
				}(elevator)
			}
		}

		// Log that the vote has been sent.
		log.Printf("Sent vote: %d\n", vote)
		
		// Update the term in which the elevator last voted.
		e.VotedInTerm = e.Term
	}
}

// ReceiveVote receives a vote message from another elevator.
func (e *ElectionHandler) ReceiveVote(voteMsg VoteMsg) {
	// Process the received vote.
	fmt.Printf("Received vote for term %d\n", voteMsg.Term)
	if voteMsg.Term > e.Term {
		e.Term = voteMsg.Term
		e.Leader = false
	} else if voteMsg.Term == e.Term {
		e.VotesReceived++
	}
}

// ReceiveVotes receives votes from other elevators.
func (e *ElectionHandler) ReceiveVotes() {
	// If this elevator did not start the election, return immediately.
	if !e.IsElectionStarter {
		return
	}

	// Create a new channel to receive vote messages.
	voteChan := make(chan VoteMsg)
	
	// Start receiving vote messages on the specified port and send them to the voteChan.
	go bcast.Receiver(e.PORT, voteChan)

	// Create a timer for the vote receiving process.
	// After the specified duration, the timer will send a message on its channel.
	voteTimeout := time.NewTimer(5 * time.Second) // adjust the duration as needed

	// Start an infinite loop to continuously receive votes.
	for {
		select {
		// When a vote is received on the voteChan...
		case voteMsg := <-voteChan:
			// Process the vote.
			e.ReceiveVote(voteMsg)
		// When the timer expires...
		case <-voteTimeout.C:
			// Stop the vote receiving process and return.
			fmt.Println("Vote receiving timed out")
			return
		}
	}
}

func (e *ElectionHandler) CheckAcknowledgement() {
	// Create a WaitGroup to wait for all goroutines to finish.
	var wg sync.WaitGroup
	// Create a channel to signal when all acknowledgments are received.
	allAckReceived := make(chan struct{})

	// Iterate over all elevators.
	for _, elevator := range e.Elevators {
		// Increment the WaitGroup counter.
		wg.Add(1)
		// Start a new goroutine for each elevator.
		go func(elevator *ElectionHandler) {
			// Decrement the WaitGroup counter when the goroutine completes.
			defer wg.Done()

			// Create a SystemData object with the elevator's ID.
			systemData := &elev_structs.SystemData{ID: e.Id}

			// Simulate sending the elevator's ID over TCP to the current elevator.
			// In a real implementation, replace this with actual TCP communication.
			fmt.Printf("Sending ID %d to elevator %d over TCP\n", e.Id, elevator.Id)

			// Update the code to use the Address field.
			tcp_interface.SendSystemData(elevator.Address, systemData)

			// Simulate receiving acknowledgment from the current elevator.
			// In a real implementation, replace this with acknowledgment logic.
			// Here, we assume that the acknowledgment is received immediately after sending the ID.
			fmt.Printf("Received acknowledgment from elevator %d\n", elevator.Id)

			// Send true on the Acknowledgement channel.
			elevator.Acknowledgement <- true
		}(elevator)
	}

	// Start a goroutine to wait for all acknowledgments.
	go func() {
		// Wait for all goroutines to finish.
		wg.Wait()
		// Signal that all acknowledgments are received.
		close(allAckReceived)
	}()

	// Wait for all acknowledgments to be received or for a timeout.
	select {
	case <-allAckReceived:
		fmt.Println("All acknowledgments received")
	case <-time.After(10 * time.Second): // Adjust timeout duration as needed
		fmt.Println("Timeout: Not all acknowledgments received")
		// If acknowledgments have not been received from all elevators, handle the situation accordingly.
		// For example, you could start a new election here.
	}
}

//TODO: Er denne viktig?
// func (e *ElectionHandler) AddElevator(addr string) {
// 	//connect to the other elevator
// 	conn, err := net.Dial("tcp", addr)
// 	if err != nil {
// 		fmt.Println("Failed to connect to elevator:", err)
// 		return
// 	}

// 	e.Conn = conn
// }



func (e *ElectionHandler) StartElection() {
	//TODO: Gi dem et fint hjem
	//Ports for checking for life
	broadcast_port := 33344
	peers_port := 33224

	e.IsElectionStarter = true
	// Increment the term number when starting a new election.
    e.Term++

	// Add a random delay before starting the election to reduce the chance of simultaneous broadcasts.
	// The delay is a random duration between 0 and 1000 milliseconds.
	delay := time.Duration(rand.Intn(1000)) * time.Millisecond

	// Pause the execution of the current goroutine for the duration of the delay.
	time.Sleep(delay)

	// Broadcast the elevator's ID to all other elevators. FUNKER
	e.BroadcastID()

	//Receive Broadcast
	e.ReceiveBroadcast(e.Id)

	// Send a vote for itself to all other elevators.
	e.SendVote(e.Id)

	// Receive votes from other elevators. If a higher vote is received, this elevator is not the leader.
	e.ReceiveVotes()

	// If no acknowledgements are received, declare this elevator as the leader.
	if e.Leader {
		fmt.Println("This elevator is the leader.")
		// Start a new goroutine to check the leader's heartbeat.
		go e.CheckHeartbeat(e.ID, peers_port, broadcast_port)
	} else {
		fmt.Println("This elevator is not the leader.")
	}
}

// SaveState saves the current state of the elevator to a JSON file.
// The state includes the ID of the leader and the time of the last heartbeat.
func (e *ElectionHandler) SaveState() {
	// Create a new state object with the current leader ID and last heartbeat time.
	state := State{
		LeaderID:      e.LeaderID,
		LastHeartbeat: e.LastHeartbeat,
	}

	// Create a new file to save the state. The file name includes the ID of the elevator.
	file, err := os.Create(fmt.Sprintf("elevator_%d_state.json", e.Id))
	if err != nil {
		// If there's an error creating the file, log the error and stop the program.
		log.Fatal(err)
	}

	// Ensure the file gets closed once the function finishes.
	defer file.Close()

	// Encode the state object as JSON and save it to the file.
	json.NewEncoder(file).Encode(state)
}

// LoadState loads the state of the elevator from a JSON file.
// The state includes the ID of the leader and the time of the last heartbeat.
func (e *ElectionHandler) LoadState() {
	// Open the file that contains the saved state. The file name includes the ID of the elevator.
	file, err := os.Open(fmt.Sprintf("elevator_%d_state.json", e.Id))
	if err != nil {
		// If there's an error opening the file, log the error and stop the program.
		log.Fatal(err)
	}

	// Ensure the file gets closed once the function finishes.
	defer file.Close()

	// Create a new state object to hold the loaded state.
	state := State{}

	// Decode the JSON from the file into the state object.
	json.NewDecoder(file).Decode(&state)

	// Update the elevator's state with the loaded state.
	e.LeaderID = state.LeaderID
	e.LastHeartbeat = state.LastHeartbeat
}

func (e *ElectionHandler) WriteStateToFile() {
	//Open the file for writing
	file, err := os.OpenFile("elevator_state.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Failed to open file:", err)
		return
	}
	defer file.Close()

	// Write the elevator's state to the file
	_, err = file.WriteString(fmt.Sprintf("Elevator %d: Leader = %v\n", e.ID, e.Leader))
	if err != nil {
		fmt.Println("Failed to write to file:", err)
		return
	}
}

func TestElevator(argsForPorts string) {
	ports := strings.Split(argsForPorts, ",")

	//Initialize the elevators
	elevators := make([]*ElectionHandler, len(ports))
	for i, port := range ports {
		//Connect to the other elevator
		conn, err := net.Dial("tcp", "localhost:"+port)
		if err != nil {
			fmt.Println("Failed to connect to elevator:", err)
			return
		}

		//Create a new elevator
		elevators[i] = NewElevator(i, conn)
	}

	for _, elevator := range elevators {
		// Start election
		elevator.StartElection()
		// Wait for some time to allow for election to complete
		time.Sleep(1 * time.Second)

		if elevator.Leader {
			fmt.Println("Elevator is the leader")
			// Start the Master-Slave algorithm with the leader as the Master
			//StartMasterSlave(elevator)
		} else {
			fmt.Println("Elevator is not the leader")
		}
	}

	// Prevent the program from terminating
	for {
		time.Sleep(time.Minute)
	}
}




