package reader

import (
	"fmt"
	"io"
	"net"

	"gomavlink/internal/mavlink"

	"github.com/tarm/serial"
)

type MavlinkReader struct {
	serialPort    *serial.Port
	listenPort    string
	Conn          net.Conn
	useNetwork    bool
	msgChan       chan mavlink.DecodedMessage
	CurrentStates CurrentStates
}

type CurrentStates struct {
	GlobalPositionIntState mavlink.DecodedPayload
	VFRHUDState            mavlink.DecodedPayload
	Heartbeat              mavlink.DecodedPayload
}

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

	return &MavlinkReader{serialPort: port, Conn: conn, useNetwork: useNetwork, msgChan: make(chan mavlink.DecodedMessage)}, nil
}

func (r *MavlinkReader) readMessage() ([]byte, error) {
	var source io.Reader
	const minMsgLength = 8

	if r.useNetwork {
		source = r.Conn
	} else {
		source = r.serialPort
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
func (r *MavlinkReader) Close() error {
	if r.serialPort != nil {
		return r.serialPort.Close()
	}
	return nil
}

// Begin reading messages. Spawns its own goroutine so that the
// main loop can continue to run (the Read method in the readMessage
// function is blocking)
func (r *MavlinkReader) Start() {
	if r.useNetwork {
		r.listenPort = r.Conn.LocalAddr().String()
	} else {
		r.listenPort = fmt.Sprintf("%v", r.serialPort)
	}
	fmt.Printf("Starting MavlinkReader, listening on %v\n", r.listenPort)
	go func() {
		for {
			msg, err := r.readMessage()
			if err != nil {
				fmt.Println("Error reading message: ", err)
				continue
			}
			m, err := mavlink.NewRawMessage(msg)
			if err != nil {
				fmt.Println("Error parsing message:", err)
				continue
			}

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

			r.msgChan <- decodedMessage

		}
	}()

}

// This function should return a channel that will be used to send messages
func (r *MavlinkReader) Messages() <-chan mavlink.DecodedMessage {
	return r.msgChan
}
