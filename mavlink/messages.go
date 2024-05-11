package mavlink

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
)

type DecodedMessage interface {
	GetMessageID() int
	GetMessageName() string
	MessageData() DecodedPayload
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

type MavlinkMessage struct {
	MessageID   int
	MessageName string
	Payload     DecodedPayload
}

type DecodedPayload map[string]interface{}

var MessagePayloadMap map[string][]string

func init() {
	messageData, err := os.ReadFile("./messages.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(messageData, &MessagePayloadMap); err != nil {
		panic(err)
	}
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

func DecodeMessage(data *RawMessage) (DecodedMessage, error) {
	switch data.MessageID {
	case 0:
		return decodeHeartbeat(data)
	// case 33:
	// 	return decodeGlobalPositionInt(data)
	default:
		return nil, fmt.Errorf("unknown message ID: %d", data.MessageID)
	}
}

// Placeholder function until message mapping is done
func lookup(messageID int) string {
	switch messageID {
	case 0:
		return "HEARTBEAT"
	case 33:
		return "GLOBAL_POSITION_INT"
	default:
		return "UNKNOWN"
	}
}

func decodeHeartbeat(data *RawMessage) (*HeartbeatMessage, error) {
	payload := data.Payload
	if len(payload) != 9 {
		return nil, fmt.Errorf("invalid payload length for HEARTBEAT message")
	}
	newMessage := &HeartbeatMessage{
		MavlinkMessage: MavlinkMessage{
			MessageID:   data.MessageID,
			MessageName: lookup(data.MessageID),
		},
		Type:         data.Payload[5],
		Autopilot:    data.Payload[6],
		BaseMode:     data.Payload[7],
		SystemStatus: data.Payload[8],
	}
	return newMessage, nil
}

type HeartbeatMessage struct {
	MavlinkMessage
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

// return all the fields of the message as a DecodedPayload
func (h *HeartbeatMessage) MessageData() DecodedPayload {
	return DecodedPayload{
		"Type":         h.Type,
		"Autopilot":    h.Autopilot,
		"BaseMode":     h.BaseMode,
		"SystemStatus": h.SystemStatus,
	}
}
