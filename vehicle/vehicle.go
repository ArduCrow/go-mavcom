package vehicle

import (
	"fmt"
	"gomavlink/reader"
)

type Vehicle struct {
	// TODO - Add fields to represent the vehicle's state
	connection *reader.MavlinkReader
}

func NewVehicle(port string, baud int, network bool) (*Vehicle, error) {
	// TODO - Implement the vehicle's main loop
	fmt.Println("New vehicle")
	r, err := reader.NewMavlinkReader("127.0.0.1:14552", 115200, network)
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
