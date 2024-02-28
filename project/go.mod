module elevator

go 1.16

replace Network-go => ./Network-go

replace Driver-go => ./driver-go

replace singleElevator => ./single-elevator

require (
	Driver-go v0.0.0-00010101000000-000000000000 // indirect
	singleElevator v0.0.0-00010101000000-000000000000 // indirect
)
