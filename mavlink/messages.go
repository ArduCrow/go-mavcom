package mavlink

import (
	"encoding/binary"
	"fmt"
)

type RawMessage struct {
	Length      uint8
	Sequence    uint8
	SystemID    uint8
	ComponentID uint8
	MessageID   int
	Payload     []byte
	CRC         uint8
}

type DecodedMessage interface {
	GetMessageID() int
	GetMessageName() string
	MessageData() DecodedPayload
}

type DecodedMavlinkMessage struct {
	MessageID   int
	MessageName string
	Payload     DecodedPayload
}

type DecodedPayload map[interface{}]interface{}

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

func DecodeMessage(r *RawMessage) (DecodedMessage, error) {
	switch r.MessageID {
	case 0:
		return decodeHeartbeat(*r)
	default:
		return nil, fmt.Errorf("unknown message ID: %d", r.MessageID)
	}
}

type HeartbeatMessage struct {
	DecodedMavlinkMessage
	Type         uint8
	Autopilot    uint8
	BaseMode     uint8
	SystemStatus uint8
}

func (h *HeartbeatMessage) GetMessageID() int {
	return h.MessageID
}

func (h *HeartbeatMessage) GetMessageName() string {
	return h.MessageName
}

func (h *HeartbeatMessage) MessageData() DecodedPayload {
	return h.Payload
}

func decodeHeartbeat(data RawMessage) (DecodedMessage, error) {
	if len(data.Payload) != 9 {
		return nil, fmt.Errorf("invalid payload length for Heartbeat message")
	}
	newMessage := &HeartbeatMessage{
		DecodedMavlinkMessage: DecodedMavlinkMessage{
			MessageID:   data.MessageID,
			MessageName: lookup(data.MessageID),
			Payload: DecodedPayload{
				"Type":         data.Payload[5],
				"Autopilot":    data.Payload[6],
				"BaseMode":     data.Payload[7],
				"SystemStatus": data.Payload[8],
			},
		},
	}
	return newMessage, nil
}

// placeholder function until message mapping is done
func lookup(messageID int) string {
	switch messageID {
	case 0:
		return "HEARTBEAT"
	default:
		return "UNKNOWN"
	}
}
