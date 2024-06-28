package reader

import (
	"fmt"
	"io"
	"net"

	"gomavlink/mavlink"

	"github.com/tarm/serial"
)

type MavlinkReader struct {
	serialPort *serial.Port
	conn       net.Conn
	useNetwork bool
}

// Create a new MavlinkReader
func NewMavlinkReader(portName string, baud int, useNetwork bool) (*MavlinkReader, error) {
	var port *serial.Port
	var conn net.Conn
	var err error

	if useNetwork {
		udpAddr, err := net.ResolveUDPAddr("udp", portName)
		if err != nil {
			return nil, err
		}
		conn, err = net.ListenUDP("udp", udpAddr)
		if err != nil {
			return nil, err
		}
	} else {
		port, err = serial.OpenPort(&serial.Config{Name: portName, Baud: baud})
		if err != nil {
			return nil, err
		}
	}

	return &MavlinkReader{serialPort: port, conn: conn, useNetwork: useNetwork}, nil
}

// Read a message from the serial port
func (r *MavlinkReader) readMessage() ([]byte, error) {
	var source io.Reader

	if r.useNetwork {
		source = r.conn
	} else {
		source = r.serialPort
	}

	buf := make([]byte, 1024)
	n, err := source.Read(buf)
	if err != nil {
		return nil, err
	}

	// TODO - Check if the message is a valid mavlink message and parse it

	return buf[:n], nil
}

// Close the connection
func (r *MavlinkReader) Close() error {
	if r.serialPort != nil {
		return r.serialPort.Close()
	}
	return nil
}

// Begin reading messages
func (r *MavlinkReader) Start() {
	var listenPort string
	if r.useNetwork {
		listenPort = r.conn.LocalAddr().String()
	} else {
		listenPort = fmt.Sprintf("%v", r.serialPort)
	}
	fmt.Printf("Starting MavlinkReader, listening on %v\n", listenPort)
	for {
		msg, err := r.readMessage()
		if err != nil {
			fmt.Println("Error reading message: ", err)
			break
		}
		m, err := mavlink.NewRawMessage(msg)
		if err != nil {
			fmt.Println("Error getting message:", err)
		}

		if m.MessageID == 0 {
			fmt.Println("Heartbeat received")
		}
		if m.MessageID == 74 {
			gpi, err := mavlink.DecodeMessage(m)
			if err != nil {
				fmt.Println("Error decoding payload: ", err)
			} else {
				fmt.Println(gpi.GetMessageName())
			}
		}

		// if m.MessageID == 0 {
		// 	fmt.Println("Heartbeat received")
		// 	// fmt.Printf("Length: %d, Sequence: %d, SysID: %d, CompID: %d, MessID: %d, Payload: %v, CRC: %d\n", m.Length, m.Sequence, m.SystemID, m.ComponentID, m.MessageID, m.Payload, m.CRC)

		// 	hbt, err := mavlink.DecodeMessage(m)
		// 	if err != nil {
		// 		fmt.Println("Error decoding payload: ", err)
		// 	} else {
		// 		fmt.Println(hbt.MessageData())

		// 		// two ways to convert from uint8 to float64

		// 		// 1. using reflect
		// 		// valConv := float64(reflect.ValueOf(mData["SystemStatus"]).Uint())
		// 		// fmt.Println(reflect.TypeOf(valConv))

		// 		// 2. using type assertion
		// 		// val := mData["SystemStatus"]
		// 		// valConv := float64(val.(uint8))
		// 		// fmt.Println(valConv)
		// 	}
		// }
		if m.MessageID == 74 {
			fmt.Println(m)
			fmt.Printf("%02X\n", m)
			gpi, err := mavlink.DecodeMessage(m)
			if err != nil {
				fmt.Println("Error decoding payload: ", err)
			} else {
				// fmt.Printf("Length: %d, Sequence: %d, SysID: %d, CompID: %d, MessID: %d, Payload: %v, CRC: %d\n", m.Length, m.Sequence, m.SystemID, m.ComponentID, m.MessageID, m.Payload, m.CRC)
				fmt.Println(gpi.MessageData())
				// mData := gpi.MessageData()
				// fmt.Println("Decoded VFR_HUD message:")
				// fmt.Printf("Airspeed: %.2f m/s\n", mData["Airspeed"])
				// fmt.Printf("Groundspeed: %.2f m/s\n", mData["Groundspeed"])
				// fmt.Printf("Heading: %.2f degrees\n", mData["Heading"])
				// fmt.Printf("Throttle: %.2f %%\n", mData["Throttle"])
				// fmt.Printf("Altitude: %.2f m\n", mData["Altitude"])
				// fmt.Printf("Climb: %.2f m/s\n", mData["Climb"])
			}
		}
		// if m.MessageID == 33 {
		// 	gpi, err := mavlink.DecodeMessage(m)
		// 	if err != nil {
		// 		fmt.Println("Error decoding payload: ", err)
		// 	} else {
		// 		fmt.Printf("Length: %d, Sequence: %d, SysID: %d, CompID: %d, MessID: %d, Payload: %v, CRC: %d\n", m.Length, m.Sequence, m.SystemID, m.ComponentID, m.MessageID, m.Payload, m.CRC)
		// 		fmt.Println(gpi.MessageData())
		// 	}
		// }

	}
}
