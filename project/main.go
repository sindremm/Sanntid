package main

import (
	// "Network-go/network/bcast"
	// "Network-go/network/localip"
	// "Network-go/network/peers"
	"Driver-go/elevio"
	"time"
	singleelev "elevator/single-elevator"
	master "elevator/master-slave"
	"elevator/structs"
)



func main() {

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

	

	// Create elevator and start main loop
	elevator := singleelev.MakeElevator()

	// Create master slave
	master_slave := master.MakeMasterSlave(1, elevator)

	// Start reading elevator channels
	go elevator.ReadChannels(drv_buttons, drv_floors, drv_obstr, drv_stop)

	// Start master main loop
	go master_slave.MainLoop()

	// Prevent the program from terminating
	for { 
		time.Sleep(time.Minute)
	}

}
