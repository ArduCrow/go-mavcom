package main

import (
	"flag"
	"fmt"
	"io"
	"net"

	"github.com/tarm/serial"
)

type MavlinkReader struct {
	serialPort *serial.Port
	conn       net.Conn
}

var useNetwork bool

func init() {
	flag.BoolVar(&useNetwork, "t", false, "Use network connection instead of serial port")
	flag.Parse()
}

// Create a new MavlinkReader
func NewMavlinkReader(portName string, baud int) (*MavlinkReader, error) {
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

	return &MavlinkReader{serialPort: port, conn: conn}, nil
}

// Read a message from the serial port
func (r *MavlinkReader) ReadMessage() ([]byte, error) {
	var source io.Reader

	if useNetwork {
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
	fmt.Println("Starting MavlinkReader, listening on ")
	for {
		msg, err := r.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message: ", err)
			break
		}
		fmt.Println("Received message: ", msg)
	}
}

func main() {
	r, err := NewMavlinkReader("127.0.0.1:14551", 115200)
	if err != nil {
		fmt.Println("Error creating MavlinkReader: ", err)
		return
	}
	defer r.Close()
	r.Start()
}
