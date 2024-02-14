package main

import (
	// "Network-go/network/bcast"
	// "Network-go/network/localip"
	// "Network-go/network/peers"
	"Driver-go/elevio"
	// "flag"
	"fmt"
	// "os"
	"time"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//
//	will be received as zero-values.
type HelloMsg struct {
	Message string
	Iter    int
}

/*
// func main() {
// 	// Our id can be anything. Here we pass it on the command line, using
// 	//  `go run main.go -id=our_id`
// 	var id string
// 	flag.StringVar(&id, "id", "", "id of this peer")
// 	flag.Parse()

// 	// ... or alternatively, we can use the local IP address.
// 	// (But since we can run multiple programs on the same PC, we also append the
// 	//  process ID)
// 	if id == "" {
// 		localIP, err := localip.LocalIP()
// 		if err != nil {
// 			fmt.Println(err)
// 			localIP = "DISCONNECTED"
// 		}
// 		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
// 	}

// 	// We make a channel for receiving updates on the id's of the peers that are
// 	//  alive on the network
// 	peerUpdateCh := make(chan peers.PeerUpdate)
// 	// We can disable/enable the transmitter after it has been started.
// 	// This could be used to signal that we are somehow "unavailable".
// 	peerTxEnable := make(chan bool)
// 	go peers.Transmitter(15647, id, peerTxEnable)
// 	go peers.Receiver(15647, peerUpdateCh)

// 	// We make channels for sending and receiving our custom data types
// 	helloTx := make(chan HelloMsg)
// 	helloRx := make(chan HelloMsg)
// 	// ... and start the transmitter/receiver pair on some port
// 	// These functions can take any number of channels! It is also possible to
// 	//  start multiple transmitters/receivers on the same port.
// 	go bcast.Transmitter(16569, helloTx)
// 	go bcast.Receiver(16569, helloRx)

// 	// The example message. We just send one of these every second.
// 	go func() {
// 		helloMsg := HelloMsg{"Hello from " + id, 0}
// 		for {
// 			helloMsg.Iter++
// 			helloTx <- helloMsg
// 			time.Sleep(1 * time.Second)
// 		}
// 	}()

// 	fmt.Println("Started")
// 	for {
// 		select {
// 		case p := <-peerUpdateCh:
// 			fmt.Printf("Peer update:\n")
// 			fmt.Printf("  Peers:    %q\n", p.Peers)
// 			fmt.Printf("  New:      %q\n", p.New)
// 			fmt.Printf("  Lost:     %q\n", p.Lost)

// 		case a := <-helloRx:
// 			fmt.Printf("Received: %#v\n", a)
// 		}
// 	}
// }
*/

var numFloors int = 4

type State int
const (
	IDLE State = iota
	MOVING
    STOPPED
    AT_FLOOR
    DOOR_OPEN
)

type Direction int
const (
    UP Direction = iota
    DOWN 
    STILL
)

type Elevator struct {  
    // The buffer values received from the elevio interface 
    button_order elevio.ButtonEvent
    current_floor int
    is_obstructed bool
    is_stopped bool

    // Arrays that show awhich buttons have been pressed
    up_button_array [4]bool
    down_button_array [4]bool
    internal_button_array [4]bool

    // Variable containing the current state
    internal_state State

    // Variable showing the last visited floor
    at_floor int

    // The current target of the elevator (-1 for no target)
    target_floor int

    // Variable for the direction of the elevator
    moving_direction Direction

    // Variable for keeping track of when interrupt ends
    interrupt_end time.Time;
}


func makeElevator() (Elevator) {
    // Set state to idle
    var start_state State = IDLE;

    
    // Exception value
    starting_floor:= -1

    
    // Initialize empty button arrays
    up_array := [4]bool{};
    down_array := [4]bool{};
    internal_array := [4]bool{};

    return Elevator{
        elevio.ButtonEvent{},
        -1,
        false,
        false,
        up_array,
        down_array,
        internal_array,
        start_state,
        starting_floor,
        -1,
        STILL,
        time.Now()}

}

func (e Elevator) Main() {

    // TODO: Write state machine

    fmt.Printf("%s", e.internal_state)
    for {      
        // Check for stop-button press

        if e.is_stopped {
            fmt.Print("Stop")
            e.Stop()
            continue
        }

        switch state := e.internal_state; state {
        case IDLE:
            fmt.Printf("Idle")
            e.pickFloor()

        case MOVING:
            fmt.Printf("Moving")
            // Handle orders when at floor
            
            if e.current_floor != 1 {
                e.at_floor = e.current_floor
                e.visit_floor()
            }

        case DOOR_OPEN:
            fmt.Printf("open door")
            e.OpenDoor()
        }
    }
}

func (e Elevator) readChannels(button_order chan elevio.ButtonEvent, current_floor chan int, is_obstructed chan bool, is_stopped chan bool) {
    // Read from the channels and put data into variables
    for {
        select {
        case bo := <-button_order:
            // Transform order to readable format
            floor, btn := e.readOrder(bo)
            // Add order to internal array and set lights
            e.addOrders(floor, btn)

        case cf := <-current_floor:
            e.current_floor = cf

        case io := <-is_obstructed:
            e.is_obstructed = io

        case is := <-is_stopped:
            e.is_stopped = is
        }
    }
}

func (e Elevator) pickFloor() {
    // Sets new target to closest floor, prioritizing floors above
    new_target := -1

    // TODO: Add check to see if there are new orders instead of running this loop every time

    // This code can be reworked to better adhere to the DRY-principle
    for i := 1; i < 4; i++ {

        // Check floors above
        check_floor := e.at_floor + i

        if check_floor < 0 || check_floor > 4 {
            continue
        }

        if e.up_button_array[check_floor] || e.down_button_array[check_floor]{
            new_target = check_floor
            e.internal_state = MOVING
            break
        }

        // Check floors below
        check_floor = e.at_floor - i

        if check_floor < 0 || check_floor > 4 {
            continue
        }

        if e.up_button_array[check_floor] || e.down_button_array[check_floor]{
            new_target = check_floor
            e.internal_state = MOVING
            break
        }
    }

    e.target_floor = new_target
}

func (e Elevator) addOrders(floor int, button elevio.ButtonType) {
    // Set elevator lights
    elevio.SetButtonLamp(button, floor, true);
    switch button {
    case 0:
        e.up_button_array[floor] = true;
    case 1:
        e.down_button_array[floor] = true;
    case 2:
        e.internal_button_array[floor] = true;
    }

}



func (e Elevator) readOrder(button_order elevio.ButtonEvent) (floor int, button elevio.ButtonType){
    order_floor := button_order.Floor;
    order_button := button_order.Button;
    
    return order_floor, order_button
}


func (e Elevator) visit_floor() {


    // Remove internal order when opening door at requested floor, and opens door
    if e.internal_button_array[e.at_floor] {
        e.internal_button_array[e.at_floor] = false;
        e.internal_state = DOOR_OPEN;
    }

    // Reset internal button
    elevio.SetButtonLamp(2, e.at_floor, false)

    // Remove orders in same direction, and sets door to open
    switch dir := e.moving_direction; dir {
    case UP:
        e.internal_state = DOOR_OPEN;
        e.up_button_array[e.at_floor] = false;
        // Reset upwards button
        elevio.SetButtonLamp(0, e.at_floor, false)
    case DOWN:
        e.internal_state = DOOR_OPEN;
        e.down_button_array[e.at_floor] = false;
        // Reset downwards button
        elevio.SetButtonLamp(1, e.at_floor, false)
    }

}





func (e Elevator) OpenDoor() {

    //Runs only if door is not obstructed
    obstruction_check := <-e.is_obstructed

    if !(obstruction_check){
        elevio.SetDoorOpenLamp(true);
        time.Sleep(3*time.Second);
        elevio.SetDoorOpenLamp(false);

        // Makes the elevator idle if it has arrived at the requested floor, and makes it keep moving otherwise
        if e.target_floor == e.at_floor {
            e.internal_state = IDLE
        } else {
            e.internal_state = MOVING
        }

    }

    
}

func (e Elevator)  MoveToOrder() {
    if e.target_floor == -1 {
        return
    }

    e.internal_state = MOVING

    if e.target_floor > e.at_floor {
        e.moving_direction = UP
        elevio.SetMotorDirection(elevio.MD_Up)
    } else {
        e.moving_direction = DOWN
        elevio.SetMotorDirection(elevio.MD_Down)
    }
}

func (e Elevator) Stop() {
    // Handles stopping

    elevator_stop := e.is_stopped;

    if e.internal_state == AT_FLOOR{
        elevio.SetDoorOpenLamp(true)
    }

    e.internal_state = STOPPED
    elevio.SetStopLamp(true)
    elevio.SetMotorDirection(elevio.MD_Stop)

    if !elevator_stop{
        time.Sleep(3*time.Second);
        e.internal_state = IDLE
        elevio.SetStopLamp(false)
        elevio.SetDoorOpenLamp(false)
        fmt.Print(e.internal_state)
    }

}
    

    // Handle the current state
    /*
    switch e.internal_state {
    case MOVING:
        e.internal_state = STOPPED
        elevio.SetMotorDirection(elevio.MD_Stop)
        elevio.SetStopLamp(true)

    case AT_FLOOR:
        e.internal_state = STOPPED
        elevio.SetMotorDirection(elevio.MD_Stop)
        elevio.SetStopLamp(true)
        elevio.SetDoorOpenLamp(true)
    
    case DOOR_OPEN:
        e.internal_state = STOPPED
        elevio.SetStopLamp(true)
    }
    */

    // Reset all order lights
    //resetLights()

    // Reset timer every time button is pressed
    /*
    for {
        if elevator_stop {
            e.interrupt_end = time.Now().Add(3*time.Second);
        }
        if e.interrupt_end.After(time.Now()) {
            break
        }
    }
    */
    
    // Set state to IDLE
    /*
    e.internal_state = IDLE
    elevio.SetStopLamp(false)
    elevio.SetDoorOpenLamp(false)
    */


func resetLights() {
    // Reset all order lights
    for f := 0; f < numFloors; f++ {
        for b := elevio.ButtonType(0); b < 3; b++ {
            elevio.SetButtonLamp(b, f, false)
        }
    }
}


func main() {


    elevio.Init("localhost:15657", numFloors)
    

    // Initialize the channels for receiving data from the elevio interface
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)    
    
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)

    fmt.Print(drv_obstr)
    

    // Create elevator and start main loop
    main_elevator := makeElevator()
    
    // Start threads
    go main_elevator.readChannels(drv_buttons, drv_floors, drv_obstr, drv_stop)
    go main_elevator.Main()
    
    

    for {}
    
    // for {
    //     select {
    //     case a := <- drv_buttons:
    //         fmt.Printf("%+v\n", a)
    //         elevio.SetButtonLamp(a.Button, a.Floor, true)
            
    //     case a := <- drv_floors:
    //         fmt.Printf("%+v\n", a)
    //         if a == numFloors-1 {
    //             d = elevio.MD_Down
    //         } else if a == 0 {
    //             d = elevio.MD_Up
    //         }
    //         elevio.SetMotorDirection(d)
            
            
    //     case a := <- drv_obstr:
    //         fmt.Printf("%+v\n", a)
    //         if a {
    //             elevio.SetMotorDirection(elevio.MD_Stop)
    //         } else {
    //             elevio.SetMotorDirection(d)
    //         }
            
    //     case a := <- drv_stop:
    //         fmt.Printf("%+v\n", a)
    //         for f := 0; f < numFloors; f++ {
    //             for b := elevio.ButtonType(0); b < 3; b++ {
    //                 elevio.SetButtonLamp(b, f, false)
    //             }
    //         }
    //     }
    // }  
}