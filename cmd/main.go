package main

import (
	"flag"
	"fmt"
	"gomavlink/pkg/vehicle"
	"time"
)

var useNetwork bool

func init() {
	flag.BoolVar(&useNetwork, "t", false, "Use network connection instead of serial port")
	flag.Parse()
}

func main() {
	v, err := vehicle.NewVehicle("127.0.0.1:14550", 115200, useNetwork)
	// v, err := vehicle.NewVehicle("/dev/ttyACM0", 115200, useNetwork)
	fmt.Printf("Vehicle: %v\n", v)
	if err != nil {
		fmt.Println("Error creating Vehicle: ", err)
		return
	}
	v.Start()

	// sleep 4 seconds
	time.Sleep(1 * time.Second)
	err = v.Connection.ArmMotors()
	if err != nil {
		fmt.Println("Error arming motors: ", err)
	}
	time.Sleep(6 * time.Second)
	err = v.Connection.ArmMotors()
	if err != nil {
		fmt.Println("Error arming motors: ", err)
	}
	time.Sleep(6 * time.Second)
	err = v.Connection.ArmMotors()
	if err != nil {
		fmt.Println("Error arming motors: ", err)
	}
	time.Sleep(6 * time.Second)
	err = v.Connection.ArmMotors()
	if err != nil {
		fmt.Println("Error arming motors: ", err)
	}
	// v.Connection.SendCommandLong()
	// for i := 0; i < 100; i++ {
	// 	v.Connection.SendStatusText(6, "AUTOPILOT CONNECTED")
	// 	time.Sleep(1 * time.Second)
	// }

	// Since both the reader and the vehicle are running in their own goroutines,
	// we use the select statement to keep the main goroutine running,
	// which it does because select {} is a blocking statement. This is a common
	// pattern in Go for keeping the main goroutine running while other goroutines
	// do work.
	select {}

	// msg, _ := sender.EncodeSetPositionTargetGlobalInt(37.7749, -122.4194)
	// sender.SendMAVLinkMessage(msg, "127.0.0.1:14551")
}
