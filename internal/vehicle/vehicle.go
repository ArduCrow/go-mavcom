package vehicle

import (
	"fmt"
	"gomavlink/internal/mavlink"
	"gomavlink/internal/reader"
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

type BatteryState struct {
	Voltage float64
	Current float64
}

type Position struct {
	Latitude         float64
	Longitude        float64
	AltitudeRelative float64
	AltitudeAMSL     float64
	Heading          float64
}

type Vehicle struct {
	// TODO - Add fields to represent the vehicle's state
	connection   *reader.MavlinkReader
	Airframe     Airframe
	BatteryState BatteryState
	Position     Position
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

	go func() {
		for msg := range v.connection.Messages() {
			v.updateStates(msg)
		}
	}()
	v.connection.Start()

}

func (v *Vehicle) Travel(lat float64, lon float64, alt int) {
	fmt.Printf("PLACEHOLDER. Vehicle travel to to %v, %v\n", lat, lon)
}

func (v *Vehicle) updateStates(msg mavlink.DecodedMessage) {

	switch msg.GetMessageID() {
	case 0:
		// Heartbeat
		rawAirframe := msg.MessageData()["Type"].(float64)
		intAirframe := int(rawAirframe)
		v.Airframe = Airframe(intAirframe)
		fmt.Println("Airframe: ", v.Airframe.String())
	case 33:
		// GlobalPositionInt
		v.connection.CurrentStates.GlobalPositionIntState = msg.MessageData()
	case 74:
		// VFR_HUD
		v.connection.CurrentStates.VFRHUDState = msg.MessageData()
	default:
		fmt.Println("Unknown message ID: ", msg.GetMessageID())
	}

	// v.updateBatteryState(latestVFRHUDMessageData["BatteryVoltage"].(float64), latestVFRHUDMessageData["BatteryCurrent"].(float64))
	v.updatePosition(
		v.connection.CurrentStates.GlobalPositionIntState["Lat"].(float64),
		v.connection.CurrentStates.GlobalPositionIntState["Lon"].(float64),
		v.connection.CurrentStates.GlobalPositionIntState["RelativeAlt"].(float64),
		v.connection.CurrentStates.GlobalPositionIntState["Alt"].(float64),
		v.connection.CurrentStates.GlobalPositionIntState["Hdg"].(float64),
	)
	fmt.Println("Position state update: ", v.Position)
}

// func (v *Vehicle) updateBatteryState(voltage float64, current float64) {
// 	v.BatteryState = BatteryState{
// 		Voltage: voltage,
// 		Current: current,
// 	}
// }

func (v *Vehicle) updatePosition(lat float64, lon float64, altRel float64, altAMSL float64, heading float64) {
	v.Position = Position{
		Latitude:         lat,
		Longitude:        lon,
		AltitudeRelative: altRel,
		AltitudeAMSL:     altAMSL,
		Heading:          heading,
	}
}
