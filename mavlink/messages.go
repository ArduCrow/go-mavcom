package mavlink

import (
	// "encoding/binary"
	"encoding/binary"
	"fmt"
)

type MavlinkMessage interface {
	NewMessage()
}

type RawMessage struct {
	Length      uint8
	Sequence    uint8
	SystemID    uint8
	ComponentID uint8
	MessageID   int
	Payload     []byte
	CRC         uint8
}

func NewRawMessage(data []byte) (*RawMessage, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("insufficient data for MAVLink message")
	}
	payloadLength := int(data[1])
	messageID := binary.LittleEndian.Uint16(data[7:9])

	newMessage := &RawMessage{
		Length:      data[1],
		Sequence:    data[4],
		SystemID:    data[5],
		ComponentID: data[6],
		MessageID:   int(messageID),
		Payload:     data[9 : 9+payloadLength],
		CRC:         data[len(data)-1],
	}
	return newMessage, nil
}

type HeartbeatMessage struct {
	Type      uint8
	Autopilot uint8
	BaseMode  uint8
	// CustomMode     uint32
	SystemStatus uint8
	// MavlinkVersion uint8
}

func NewHeartbeat(data []byte) (*HeartbeatMessage, error) {
	if len(data) != 9 {
		return nil, fmt.Errorf("invalid payload length for Heartbeat message")
	}
	newMessage := &HeartbeatMessage{
		Type:         data[5],
		Autopilot:    data[6],
		BaseMode:     data[7],
		SystemStatus: data[8],
	}
	return newMessage, nil
}