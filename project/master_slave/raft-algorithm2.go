package master_slave

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

	//TCP connection
	Conn   net.Conn
	ID     int
	Master bool
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

//Map for the elevators
var ElevatorMap = make(map[int]string)

func UpdateElevatorMap(int, string){

}

//Encodes systemData to []byte to be sent by TCP
func EncodeSystemData(s *SystemData) ([]byte, error){
	b, err = json.Marshal(s)
	if err!= nil {
		fmt.Print("Error with Marshal \n")
	}
	return b
}

//Decodes SystemData 
func DecodeSystemData(data []byte, v any) SystemData{
	var systemData SystemData

	err := json.Unmarshal([]byte(data), &systemData)
	if err != nil {
        log.Fatalf("Error with decoding:  %s", err)
    }
	return systemData
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

// SendVote sends a vote to another elevator.
func (e *Elevator) SendVote(vote int) {
    // Only send a vote if the elevator has not already voted in the current term.
    if e.VotedInTerm != e.Term {
        _, err := fmt.Fprintf(e.Conn, "VOTE:%d\n", vote)
        if err != nil {
            log.Printf("Failed to send vote: %v", err)
        }
        fmt.Printf("Sent vote: %d\n", vote) // add this line
        // Update the term in which the elevator last voted.
        e.VotedInTerm = e.Term
    }
}

// ReceiveVotes receives votes from other elevators.
func (e *Elevator) ReceiveVotes() {
	reader := bufio.NewReader(e.Conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Failed to read vote: %v", err)
			return
		}
		fmt.Printf("Received message: %s\n", message) // add this line
		if strings.HasPrefix(message, "VOTE:") {
			vote, err := strconv.Atoi(strings.TrimPrefix(message, "VOTE:"))
			if err != nil {
				log.Printf("Failed to parse vote: %v", err)
				return
			}
			// Only count votes that are for the current term.
			if vote > e.Term {
				e.Term = vote
				e.Leader = false
			} else if vote == e.Term {
				e.VotesReceived++
			}
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
	for _, elevator := range e.Elevators {
		go elevator.ReceiveHeartbeat(e.Id)
	}
}

// ReceiveHeartbeat receives a heartbeat message from another elevator.
func (e *Elevator) ReceiveHeartbeat(id int) {
	if id == e.LeaderID {
		e.LastHeartbeat = time.Now()
	}
}

// CheckHeartbeat checks if a heartbeat has been received from the leader.
func (e *Elevator) CheckHeartbeat() {
	for {
		time.Sleep(100 * time.Millisecond)                // Check every 100 milliseconds
		if time.Since(e.LastHeartbeat) > 30*time.Second { // Timeout after 30 seconds
			fmt.Println("Leader failure detected. Starting new election.")
			e.StartElection()
		}
	}
}

func (e *Elevator) StartElection() {
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
		go e.CheckHeartbeat()
	} else {
		fmt.Println("This elevator is not the leader.")
	}
	fmt.Printf("Hei 6")
}

// SaveState saves the state of the elevator to a file.
func (e *Elevator) SaveState() {
	state := State{
		LeaderID:      e.LeaderID,
		LastHeartbeat: e.LastHeartbeat,
	}
	file, err := os.Create(fmt.Sprintf("elevator_%d_state.json", e.Id))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	json.NewEncoder(file).Encode(state)
}

// LoadState loads the state of the elevator from a file.
func (e *Elevator) LoadState() {
	file, err := os.Open(fmt.Sprintf("elevator_%d_state.json", e.Id))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	state := State{}
	json.NewDecoder(file).Decode(&state)
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


