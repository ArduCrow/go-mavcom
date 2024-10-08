package communicator

import (
	"encoding/binary"
	"fmt"
)

const (
	MAVLINK_MSG_ID_REQUEST_DATA_STREAM = 66
	MAV_CMD_COMPONENT_ARM_DISARM       = 400
	MAVLINK_MSG_ID_COMMAND_LONG        = 76
	MAVLINK_MSG_ID_STATUSTEXT          = 253
	EXTRA_CRC_STATUSTEXT               = 83
	MAVLINK_SYSTEM_ID                  = 255
	MAVLINK_COMPONENT_ID               = 1
	TARGET_SYSTEM                      = 1
	TARGET_COMPONENT                   = 1
	START_STREAM                       = 1
)

func (mc *MavlinkCommunicator) RequestDataStream(streamID, rateHz uint8) error {
	payload := make([]byte, 6)
	payload[0] = TARGET_SYSTEM
	payload[1] = TARGET_COMPONENT
	payload[2] = streamID
	binary.LittleEndian.PutUint16(payload[3:], uint16(rateHz))
	payload[5] = START_STREAM

	msg, err := mc.createMessage(1, MAVLINK_MSG_ID_REQUEST_DATA_STREAM, MAVLINK_SYSTEM_ID, MAVLINK_COMPONENT_ID, payload)
	if err != nil {
		return fmt.Errorf("error creating request data stream message: %v", err)
	}

	err = mc.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("error sending request data stream message: %v", err)
	}

	fmt.Printf("Requested data stream (ID: %d) at %d Hz\n", streamID, rateHz)
	return nil
}

func (mc *MavlinkCommunicator) ArmMotors() error {
	fmt.Println("Arming motors")

	payload := make([]byte, 33)

	// Command ID: MAV_CMD_COMPONENT_ARM_DISARM (400)
	binary.LittleEndian.PutUint16(payload[0:2], MAV_CMD_COMPONENT_ARM_DISARM)

	// Target system and component (drone's system ID and component ID)
	payload[2] = 1 // Target system (drone)
	payload[3] = 1 // Target component (drone)

	// Confirmation (0)
	payload[4] = 0

	// Param1: 1 to arm
	binary.LittleEndian.PutUint32(payload[5:], 1)

	// Param2: 0 (force arm is optional)
	binary.LittleEndian.PutUint32(payload[9:], 0)

	// Param3 to Param7: all 0s (not used)

	// Send the message
	msg, err := mc.createMessage(2, MAVLINK_MSG_ID_COMMAND_LONG, 255, 1, payload)
	if err != nil {
		return fmt.Errorf("error creating arm motors message: %v", err)
	}

	fmt.Println("Sending Arm Motors message:", msg)

	err = mc.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("error sending arm motors message: %v", err)
	}

	return nil
}
func (mc *MavlinkCommunicator) SendStatusText(severity uint8, text string) error {
	fmt.Println("Sending status text")
	payload := make([]byte, 51)
	payload[0] = severity
	copy(payload[1:], text) // Copy the text into the payload

	// Ensure the text is null-terminated (for MAVLink)
	if len(text) < 50 {
		payload[len(text)+1] = 0
	}

	// Create the MAVLink 2.0 message
	msg, err := mc.createMessage(2, MAVLINK_MSG_ID_STATUSTEXT, MAVLINK_SYSTEM_ID, MAVLINK_COMPONENT_ID, payload)
	if err != nil {
		return fmt.Errorf("error creating status text message: %v", err)
	}

	// Send the message
	err = mc.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("error sending status text message: %v", err)
	}

	return nil
}
