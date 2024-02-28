package main

import (
	// "Network-go/network/bcast"
	// "Network-go/network/localip"
	// "Network-go/network/peers"
	"Driver-go/elevio"
	"singleElevator/singleelev"
	"time"
)



func main() {
	var numFloors int = 4

	elevio.Init("localhost:15657", numFloors)

	singleelev.resetElevator()

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
	main_elevator := singleelev.makeElevator()

	// Start threads
	go main_elevator.readChannels(drv_buttons, drv_floors, drv_obstr, drv_stop)
	go main_elevator.Main()

	// Prevent the program from terminating
	for { 
		time.Sleep(time.Minute)
	}

}
