package communicator

import (
	"fmt"
	"io"
	"net"

	"gomavlink/internal/mavlink"

	"github.com/tarm/serial"
)

type MavlinkCommunicator struct {
	serialPort    *serial.Port
	listenPort    string
	Conn          net.Conn
	useNetwork    bool
	msgChan       chan mavlink.DecodedMessage
	CurrentStates CurrentStates
	seqNumber     uint8
}

type CurrentStates struct {
	GlobalPositionIntState mavlink.DecodedPayload
	VFRHUDState            mavlink.DecodedPayload
	Heartbeat              mavlink.DecodedPayload
}

func NewMavlinkCommunicator(portName string, baud int, useNetwork bool) (*MavlinkCommunicator, error) {
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

	return &MavlinkCommunicator{serialPort: port, Conn: conn, useNetwork: useNetwork, msgChan: make(chan mavlink.DecodedMessage), seqNumber: 1}, nil
}

func (mc *MavlinkCommunicator) readMessage() ([]byte, error) {
	var source io.Reader
	const minMsgLength = 8

	if mc.useNetwork {
		source = mc.Conn
	} else {
		source = mc.serialPort
	}

	buf := make([]byte, 1024)
	n, err := source.Read(buf)
	if n < minMsgLength {
		return nil, fmt.Errorf("message too short")
	}
	if err != nil {
		return nil, err
	}

	// TODO - Check if the message is a valid mavlink message and parse it

	return buf[:n], nil
}

// Close the connection
func (mc *MavlinkCommunicator) Close() error {
	if mc.serialPort != nil {
		return mc.serialPort.Close()
	}
	return nil
}

// Begin reading messages. Spawns its own goroutine so that the
// main loop can continue to run (the Read method in the readMessage
// function is blocking)
func (mc *MavlinkCommunicator) Start() {
	if mc.useNetwork {
		mc.listenPort = mc.Conn.LocalAddr().String()
	} else {
		mc.listenPort = fmt.Sprintf("%v", mc.serialPort)
	}

	// streamIDs := []uint8{0, 1, 2, 3, 4, 6, 10} // Add other stream IDs as needed
	// for _, streamID := range streamIDs {
	// 	err := mc.requestDataStream(streamID, 5)
	// 	if err != nil {
	// 		fmt.Println("Error requesting data stream: ", err)
	// 	}
	// }
	fmt.Printf("Starting MavlinkCommunicator, listening on %v\n", mc.listenPort)
	go func() {
		for {
			msg, err := mc.readMessage()
			if err != nil {
				fmt.Println("Error reading message: ", err)
				continue
			}
			m, err := mavlink.NewRawMessage(msg)
			if err != nil {
				fmt.Println("Error parsing message:", err)
				continue
			}

			// fmt.Println("Message ID: ", m.MessageID)

			if m.MessageID == 0 {
				fmt.Println("Heartbeat received")
			}
			// if the message ID is not 0, 33 or 74 then ignore it
			if m.MessageID != 0 && m.MessageID != 33 && m.MessageID != 74 {
				continue
			}
			decodedMessage, err := mavlink.DecodeMessage(m)
			if err != nil {
				fmt.Println("Error decoding message: ", err)
				continue
			}
			// fmt.Println(decodedMessage.GetMessageName())

			mc.msgChan <- decodedMessage

		}
	}()

}
