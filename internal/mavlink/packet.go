package mavlink

import (
	"bytes"
	"encoding/binary"
)

const (
	MAX_PAYLOAD_LEN   = 255
	MAX_PACKET_LEN    = 263 // 6 + 255 + 2 (header + payload + footer)
	X25_INIT_CRC      = uint16(0xffff)
	X25_VALIDATE_CRC  = uint16(0xf0b8)
	CRC_EXTRA_ENABLED = true
	FRAME_START       = 0xFE
)

// Each Message's ID will be the index in this slice, the value of which is that message's CRC
var (
	messageCrcs = []byte{50, 124, 137, 0, 237, 217, 104, 119, 0, 0, 0, 89, 0, 0, 0, 0, 0, 0, 0, 0, 214, 159, 220, 168, 24, 23, 170, 144, 67, 115, 39, 246, 185, 104, 237, 244, 222, 212, 9, 254, 230, 28, 28, 132, 221, 232, 11, 153, 41, 39, 214, 223, 141, 33, 15, 3, 100, 24, 239, 238, 30, 240, 183, 130, 130, 0, 148, 21, 0, 243, 124, 0, 0, 0, 20, 0, 152, 143, 0, 0, 127, 106, 0, 0, 0, 0, 0, 0, 0, 231, 183, 63, 54, 0, 0, 0, 0, 0, 0, 0, 175, 102, 158, 208, 56, 93, 0, 0, 0, 0, 235, 93, 124, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42, 241, 15, 134, 219, 208, 188, 84, 22, 19, 21, 134, 0, 78, 68, 189, 127, 111, 21, 21, 144, 1, 234, 73, 181, 22, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 204, 49, 170, 44, 83, 46, 0}
)

// RawMavlinkPacket is a struct that contains a MavlinkPacket and a buffer that contains the raw bytes of the packet
type RawMavlinkPacket struct {
	Packet    *MavlinkPacket
	RawBuffer bytes.Buffer
}

type MavlinkHeader struct {
	FrameStart     uint8
	PayloadLen     uint8
	PacketSequence uint8
	SystemID       uint8
	ComponentID    uint8
	MessageID      uint8
}

type MavlinkPacket struct {
	Header   MavlinkHeader
	Message  MavlinkMessage
	Checksum uint16
}

func (h *MavlinkHeader) HeaderSize() uint8 {
	return 6
}

func (rmp *RawMavlinkPacket) computeChecksum() uint16 {
	rmp.Packet.crcInit()
	/* 1 represents indexOf header.PayloadLength */
	for _, v := range rmp.RawBuffer.Bytes()[1 : rmp.Packet.Header.HeaderSize()+rmp.Packet.Message.MessageSize()] {
		rmp.Packet.crcAccumulate(v)
	}

	if CRC_EXTRA_ENABLED {
		rmp.Packet.crcAccumulate(messageCrcs[rmp.Packet.Message.MessageID()])
	}
	return rmp.Packet.Checksum
}

// Start off the checksum with the X25_INIT_CRC value
func (mp *MavlinkPacket) crcInit() {
	mp.Checksum = X25_INIT_CRC
}

// For each byte of the message, accumulate it into the checksum
func (mp *MavlinkPacket) crcAccumulate(data uint8) {
	// Accumulate the data into the checksum
	var tmp uint8

	tmp = data ^ uint8(mp.Checksum&0xff) // XOR's this byte (data) with the lower byte of the checksum (XOR compares each bit of each byte and makes a new byte with each bit being 1 if the 0 and 1 of the two bytes are not the same, and 0 if they are the same)
	// XOR introduces complexity and randomness into the checksum and means that a small change in data is a big change in the checksum, so we can detect errors more easily
	tmp ^= tmp << 4 // Shifts the bits of tmp to the left by 4 (this is part of the CRC calculation, adds more complexity to the checksum to help detect errors)

	mp.Checksum = (mp.Checksum >> 8) ^ (uint16(tmp) << 8) ^ (uint16(tmp) << 3) ^ (uint16(tmp) >> 4) // Shifts the checksum to the right by 8, then XOR's it with tmp shifted to the left by 8, then XOR's it with tmp shifted to the left by 3, then XOR's it with tmp shifted to the right by 4
	// this last bit is basically just to add more complexity to the checksum, to make it more likely that a small change in the data will result in a big change in the checksum
}

func (mp *MavlinkPacket) computeChecksum() uint16 {
	mp.crcInit()

	// loop over the header and payload bytes and update the checksum
	for _, packetByte := range mp.Bytes()[1 : mp.Header.HeaderSize()+mp.Message.MessageSize()] {
		mp.crcAccumulate(packetByte)
	}

	if CRC_EXTRA_ENABLED {
		// Add the message ID to the checksum
		mp.crcAccumulate(messageCrcs[mp.Header.MessageID])
	}
	return mp.Checksum
}

// Creates a byte buffer and writes the header, message, and checksum to it
func (mp *MavlinkPacket) Bytes() []byte {
	var r bytes.Buffer
	var err error

	if err = binary.Write(&r, binary.LittleEndian, mp.Header); err != nil {
		return []byte{}
	}

	if err = binary.Write(&r, binary.LittleEndian, mp.Message); err != nil {
		return []byte{}
	}
	if err = binary.Write(&r, binary.LittleEndian, mp.Checksum); err != nil {
		return []byte{}
	}
	return r.Bytes()
}
