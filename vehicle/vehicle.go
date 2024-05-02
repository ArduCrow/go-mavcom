package vehicle

import (
	"fmt"
	"gomavlink/reader"
)

type Airframe int

const (
	FixedWing  Airframe = 1
	Quadcopter Airframe = 2
)

func (a Airframe) String() string {
	switch a {
	case FixedWing:
		return "FIXED WING"
	case Quadcopter:
		return "QUADCOPTER"
	default:
		return "UNKNOWN"
	}

}

type Vehicle struct {
	// TODO - Add fields to represent the vehicle's state
	connection *reader.MavlinkReader
	Airframe   Airframe
	// TODO - make the reader a field of the vehicle, make the reader update the vehicle's state as updated messages are received
}

func NewVehicle(port string, baud int, network bool) (*Vehicle, error) {
	// TODO - Implement the vehicle's main loop
	fmt.Println("New vehicle")
	r, err := reader.NewMavlinkReader(port, baud, network)
	if err != nil {
		fmt.Println("Error creating Vehicle: ", err)
		return nil, err
	}
	defer r.Close()

	return &Vehicle{connection: r}, nil
}

func (v *Vehicle) Start() {
	fmt.Println("Starting Vehicle...")
	v.connection.Start()
}
