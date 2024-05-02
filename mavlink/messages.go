package mavlink

type MavlinkMessage interface {
	GetData() map[string]interface{}
}

type Message struct {
	SystemID    uint8
	ComponentID uint8
	Sequence    uint8
	Length      uint8
	CRC         uint8
}

type HeartbeatMessage struct {
	Message
	Type           uint8
	Autopilot      uint8
	BaseMode       uint8
	CustomMode     uint32
	SystemStatus   uint8
	MavlinkVersion uint8
}

func (m *HeartbeatMessage) GetData() map[string]interface{} {
	return map[string]interface{}{
		"Type":           m.Type,
		"Autopilot":      m.Autopilot,
		"BaseMode":       m.BaseMode,
		"CustomMode":     m.CustomMode,
		"SystemStatus":   m.SystemStatus,
		"MavlinkVersion": m.MavlinkVersion,
	}
}