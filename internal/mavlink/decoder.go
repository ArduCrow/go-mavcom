package mavlink

import (
	"encoding/binary"
	"fmt"
	"math"
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
	case 33:
		return decodeGlobalPositionInt(data)
	case 74:
		return decodeVfrHud(data)
	default:
		return nil, fmt.Errorf("unknown message ID: %d", data.MessageID)
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
			MessageName: "HEARTBEAT",
		},
		Type:         float64(payload[5]),
		Autopilot:    float64(payload[6]),
		BaseMode:     float64(payload[7]),
		SystemStatus: float64(payload[8]),
	}
	return newMessage, nil
}

type HeartbeatMessage struct {
	MavlinkMessage
	Type         float64
	Autopilot    float64
	BaseMode     float64
	SystemStatus float64
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

func decodeGlobalPositionInt(data *RawMessage) (*GlobalPositionIntMessage, error) {
	payload := data.Payload
	if len(payload) != 28 {
		return nil, fmt.Errorf("invalid payload length for GLOBAL_POSITION_INT message")
	}
	newMessage := &GlobalPositionIntMessage{
		MessageID:   data.MessageID,
		MessageName: "GLOBAL_POSITION_INT",
		TimeBootMs:  float64(binary.LittleEndian.Uint32(payload[1:5])),
		Lat:         float64(int32(binary.LittleEndian.Uint32(payload[5:9]))) / 10000000,
		Lon:         float64(int32(binary.LittleEndian.Uint32(payload[9:13]))) / 10000000,
		Alt:         float64(int32(binary.LittleEndian.Uint32(payload[13:17]))) / 1000,
		RelativeAlt: float64(int32(binary.LittleEndian.Uint32(payload[17:21]))) / 1000,
		Vx:          float64(int16(binary.LittleEndian.Uint16(payload[21:23]))) / 100,
		Vy:          float64(int16(binary.LittleEndian.Uint16(payload[23:25]))) / 100,
		Vz:          float64(int16(binary.LittleEndian.Uint16(payload[25:27]))) / 100,
		Hdg:         float64(binary.LittleEndian.Uint16(payload[27:29])) / 100,
	}
	return newMessage, nil
}

type GlobalPositionIntMessage struct {
	MessageID   int
	MessageName string
	TimeBootMs  float64
	Lat         float64
	Lon         float64
	Alt         float64
	RelativeAlt float64
	Vx          float64
	Vy          float64
	Vz          float64
	Hdg         float64
}

func (g *GlobalPositionIntMessage) GetMessageID() int {
	return g.MessageID
}

func (g *GlobalPositionIntMessage) GetMessageName() string {
	return g.MessageName
}

func (g *GlobalPositionIntMessage) MessageData() DecodedPayload {
	return DecodedPayload{
		"TimeBootMs":  g.TimeBootMs,
		"Lat":         g.Lat,
		"Lon":         g.Lon,
		"Alt":         g.Alt,
		"RelativeAlt": g.RelativeAlt,
		"Vx":          g.Vx,
		"Vy":          g.Vy,
		"Vz":          g.Vz,
		"Hdg":         g.Hdg,
	}
}

func decodeVfrHud(data *RawMessage) (*VfrHudMessage, error) {
	payload := data.Payload

	newMessage := &VfrHudMessage{
		MessageID:   data.MessageID,
		MessageName: "VFR_HUD",
		Airspeed:    float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[1:5]))),
		Groundspeed: float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[5:9]))),
		// Throttle:    float64(uint16(binary.LittleEndian.Uint16(payload[9:11]))),
		Alt:      float64(uint16(binary.LittleEndian.Uint32(payload[9:13]))),
		Throttle: float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[11:17]))),
		Climb:    float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[13:17]))),
		Heading:  float64(uint16(binary.LittleEndian.Uint16(payload[17:19]))),
	}
	return newMessage, nil
}

type VfrHudMessage struct {
	MessageID   int
	MessageName string
	Airspeed    float64
	Groundspeed float64
	Heading     float64
	Throttle    float64
	Alt         float64
	Climb       float64
}

func (v *VfrHudMessage) GetMessageID() int {
	return v.MessageID
}

func (v *VfrHudMessage) GetMessageName() string {
	return v.MessageName
}

func (v *VfrHudMessage) MessageData() DecodedPayload {
	return DecodedPayload{
		"Airspeed":    v.Airspeed,
		"Groundspeed": v.Groundspeed,
		"Alt":         v.Alt,
		"Clb":         v.Climb,
		"Heading":     v.Heading,
		"Throttle":    v.Throttle,
	}
}
