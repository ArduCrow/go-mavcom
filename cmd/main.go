package main

import (
	"flag"
	"fmt"
	"gomavlink/pkg/vehicle"
)

var useNetwork bool

func init() {
	flag.BoolVar(&useNetwork, "t", false, "Use network connection instead of serial port")
	flag.Parse()
}

func main() {
	// v, err := vehicle.NewVehicle("127.0.0.1:14552", 115200, useNetwork)
	v, err := vehicle.NewVehicle("/dev/ttyACM0", 115200, useNetwork)
	fmt.Printf("Vehicle: %v\n", v)
	if err != nil {
		fmt.Println("Error creating Vehicle: ", err)
		return
	}
	v.Start()

	// Since both the reader and the vehicle are running in their own goroutines,
	// we use the select statement to keep the main goroutine running,
	// which it does because select {} is a blocking statement. This is a common
	// pattern in Go for keeping the main goroutine running while other goroutines
	// do work.
	select {}

	// msg, _ := sender.EncodeSetPositionTargetGlobalInt(37.7749, -122.4194)
	// sender.SendMAVLinkMessage(msg, "127.0.0.1:14551")
}
