package mavlink

const (
	// Message IDs
	MAVLINK_MSG_ID_COMMAND_LONG = 76

	// Message sizes
	MAVLINK_MSG_SIZE_COMMAND_LONG = 33

	// Command IDs
	MAV_CMD_COMPONENT_ARM_DISARM = 400
)

type MavlinkMessage interface {
	MessageID() uint8
	MessageSize() uint8
}

type CommandLong struct {
	Param1          float32
	Param2          float32
	Param3          float32
	Param4          float32
	Param5          float32
	Param6          float32
	Param7          float32
	Command         uint16
	TargetSystem    uint8 // system that should execute the command
	TargetComponent uint8 // component that should execute the command, 0 means all components
	Confirmation    uint8
}

func (msg CommandLong) MessageID() uint8 {
	return MAVLINK_MSG_ID_COMMAND_LONG
}

func (msg CommandLong) MessageSize() uint8 {
	return MAVLINK_MSG_SIZE_COMMAND_LONG
}
