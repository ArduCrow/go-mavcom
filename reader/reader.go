package reader

import (
	"fmt"
	"io"
	"net"

	"github.com/tarm/serial"
)

type MavlinkReader struct {
	serialPort *serial.Port
	conn       net.Conn
	useNetwork bool
}

// Create a new MavlinkReader
func NewMavlinkReader(portName string, baud int, useNetwork bool) (*MavlinkReader, error) {
	var port *serial.Port
	var conn net.Conn
	var err error

	if useNetwork {
		udpAddr, err := net.ResolveUDPAddr("udp", portName)
		if err != nil {
			return nil, err
		}
		conn, err = net.ListenUDP("udp", udpAddr)
		if err != nil {
			return nil, err
		}
	} else {
		port, err = serial.OpenPort(&serial.Config{Name: portName, Baud: baud})
		if err != nil {
			return nil, err
		}
	}

	return &MavlinkReader{serialPort: port, conn: conn, useNetwork: useNetwork}, nil
}

// Read a message from the serial port
func (r *MavlinkReader) readMessage() ([]byte, error) {
	var source io.Reader

	if r.useNetwork {
		source = r.conn
	} else {
		source = r.serialPort
	}

	buf := make([]byte, 1024)
	n, err := source.Read(buf)
	if err != nil {
		return nil, err
	}

	// TODO - Check if the message is a valid mavlink message and parse it

	return buf[:n], nil
}

// Close the connection
func (r *MavlinkReader) Close() error {
	if r.serialPort != nil {
		return r.serialPort.Close()
	}
	return nil
}

// Begin reading messages
func (r *MavlinkReader) Start() {
	var listenPort string
	if r.useNetwork {
		listenPort = r.conn.LocalAddr().String()
	} else {
		listenPort = fmt.Sprintf("%v", r.serialPort)
	}
	fmt.Printf("Starting MavlinkReader, listening on %v\n", listenPort)
	for {
		msg, err := r.readMessage()
		if err != nil {
			fmt.Println("Error reading message: ", err)
			break
		}
		fmt.Println("Received message: ", msg)
	}
}
