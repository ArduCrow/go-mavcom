package communicator

import (
	"bytes"
	"fmt"
	"gomavlink/internal/mavlink"
	"hash/crc32"
	"net"
)

func (mc *MavlinkCommunicator) SendMessage(message []byte) error {
	fmt.Println("Sending message")

	if mc.useNetwork {
		_, err := net.Dial("tcp", "127.0.0.1:5762")
		if err != nil {
			fmt.Printf("Error while connecting to the vehicle: %s\n", err)
		}
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
	fmt.Println("Sequence number:", mc.seqNumber)
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

// This function should return a channel that will be used to receive messages
// It is read in the main loop of the vehicle package
func (mc *MavlinkCommunicator) Messages() <-chan mavlink.DecodedMessage {
	return mc.msgChan
}
