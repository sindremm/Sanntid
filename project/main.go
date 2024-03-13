package main

import (
	// "Network-go/network/bcast"
	// "Network-go/network/localip"
	// "Network-go/network/peers"
	"Driver-go/elevio"
	"time"

	// "strings"
	master "elevator/master-slave"
	singleelev "elevator/single-elevator"
	"elevator/structs"
	"flag"
	"strconv"
)

func main() {

	elevio.Init("localhost:15680", structs.N_FLOORS)

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
	master_slave := master.MakeMasterSlave(received_id, ":"+port)
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
