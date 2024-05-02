package main

import (
	"flag"
	"fmt"
	"gomavlink/reader"
)

var useNetwork bool

func init() {
	flag.BoolVar(&useNetwork, "t", false, "Use network connection instead of serial port")
	flag.Parse()
}

func main() {
	r, err := reader.NewMavlinkReader("127.0.0.1:14551", 115200, useNetwork)
	if err != nil {
		fmt.Println("Error creating MavlinkReader: ", err)
		return
	}
	defer r.Close()
	r.Start()
}
