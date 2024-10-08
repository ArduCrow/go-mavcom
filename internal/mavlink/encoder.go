package mavlink

import (
	"io"
)

type MavlinkCommunicatorInterface interface {
	GetSequenceNumber() uint8
	IncrementSequenceNumber()
}

type Encoder struct {
	MavComInterface MavlinkCommunicatorInterface
}

func NewEncoder() *Encoder {
	return &Encoder{MavComInterface: nil}
}

func (e *Encoder) GetSequenceNumber() uint8 {
	return e.MavComInterface.GetSequenceNumber()
}

func (e *Encoder) IncrementSequenceNumber() {
	e.MavComInterface.IncrementSequenceNumber()
}

func (e *Encoder) CreatePacket(systemID uint8, componentID uint8, message MavlinkMessage) (*MavlinkPacket, error) {
	packet := &MavlinkPacket{
		Header: MavlinkHeader{
			FrameStart:     FRAME_START,
			PayloadLen:     message.MessageSize(),
			PacketSequence: e.GetSequenceNumber(),
			SystemID:       systemID,
			ComponentID:    componentID,
			MessageID:      message.MessageID(),
		},
		Message: message,
	}

	packet.Checksum = packet.computeChecksum()
	return packet, nil
}

func (e *Encoder) EncodePacket(writer io.Writer, systemID uint8, componentID uint8, message MavlinkMessage) error {
	packet, err := e.CreatePacket(systemID, componentID, message)
	if err != nil {
		return err
	}
	packetBytes := packet.Bytes()
	_, err = writer.Write(packetBytes)
	return err

}
