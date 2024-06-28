package sender

import (
	"encoding/binary"
	"fmt"
	"net"
)

func SendMAVLinkMessage(msg []byte, addr string) error {
	// Resolve the UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("error resolving UDP address: %v", err)
	}

	// Create a UDP connection
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return fmt.Errorf("error creating UDP connection: %v", err)
	}
	defer conn.Close()

	// Send the MAVLink message
	_, err = conn.Write(msg)
	if err != nil {
		return fmt.Errorf("error sending MAVLink message: %v", err)
	}

	fmt.Println("MAVLink message sent successfully")

	return nil
}

// encodeSetPositionTargetGlobalInt encodes a MAVLink message to set position target global int
func EncodeSetPositionTargetGlobalInt(lat, lon float64) ([]byte, error) {
	const msgID = 84 // MAVLink message ID for SET_POSITION_TARGET_GLOBAL_INT

	// Initialize the message buffer with the fixed length of the message payload
	msgBuf := make([]byte, 44) // MAVLink SET_POSITION_TARGET_GLOBAL_INT message has a fixed length of 44 bytes

	// Set message ID
	msgBuf[0] = 254                   // Start byte for MAVLink 2 message
	msgBuf[1] = byte(len(msgBuf) - 8) // Payload length
	msgBuf[2] = 1                     // Incompatibility flags
	msgBuf[3] = 1                     // Compatibility flags
	msgBuf[4] = 1                     // Sequence number
	msgBuf[5] = 1                     // System ID (1 for the autopilot system)
	msgBuf[6] = 1                     // Component ID (1 for the autopilot system)
	msgBuf[7] = msgID                 // Message ID

	// Set latitude and longitude
	latInt := int32(lat * 1e7) // Convert latitude to 1e7 scaling
	lonInt := int32(lon * 1e7) // Convert longitude to 1e7 scaling
	binary.LittleEndian.PutUint32(msgBuf[8:12], uint32(latInt))
	binary.LittleEndian.PutUint32(msgBuf[12:16], uint32(lonInt))

	// Set coordinate frame (MAV_FRAME_GLOBAL_RELATIVE_ALT_INT = 6)
	msgBuf[16] = 6

	// Set type mask to indicate that lat and lon are used
	binary.LittleEndian.PutUint16(msgBuf[16:18], 0b000000000011)

	// Calculate and set CRC
	crc := calculateCRC(msgBuf)
	binary.LittleEndian.PutUint16(msgBuf[42:], crc)

	return msgBuf, nil
}

// calculateCRC calculates the CRC for the MAVLink message
func calculateCRC(data []byte) uint16 {
	const poly = 0x1021
	var crc uint16

	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
	}

	return crc
}
