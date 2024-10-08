package vehicle

import (
	"fmt"
	"gomavlink/internal/communicator"
	"gomavlink/internal/mavlink"
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

type FlightState struct {
	Armed       bool
	ClimbRate   float64
	Airspeed    float64
	Groundspeed float64
	Throttle    float64
}

type Vehicle struct {
	connected   bool
	Connection  *communicator.MavlinkCommunicator
	Airframe    Airframe
	Battery     Battery
	Position    Position
	FlightState FlightState
	lock        sync.Mutex
}

func NewVehicle(port string, baud int, network bool) (*Vehicle, error) {
	// TODO - Implement the vehicle's main loop
	fmt.Println("New vehicle")
	mc, err := communicator.NewMavlinkCommunicator(port, baud, network)
	if err != nil {
		fmt.Println("Error creating Vehicle: ", err)
		return nil, err
	}
	vehicle := &Vehicle{Connection: mc}
	return vehicle, nil
}

// Begins the vehicles main loop
// Spawns a goroutine to listen for messages from the connection
// since reading from a channel is blocking
// Starts the mavlink connection once this goroutine is running
func (v *Vehicle) Start() {
	fmt.Println("Starting Vehicle...")

	go func() {
		for msg := range v.Connection.Messages() {
			v.updateStates(msg)
		}
	}()
	v.Connection.Start()

	// select {}

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
			v.Connection.CurrentStates.Heartbeat = msg.MessageData()
			// fmt.Println("HEARTBEAT SET: ", v.Connection.CurrentStates.Heartbeat)
		case 33:
			// GlobalPositionInt
			v.Connection.CurrentStates.GlobalPositionIntState = msg.MessageData()
		case 74:
			// VFR_HUD
			// fmt.Println("VFR_HUD: ", msg.MessageData())
			v.Connection.CurrentStates.VFRHUDState = msg.MessageData()
			// fmt.Println("VFR_HUD: ", v.Connection.CurrentStates.VFRHUDState)
		default:
			fmt.Println("Unknown message ID: ", msg.GetMessageID())
		}
	}

	if v.connected && v.Connection.CurrentStates.GlobalPositionIntState != nil {
		v.updatePosition()
		if v.Connection.CurrentStates.VFRHUDState != nil && v.Connection.CurrentStates.Heartbeat != nil {
			v.updateFlightState()
		}
		// mavlink.SendMessage(msg, v.Connection.Conn)
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
		Latitude:         v.Connection.CurrentStates.GlobalPositionIntState["Lat"].(float64),
		Longitude:        v.Connection.CurrentStates.GlobalPositionIntState["Lon"].(float64),
		AltitudeRelative: v.Connection.CurrentStates.GlobalPositionIntState["RelativeAlt"].(float64),
		AltitudeAMSL:     v.Connection.CurrentStates.GlobalPositionIntState["Alt"].(float64),
		Heading:          v.Connection.CurrentStates.GlobalPositionIntState["Hdg"].(float64),
	}

	// fmt.Println("Position state update: ", v.Position)
}

func (v *Vehicle) updateFlightState() {
	// fmt.Printf("Flight state: %v\n", v.Connection.CurrentStates.VFRHUDState)
	// fmt.Println("BaseMode: ", v.Connection.CurrentStates.Heartbeat["BaseMode"])
	v.FlightState = FlightState{
		Armed:       v.Connection.CurrentStates.Heartbeat["BaseMode"].(float64) >= 209,
		ClimbRate:   v.Connection.CurrentStates.VFRHUDState["Clb"].(float64),
		Airspeed:    v.Connection.CurrentStates.VFRHUDState["Airspeed"].(float64),
		Groundspeed: v.Connection.CurrentStates.VFRHUDState["Groundspeed"].(float64),
		Throttle:    v.Connection.CurrentStates.VFRHUDState["Throttle"].(float64),
	}
	// fmt.Println("Flight state update: ", v.FlightState)
}

// func (v *Vehicle) updateBatteryState(voltage float64, current float64) {
// 	v.Battery = Battery{
// 		Voltage: voltage,
// 		Current: current,
// 	}
// }
