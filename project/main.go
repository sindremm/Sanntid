package main

import (
	// "Network-go/network/bcast"
	// "Network-go/network/localip"
	// "Network-go/network/peers"
	"Driver-go/elevio"
	"time"
	// "strings"
	singleelev "elevator/single-elevator"
	master "elevator/master-slave"
	"elevator/structs"
	"flag"
	"strconv"
)

var addresses [4]string

func main() {

	addresses[0] = "ad"
	addresses[1] = "dsa"
	addresses[2] = "dsa"
	addresses[3] = "fdsa"

	elevio.Init("localhost:15657", structs.N_FLOORS)

	singleelev.ResetElevator()

	// Initialize the channels for receiving data from the elevio interface
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	//Gets elevator id from terminal
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	received_id, _ := strconv.Atoi(id)

	// Gets port from terminal
	var port string
	flag.StringVar(&port, "port", "", "port of this peer")
	flag.Parse()

	// Create elevator and start main loop
	// Create master slave
	master_slave := master.MakeMasterSlave(received_id, port)
	elevator := singleelev.MakeElevator(received_id, master_slave)
		

	// Start reading elevator channels
	go elevator.ReadChannels(drv_floors, drv_obstr, drv_stop)
	go elevator.ElevatorLoop()

	
	// Start master main loop
	go master_slave.ReadButtons(drv_buttons)
	go master_slave.MainLoop()

	
	

	// Prevent the program from terminating
	for { 
		time.Sleep(time.Minute)
	}

}
