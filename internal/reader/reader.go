package reader

import (
	"bytes"
	"encoding/binary"
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

func createMessage(messageID uint8, systemID uint8, componentID uint8, payload []byte) ([]byte, error) {
	msgLength := len(payload)

	// Create a buffer to store the message
	msg := bytes.NewBuffer([]byte{})

	// Write the MAVLink header (assuming MAVLink 1.0 for simplicity)
	msg.WriteByte(0xFE)             // Start byte
	msg.WriteByte(uint8(msgLength)) // Payload length
	msg.WriteByte(0)                // Packet sequence (dummy value for now)
	msg.WriteByte(systemID)         // System ID
	msg.WriteByte(componentID)      // Component ID
	msg.WriteByte(messageID)        // Message ID

	// Write the payload
	msg.Write(payload)

	// Calculate and append the checksum (simple XOR for demonstration)
	checksum := uint8(0)
	for _, b := range msg.Bytes() {
		checksum ^= b
	}
	msg.WriteByte(checksum)

	return msg.Bytes(), nil
}

func (r *MavlinkReader) requestDataStream(streamID, rateHz uint8) error {
	const (
		MAVLINK_MSG_ID_REQUEST_DATA_STREAM = 66
		MAVLINK_SYSTEM_ID                  = 1
		MAVLINK_COMPONENT_ID               = 1
		TARGET_SYSTEM                      = 1
		TARGET_COMPONENT                   = 1
		START_STREAM                       = 1
	)

	payload := make([]byte, 6)
	payload[0] = TARGET_SYSTEM
	payload[1] = TARGET_COMPONENT
	payload[2] = streamID
	binary.LittleEndian.PutUint16(payload[3:], uint16(rateHz))
	payload[5] = START_STREAM

	msg, err := createMessage(MAVLINK_MSG_ID_REQUEST_DATA_STREAM, MAVLINK_SYSTEM_ID, MAVLINK_COMPONENT_ID, payload)
	if err != nil {
		return fmt.Errorf("error creating request data stream message: %v", err)
	}

	if r.useNetwork {
		_, err = r.Conn.Write(msg)
	} else {
		_, err = r.serialPort.Write(msg)
	}

	if err != nil {
		return fmt.Errorf("error sending request data stream message: %v", err)
	}

	fmt.Printf("Requested data stream (ID: %d) at %d Hz\n", streamID, rateHz)
	return nil
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

	streamIDs := []uint8{0, 1, 2, 3, 4, 6, 10} // Add other stream IDs as needed
	for _, streamID := range streamIDs {
		err := r.requestDataStream(streamID, 5)
		if err != nil {
			fmt.Println("Error requesting data stream: ", err)
		}
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

			fmt.Println("Message ID: ", m.MessageID)

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
