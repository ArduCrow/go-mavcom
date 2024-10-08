package main

import (
	"log"
	"net"

	mav "github.com/Sabmit/go-mavlink"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:5762")
	if err != nil {
		log.Fatalf("Error while connecting to the vehicle: %s\n", err)
	}

	message := new(mav.RequestDataStream)
	message.ReqMessageRate = 20

	err = mav.Send(conn, 1, 200, message)
	if err != nil {
		log.Fatalf("Error while sending the packet: %s\n", err)
	}

	msg2 := new(mav.CommandLong)
	msg2.Command = 400 // MAV_CMD_COMPONENT_ARM_DISARM
	msg2.Param1 = 1    // Arm
	msg2.TargetSystem = 1
	msg2.TargetComponent = 1

	err = mav.Send(conn, 1, 200, msg2)
	if err != nil {
		log.Fatalf("Error while sending the CommandLong packet: %v", err)
	}

	log.Println("Messages sent successfully")
}
