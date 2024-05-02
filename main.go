package main

import (
	"flag"
	"fmt"
	"gomavlink/vehicle"
)

var useNetwork bool

func init() {
	flag.BoolVar(&useNetwork, "t", false, "Use network connection instead of serial port")
	flag.Parse()
}

func main() {
	v, err := vehicle.NewVehicle("127.0.0.1:14552", 115200, useNetwork)
	fmt.Printf("Vehicle: %v\n", v)
	if err != nil {
		fmt.Println("Error creating Vehicle: ", err)
		return
	}
	v.Start()
}
