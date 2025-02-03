package communicator

import (
	"fmt"
	"io"
	"net"

	"github.com/arducrow/go-mavcom/internal/mavlink"

	"github.com/tarm/serial"
)

type MavlinkCommunicator struct {
	serialPort    *serial.Port
	listenPort    string
	Conn          net.Conn
	useNetwork    bool
	msgChan       chan mavlink.DecodedMessage
	CurrentStates CurrentStates
	SeqNumber     uint8
	Encoder       *mavlink.Encoder
	// readWriteLock sync.Mutex
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
		// udpAddr, err := net.ResolveUDPAddr("udp", portName)
		conn, err = net.Dial("tcp", portName)
		if err != nil {
			return nil, err
		}
		// conn, err = net.ListenUDP("udp", udpAddr)
		if err != nil {
			return nil, err
		}
	} else {
		port, err = serial.OpenPort(&serial.Config{Name: portName, Baud: baud})
		if err != nil {
			return nil, err
		}
	}
	NewMavlinkCommunicator := &MavlinkCommunicator{
		serialPort: port,
		Conn:       conn,
		useNetwork: useNetwork,
		msgChan:    make(chan mavlink.DecodedMessage),
		SeqNumber:  uint8(0),
	}

	encoder := mavlink.NewEncoder()
	NewMavlinkCommunicator.Encoder = encoder
	encoder.MavComInterface = NewMavlinkCommunicator

	return NewMavlinkCommunicator, nil
}

func (mc *MavlinkCommunicator) GetSequenceNumber() uint8 {
	return mc.SeqNumber
}

func (mc *MavlinkCommunicator) IncrementSequenceNumber() {
	seqNumber := mc.SeqNumber
	if seqNumber == 0xFF {
		mc.SeqNumber = 0
		fmt.Println("Sequence number reset")
	} else {
		mc.SeqNumber++
	}
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
			// v := mc.Encoder.GetSequenceNumber()
			// fmt.Println("Sequence number in read loop: ", v)
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

func (mc *MavlinkCommunicator) SendArm() {
	msg := mavlink.CommandLong{
		Param1:          1,
		Param2:          0,
		Param3:          0,
		Param4:          0,
		Param5:          0,
		Param6:          0,
		Param7:          0,
		Command:         400,
		TargetSystem:    1,
		TargetComponent: 1,
		Confirmation:    0,
	}
	fmt.Println("Connection: ", mc.Conn.LocalAddr())
	err := mc.Encoder.EncodePacket(mc.Conn, 1, 0, msg)
	if err != nil {
		fmt.Println("Error sending message: ", err)
	}

}

func (mc *MavlinkCommunicator) SendTakeoff(alt float32) {
	msg := mavlink.CommandLong{
		Param1:          0,
		Param2:          0,
		Param3:          0,
		Param4:          0,
		Param5:          0,
		Param6:          0,
		Param7:          alt,
		Command:         mavlink.MAV_CMD_NAV_TAKEOFF,
		TargetSystem:    1,
		TargetComponent: 1,
		Confirmation:    0,
	}
	err := mc.Encoder.EncodePacket(mc.Conn, 1, 0, msg)
	if err != nil {
		fmt.Println("Error sending message: ", err)
	}
}

func (mc *MavlinkCommunicator) SendSetModeGuidedArmed() {
	const MAV_MODE_FLAG_CUSTOM_MODE_ENABLED = 1 << 7
	const MAV_MODE_FLAG_SAFETY_ARMED = 1 << 6
	const GUIDED_MODE = 4

	baseMode := MAV_MODE_FLAG_CUSTOM_MODE_ENABLED | MAV_MODE_FLAG_SAFETY_ARMED

	msg := mavlink.CommandLong{
		Param1:          float32(baseMode),
		Param2:          GUIDED_MODE,
		Param3:          0,
		Param4:          0,
		Param5:          0,
		Param6:          0,
		Param7:          0,
		Command:         176, // MAV_CMD_DO_SET_MODE
		TargetSystem:    1,
		TargetComponent: 0,
		Confirmation:    0,
	}
	fmt.Printf("Sending GUIDED mode command with baseMode: %d, customMode: %d\n", baseMode, GUIDED_MODE)
	err := mc.Encoder.EncodePacket(mc.Conn, 1, 0, msg)
	if err != nil {
		fmt.Println("Error sending message: ", err)
	}
}

func (mc *MavlinkCommunicator) RequestDataStream(streamID uint8, rate uint16) {
	msg := mavlink.CommandLong{
		Param1:          float32(streamID),
		Param2:          float32(rate),
		Param3:          0,
		Param4:          0,
		Param5:          0,
		Param6:          0,
		Param7:          0,
		Command:         mavlink.MAV_CMD_SET_MESSAGE_INTERVAL,
		TargetSystem:    1,
		TargetComponent: 1,
		Confirmation:    0,
	}
	err := mc.Encoder.EncodePacket(mc.Conn, 1, 0, msg)
	if err != nil {
		fmt.Println("Error sending message: ", err)
	}
}

func (mc *MavlinkCommunicator) Messages() <-chan mavlink.DecodedMessage {
	return mc.msgChan
}
