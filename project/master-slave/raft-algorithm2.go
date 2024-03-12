package election_algorithm

import (
	//"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	//"strconv"
	"strings"
	"sync"
	"time"
	"elevator/network/bcast"
	"elevator/network/peers"
	"elevator/structs"
	"elevator/network/"
)
//TODO: endre navnet fra raft-agortihm2 til election_algorithm

//TODO: Find a more appropiate name for struct
type Elevator struct {
	Id              int
	Leader          bool
	LeaderID        int
	LastHeartbeat   time.Time
	Elevators       []*Elevator
	Acknowledgement chan bool
	VotesReceived   int
	Term int
	VotedInTerm int
	IsElectionStarter bool
	PORT int

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

func NewElevator(id int, conn net.Conn) *Elevator {
	return &Elevator{
		Id:              id,
		Leader:          false,
		Elevators:       make([]*Elevator, 0),
		Acknowledgement: make(chan bool),
		Conn:            conn,
	}
}


// BroadcastID sends the ID of the current elevator to all other elevators.
func (e *Elevator) BroadcastID() {
	// Create a WaitGroup to wait for all goroutines to finish.
	var wg sync.WaitGroup
	// Iterate over all elevators.
	for _, elevator := range e.Elevators {
		// Increment the WaitGroup counter.
		wg.Add(1)
		// Start a new goroutine for each elevator.
		go func(elevator *Elevator) {
			// Decrement the WaitGroup counter when the goroutine completes.
			defer wg.Done()
			// Use a select statement to implement a timeout.
			select {
			// If no response is received within 5 seconds, assume a network failure.
			case <-time.After(5 * time.Second): // Timeout after 5 seconds
				fmt.Println("Network failure detected. Starting new election.")
				// Start a new election.
				e.StartElection()
			// If a response is received, process it.
			case <-elevator.Acknowledgement:
			}
		}(elevator)
	}
	// Wait for all goroutines to finish.
	wg.Wait()
}

func (e *Elevator) ReceiveBroadcast(id int) {
	if id > e.Id {
		e.Acknowledgement <- true
	}
}


// SendVote sends a vote to all other elevators.
func (e *Elevator) SendVote(vote int) {
	// Only send a vote if the elevator has not already voted in the current term.
	if e.VotedInTerm != e.Term {
		// Create a new vote message with the elevator's ID, the current term, and the vote.
		voteMsg := VoteMsg{ID: e.ID, Term: e.Term, Vote: vote}

		// Define the base port number. Replace with the actual base port number.
		basePort := 15657 

		// Loop over all elevators in the system.
		for i, elevator := range e.Elevators {
			// If the current elevator is not the one sending the vote...
			if elevator.ID != e.ID {
				// Create a new channel for sending vote messages.
				voteChannel := make(chan VoteMsg)
				
				// Start a new goroutine that transmits vote messages to the current elevator.
				// The port number is calculated as the base port number plus the index of the elevator.
				go bcast.Transmitter(basePort + i, voteChannel)
				
				// Send the vote message to the current elevator.
				voteChannel <- voteMsg
			}
		}

		// Log that the vote has been sent.
		log.Printf("Sent vote: %d\n", vote)
		
		// Update the term in which the elevator last voted.
		e.VotedInTerm = e.Term
	}
}

// ReceiveVotes receives votes from other elevators.
func (e *Elevator) ReceiveVotes() {
	// If this elevator did not start the election, return immediately.
	if !e.IsElectionStarter {
		return
	}

	// Create a new channel to receive Elevator objects.
	voteChan := make(chan Elevator)
	
	// Start receiving Elevator objects on the specified port and send them to the voteChan.
	bcast.Receiver(e.PORT, voteChan)

	// Create a timer for the vote receiving process.
	// After the specified duration, the timer will send a message on its channel.
	voteTimeout := time.NewTimer(5 * time.Second) // adjust the duration as needed

	// Start an infinite loop to continuously receive votes.
	for {
		select {
		// When a vote is received on the voteChan...
		case vote := <-voteChan:
			// Process the vote.
			fmt.Printf("Received vote for term %d\n", vote.Term)
			if vote.Term > e.Term {
				e.Term = vote.Term
				e.Leader = false
			} else if vote.Term == e.Term {
				e.VotesReceived++
			}
		// When the timer expires...
		case <-voteTimeout.C:
			// Stop the vote receiving process and return.
			fmt.Println("Vote receiving timed out")
			return
		}
	}
}


// CheckAcknowledgements waits for acknowledgements from other elevators.
func (e *Elevator) CheckAcknowledgements() {
	// Initialize a counter for the number of acknowledgements received.
	ackCount := 0

	// Use a loop to keep checking for acknowledgements.
	for {
		select {
		// If an acknowledgement is received, increment the counter.
		case <-e.Acknowledgement:
			ackCount++
			// If all acknowledgements have been received, return.
			if ackCount == len(e.Elevators) {
				return
			}
		// If no acknowledgements are received within a certain time, declare this elevator as the leader.
		case <-time.After(5 * time.Second):
			if e.VotesReceived > len(e.Elevators)/2 {
				e.Leader = true
			} else {
				e.Leader = false
			}
			return // Add this line to exit the function after the timeout
		}
	}
}

func (e *Elevator) AddElevator(addr string) {
	//connect to the other elevator
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Failed to connect to elevator:", err)
		return
	}

	e.Conn = conn
}

// Heartbeat sends a heartbeat message to all other elevators.
func (e *Elevator) Heartbeat() {
	heartbeatChannel := make(chan Elevator)

	go bcast.Transmitter(12345, heartbeatChannel)

	for _, elevator := range e.Elevators {
		heartbeatChannel <- Elevator{ID: e.ID, Timestamp: time.Now().UnixNano()}
		go elevator.ReceiveHeartbeat(e.Id)
	}
}

// ReceiveHeartbeat receives heartbeat messages on the given port
func (e *Elevator) ReceiveHeartbeat(port int) {
	heartbeatChannel := make(chan Elevator)

	go bcast.Receiver(port, heartbeatChannel)

	for {
		select {
		case hb := <-heartbeatChannel:
			fmt.Printf("Received heartbeat from %s at %d\n", hb.ID, hb.Timestamp)
		}
	}
}



// CheckHeartbeat checks if a heartbeat has been received from the leader.
func (e *Elevator) CheckHeartbeat(id string, peers_port int, broadcast_port int) {
	peers_update_channel := make(chan peers.PeerUpdate)
	go peers.Receiver(peers_port, peers_update_channel)

	aliveCheck := make(chan structs.AliveMsg)
	go bcast.Receiver(broadcast_port, aliveCheck)

	for {
		time.Sleep(100 * time.Millisecond) // Check every 100 milliseconds
		select {
		case p := <-peers_update_channel:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			//Updates the ElevatorMap each time a new peer appears
			if p.New != nil{
				UpdateElevatorMap(p.New)
			}
		case a := <-aliveCheck:
			fmt.Printf("Received %#v \n", a)
		default:
			if time.Since(e.LastHeartbeat) > 30*time.Second { // Timeout after 30 seconds
				fmt.Println("Leader failure detected. Starting new election.")
				e.StartElection()
			}
		}
	}
}


func (e *Elevator) StartElection() {
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

	// Send a vote for itself to all other elevators.
	e.SendVote(e.Id)
	fmt.Printf("Hei 4")

	// Check for acknowledgements from other elevators. If no acknowledgements are received,this elevator declares itself as the leader.
	e.CheckAcknowledgements()
	fmt.Printf("Hei 5")

	// Receive votes from other elevators. If a higher vote is received, this elevator is not the leader.
	e.ReceiveVotes()
	fmt.Printf("Hei 6")

	// If no acknowledgements are received, declare this elevator as the leader.
	if e.Leader {
		fmt.Println("This elevator is the leader.")
		// Start a new goroutine to check the leader's heartbeat.
		go e.CheckHeartbeat(e.ID, peers_port, broadcast_port)
	} else {
		fmt.Println("This elevator is not the leader.")
	}
	fmt.Printf("Hei 6")
}

// SaveState saves the current state of the elevator to a JSON file.
// The state includes the ID of the leader and the time of the last heartbeat.
func (e *Elevator) SaveState() {
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
func (e *Elevator) LoadState() {
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

func (e *Elevator) WriteStateToFile() {
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
	elevators := make([]*Elevator, len(ports))
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
			// Assuming StartMasterSlave is defined elsewhere in your code
			StartMasterSlave(elevator)
		} else {
			fmt.Println("Elevator is not the leader")
		}
	}

	// Prevent the program from terminating
	for {
		time.Sleep(time.Minute)
	}
}


