package master_slave

import (
    "os"
    "encoding/json"
    "log"
    "time"
    "sync"
    "fmt"
    "math/rand"
)

type Elevator struct {
    Id int
    Leader bool
    LeaderID int
    LastHeartbeat time.Time
    Elevators []*Elevator
    Acknowledgement chan bool
}

type State struct {
    //Last ID of the leader
    LeaderID      int       `json:"leader_id"`
    //Timestamp of the last heartbeat
    LastHeartbeat time.Time `json:"last_heartbeat"`
}

func NewElevator(id int) *Elevator {
    return &Elevator{
        Id: id,
        Leader: false,
        Elevators: make([]*Elevator, 0),
        Acknowledgement: make(chan bool),
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

// CheckAcknowledgements waits for acknowledgements from other elevators.
func (e *Elevator) CheckAcknowledgements() {
    // Sleep for 5 seconds to give other elevators time to send acknowledgements.
    time.Sleep(5 * time.Second)
    // Use a select statement to check if any acknowledgements have been received.
    select {
    // If an acknowledgement is received, return without doing anything.
    case <-e.Acknowledgement:
        return
    // If no acknowledgements are received, declare this elevator as the leader.
    default:
        e.Leader = true
    }
}

func (e *Elevator) AddElevator(elevator *Elevator) {
    e.Elevators = append(e.Elevators, elevator)
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
        time.Sleep(100 * time.Millisecond) // Check every 100 milliseconds
        if time.Since(e.LastHeartbeat) > 30*time.Second { // Timeout after 30 seconds
            fmt.Println("Leader failure detected. Starting new election.")
            e.StartElection()
        }
    }
}

func (e *Elevator) StartElection() {
    // Add a random delay before starting the election to reduce the chance of simultaneous broadcasts.
    // The delay is a random duration between 0 and 1000 milliseconds.
    delay := time.Duration(rand.Intn(1000)) * time.Millisecond
    // Pause the execution of the current goroutine for the duration of the delay.
    time.Sleep(delay)

    // Broadcast the elevator's ID to all other elevators.
    e.BroadcastID()
    // Check for acknowledgements from other elevators. If no acknowledgements are received,
    // this elevator declares itself as the leader.
    e.CheckAcknowledgements()
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

