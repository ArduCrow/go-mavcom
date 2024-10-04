package reader

import (
	"bytes"
	"fmt"
	"hash/crc32"
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

	return &MavlinkCommunicator{serialPort: port, Conn: conn, useNetwork: useNetwork, msgChan: make(chan mavlink.DecodedMessage)}, nil
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

func (mc *MavlinkCommunicator) SendMessage(message []byte) error {
	fmt.Println("Sending message")

	if mc.useNetwork {
		_, err := mc.Conn.Write(message)
		fmt.Println("Message sending network")
		if err != nil {
			return fmt.Errorf("error sending network message: %v", err)
		} else {
			fmt.Println("Message sent network")
		}
	} else {
		_, err := mc.serialPort.Write(message)
		fmt.Println("Message sending serial")
		if err != nil {
			return fmt.Errorf("error sending message: %v", err)
		} else {
			fmt.Println("Message sent serial")
		}
	}
	fmt.Println("Message sent:", message)
	return nil
}

func (mc *MavlinkCommunicator) createMessage(mavlinkVersion int, messageID uint32, systemID uint8, componentID uint8, payload []byte) ([]byte, error) {
	msgLength := len(payload)
	msg := bytes.NewBuffer([]byte{})

	// MAVLink 2.0 Header
	if mavlinkVersion == 2 {
		msg.WriteByte(0xFD)             // Start byte for MAVLink 2.0
		msg.WriteByte(uint8(msgLength)) // Payload length
		msg.WriteByte(0x00)             // Incompatibility flags
		msg.WriteByte(0x00)             // Compatibility flags
		msg.WriteByte(mc.seqNumber)     // Sequence number (increment with each message)
		mc.seqNumber++                  // Increment sequence number for next message
		msg.WriteByte(systemID)         // System ID
		msg.WriteByte(componentID)      // Component ID

		// Write the 3-byte message ID
		msg.WriteByte(uint8(messageID & 0xFF))         // Least significant byte
		msg.WriteByte(uint8((messageID >> 8) & 0xFF))  // Middle byte
		msg.WriteByte(uint8((messageID >> 16) & 0xFF)) // Most significant byte
	} else {
		// MAVLink 1.0 Header (if required in the future)
		msg.WriteByte(0xFE)             // Start byte for MAVLink 1.0
		msg.WriteByte(uint8(msgLength)) // Payload length
		msg.WriteByte(mc.seqNumber)     // Sequence number
		mc.seqNumber++                  // Increment sequence number
		msg.WriteByte(systemID)         // System ID
		msg.WriteByte(componentID)      // Component ID
		msg.WriteByte(uint8(messageID)) // Message ID (1 byte in MAVLink 1.0)
	}

	// Write the payload
	msg.Write(payload)

	// Calculate checksum (with extra CRC)
	checksum := mc.calculateChecksum(messageID, payload)
	msg.Write(checksum)

	return msg.Bytes(), nil
}

func (mc *MavlinkCommunicator) calculateChecksum(messageID uint32, payload []byte) []byte {
	crc := crc32.NewIEEE() // Use CRC-16-X25 for MAVLink messages

	// Write payload to the CRC calculator
	crc.Write(payload)

	// Add message ID and extra CRC
	crc.Write([]byte{uint8(messageID & 0xFF)})         // Least significant byte of message ID
	crc.Write([]byte{uint8((messageID >> 8) & 0xFF)})  // Middle byte
	crc.Write([]byte{uint8((messageID >> 16) & 0xFF)}) // Most significant byte
	crc.Write([]byte{EXTRA_CRC_STATUSTEXT})            // The extra CRC byte specific to the STATUSTEXT message

	// Finalize the CRC-16 calculation
	checksum := crc.Sum(nil)
	return checksum[:2] // Only return the first 2 bytes of the checksum
}

// This function should return a channel that will be used to send messages
func (mc *MavlinkCommunicator) Messages() <-chan mavlink.DecodedMessage {
	return mc.msgChan
}
