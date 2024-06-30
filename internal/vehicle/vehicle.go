package vehicle

import (
	"fmt"
	"gomavlink/internal/mavlink"
	"gomavlink/internal/reader"
	"sync"
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

type Battery struct {
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
	connected  bool
	connection *reader.MavlinkReader
	Airframe   Airframe
	Battery    Battery
	Position   Position
	lock       sync.Mutex
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

// Begins the vehicles main loop
// Spawns a goroutine to listen for messages from the connection
// since reading from a channel is blocking
// Starts the mavlink connection once this goroutine is running
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
	// prevent race conditions/multiple subprocecces from updating the states
	v.lock.Lock()
	defer v.lock.Unlock()

	if msg.GetMessageID() == 0 && !v.connected {
		v.processInitialHeartbeat(msg)
	} else if v.connected {
		// Process other messages only if connected is true
		switch msg.GetMessageID() {
		case 0:
			// Heartbeat
		case 33:
			// GlobalPositionInt
			v.connection.CurrentStates.GlobalPositionIntState = msg.MessageData()
		case 74:
			// VFR_HUD
			v.connection.CurrentStates.VFRHUDState = msg.MessageData()
		default:
			fmt.Println("Unknown message ID: ", msg.GetMessageID())
		}
	}

	if v.connected && v.connection.CurrentStates.GlobalPositionIntState != nil {
		v.updatePosition()
	}
}

func (v *Vehicle) processInitialHeartbeat(msg mavlink.DecodedMessage) {
	rawAirframe := msg.MessageData()["Type"].(float64)
	intAirframe := int(rawAirframe)
	v.Airframe = Airframe(intAirframe)
	fmt.Println("Airframe: ", v.Airframe.String())
	v.connected = true
}

func (v *Vehicle) updatePosition() {
	v.Position = Position{
		Latitude:         v.connection.CurrentStates.GlobalPositionIntState["Lat"].(float64),
		Longitude:        v.connection.CurrentStates.GlobalPositionIntState["Lon"].(float64),
		AltitudeRelative: v.connection.CurrentStates.GlobalPositionIntState["RelativeAlt"].(float64),
		AltitudeAMSL:     v.connection.CurrentStates.GlobalPositionIntState["Alt"].(float64),
		Heading:          v.connection.CurrentStates.GlobalPositionIntState["Hdg"].(float64),
	}

	fmt.Println("Position state update: ", v.Position)
}

// func (v *Vehicle) updateBatteryState(voltage float64, current float64) {
// 	v.Battery = Battery{
// 		Voltage: voltage,
// 		Current: current,
// 	}
// }
